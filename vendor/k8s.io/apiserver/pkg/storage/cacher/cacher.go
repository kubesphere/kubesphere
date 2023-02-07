/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cacher

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/audit"
	"k8s.io/apiserver/pkg/features"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/cacher/metrics"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	utilflowcontrol "k8s.io/apiserver/pkg/util/flowcontrol"
	"k8s.io/client-go/tools/cache"
	"k8s.io/component-base/tracing"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

var (
	emptyFunc = func(bool) {}
)

const (
	// storageWatchListPageSize is the cacher's request chunk size of
	// initial and resync watch lists to storage.
	storageWatchListPageSize = int64(10000)
	// defaultBookmarkFrequency defines how frequently watch bookmarks should be send
	// in addition to sending a bookmark right before watch deadline.
	//
	// NOTE: Update `eventFreshDuration` when changing this value.
	defaultBookmarkFrequency = time.Minute
)

// Config contains the configuration for a given Cache.
type Config struct {
	// An underlying storage.Interface.
	Storage storage.Interface

	// An underlying storage.Versioner.
	Versioner storage.Versioner

	// The GroupResource the cacher is caching. Used for disambiguating *unstructured.Unstructured (CRDs) in logging
	// and metrics.
	GroupResource schema.GroupResource

	// The Cache will be caching objects of a given Type and assumes that they
	// are all stored under ResourcePrefix directory in the underlying database.
	ResourcePrefix string

	// KeyFunc is used to get a key in the underlying storage for a given object.
	KeyFunc func(runtime.Object) (string, error)

	// GetAttrsFunc is used to get object labels, fields
	GetAttrsFunc func(runtime.Object) (label labels.Set, field fields.Set, err error)

	// IndexerFuncs is used for optimizing amount of watchers that
	// needs to process an incoming event.
	IndexerFuncs storage.IndexerFuncs

	// Indexers is used to accelerate the list operation, falls back to regular list
	// operation if no indexer found.
	Indexers *cache.Indexers

	// NewFunc is a function that creates new empty object storing a object of type Type.
	NewFunc func() runtime.Object

	// NewList is a function that creates new empty object storing a list of
	// objects of type Type.
	NewListFunc func() runtime.Object

	Codec runtime.Codec

	Clock clock.Clock
}

type watchersMap map[int]*cacheWatcher

func (wm watchersMap) addWatcher(w *cacheWatcher, number int) {
	wm[number] = w
}

func (wm watchersMap) deleteWatcher(number int, done func(*cacheWatcher)) {
	if watcher, ok := wm[number]; ok {
		delete(wm, number)
		done(watcher)
	}
}

func (wm watchersMap) terminateAll(done func(*cacheWatcher)) {
	for key, watcher := range wm {
		delete(wm, key)
		done(watcher)
	}
}

type indexedWatchers struct {
	allWatchers   watchersMap
	valueWatchers map[string]watchersMap
}

func (i *indexedWatchers) addWatcher(w *cacheWatcher, number int, value string, supported bool) {
	if supported {
		if _, ok := i.valueWatchers[value]; !ok {
			i.valueWatchers[value] = watchersMap{}
		}
		i.valueWatchers[value].addWatcher(w, number)
	} else {
		i.allWatchers.addWatcher(w, number)
	}
}

func (i *indexedWatchers) deleteWatcher(number int, value string, supported bool, done func(*cacheWatcher)) {
	if supported {
		i.valueWatchers[value].deleteWatcher(number, done)
		if len(i.valueWatchers[value]) == 0 {
			delete(i.valueWatchers, value)
		}
	} else {
		i.allWatchers.deleteWatcher(number, done)
	}
}

func (i *indexedWatchers) terminateAll(groupResource schema.GroupResource, done func(*cacheWatcher)) {
	// note that we don't have to call setDrainInputBufferLocked method on the watchers
	// because we take advantage of the default value - stop immediately
	// also watchers that have had already its draining strategy set
	// are no longer available (they were removed from the allWatchers and the valueWatchers maps)
	if len(i.allWatchers) > 0 || len(i.valueWatchers) > 0 {
		klog.Warningf("Terminating all watchers from cacher %v", groupResource)
	}
	i.allWatchers.terminateAll(done)
	for _, watchers := range i.valueWatchers {
		watchers.terminateAll(done)
	}
	i.valueWatchers = map[string]watchersMap{}
}

// As we don't need a high precision here, we keep all watchers timeout within a
// second in a bucket, and pop up them once at the timeout. To be more specific,
// if you set fire time at X, you can get the bookmark within (X-1,X+1) period.
type watcherBookmarkTimeBuckets struct {
	lock sync.Mutex
	// the key of watcherBuckets is the number of seconds since createTime
	watchersBuckets   map[int64][]*cacheWatcher
	createTime        time.Time
	startBucketID     int64
	clock             clock.Clock
	bookmarkFrequency time.Duration
}

func newTimeBucketWatchers(clock clock.Clock, bookmarkFrequency time.Duration) *watcherBookmarkTimeBuckets {
	return &watcherBookmarkTimeBuckets{
		watchersBuckets:   make(map[int64][]*cacheWatcher),
		createTime:        clock.Now(),
		startBucketID:     0,
		clock:             clock,
		bookmarkFrequency: bookmarkFrequency,
	}
}

// adds a watcher to the bucket, if the deadline is before the start, it will be
// added to the first one.
func (t *watcherBookmarkTimeBuckets) addWatcher(w *cacheWatcher) bool {
	// note that the returned time can be before t.createTime,
	// especially in cases when the nextBookmarkTime method
	// give us the zero value of type Time
	// so buckedID can hold a negative value
	nextTime, ok := w.nextBookmarkTime(t.clock.Now(), t.bookmarkFrequency)
	if !ok {
		return false
	}
	bucketID := int64(nextTime.Sub(t.createTime) / time.Second)
	t.lock.Lock()
	defer t.lock.Unlock()
	if bucketID < t.startBucketID {
		bucketID = t.startBucketID
	}
	watchers := t.watchersBuckets[bucketID]
	t.watchersBuckets[bucketID] = append(watchers, w)
	return true
}

func (t *watcherBookmarkTimeBuckets) popExpiredWatchers() [][]*cacheWatcher {
	currentBucketID := int64(t.clock.Since(t.createTime) / time.Second)
	// There should be one or two elements in almost all cases
	expiredWatchers := make([][]*cacheWatcher, 0, 2)
	t.lock.Lock()
	defer t.lock.Unlock()
	for ; t.startBucketID <= currentBucketID; t.startBucketID++ {
		if watchers, ok := t.watchersBuckets[t.startBucketID]; ok {
			delete(t.watchersBuckets, t.startBucketID)
			expiredWatchers = append(expiredWatchers, watchers)
		}
	}
	return expiredWatchers
}

type filterWithAttrsFunc func(key string, l labels.Set, f fields.Set) bool

type indexedTriggerFunc struct {
	indexName   string
	indexerFunc storage.IndexerFunc
}

// Cacher is responsible for serving WATCH and LIST requests for a given
// resource from its internal cache and updating its cache in the background
// based on the underlying storage contents.
// Cacher implements storage.Interface (although most of the calls are just
// delegated to the underlying storage).
type Cacher struct {
	// HighWaterMarks for performance debugging.
	// Important: Since HighWaterMark is using sync/atomic, it has to be at the top of the struct due to a bug on 32-bit platforms
	// See: https://golang.org/pkg/sync/atomic/ for more information
	incomingHWM storage.HighWaterMark
	// Incoming events that should be dispatched to watchers.
	incoming chan watchCacheEvent

	resourcePrefix string

	sync.RWMutex

	// Before accessing the cacher's cache, wait for the ready to be ok.
	// This is necessary to prevent users from accessing structures that are
	// uninitialized or are being repopulated right now.
	// ready needs to be set to false when the cacher is paused or stopped.
	// ready needs to be set to true when the cacher is ready to use after
	// initialization.
	ready *ready

	// Underlying storage.Interface.
	storage storage.Interface

	// Expected type of objects in the underlying cache.
	objectType reflect.Type
	// Used for logging, to disambiguate *unstructured.Unstructured (CRDs)
	groupResource schema.GroupResource

	// "sliding window" of recent changes of objects and the current state.
	watchCache *watchCache
	reflector  *cache.Reflector

	// Versioner is used to handle resource versions.
	versioner storage.Versioner

	// newFunc is a function that creates new empty object storing a object of type Type.
	newFunc func() runtime.Object

	// indexedTrigger is used for optimizing amount of watchers that needs to process
	// an incoming event.
	indexedTrigger *indexedTriggerFunc
	// watchers is mapping from the value of trigger function that a
	// watcher is interested into the watchers
	watcherIdx int
	watchers   indexedWatchers

	// Defines a time budget that can be spend on waiting for not-ready watchers
	// while dispatching event before shutting them down.
	dispatchTimeoutBudget timeBudget

	// Handling graceful termination.
	stopLock sync.RWMutex
	stopped  bool
	stopCh   chan struct{}
	stopWg   sync.WaitGroup

	clock clock.Clock
	// timer is used to avoid unnecessary allocations in underlying watchers.
	timer *time.Timer

	// dispatching determines whether there is currently dispatching of
	// any event in flight.
	dispatching bool
	// watchersBuffer is a list of watchers potentially interested in currently
	// dispatched event.
	watchersBuffer []*cacheWatcher
	// blockedWatchers is a list of watchers whose buffer is currently full.
	blockedWatchers []*cacheWatcher
	// watchersToStop is a list of watchers that were supposed to be stopped
	// during current dispatching, but stopping was deferred to the end of
	// dispatching that event to avoid race with closing channels in watchers.
	watchersToStop []*cacheWatcher
	// Maintain a timeout queue to send the bookmark event before the watcher times out.
	bookmarkWatchers *watcherBookmarkTimeBuckets
	// expiredBookmarkWatchers is a list of watchers that were expired and need to be schedule for a next bookmark event
	expiredBookmarkWatchers []*cacheWatcher
}

// NewCacherFromConfig creates a new Cacher responsible for servicing WATCH and LIST requests from
// its internal cache and updating its cache in the background based on the
// given configuration.
func NewCacherFromConfig(config Config) (*Cacher, error) {
	stopCh := make(chan struct{})
	obj := config.NewFunc()
	// Give this error when it is constructed rather than when you get the
	// first watch item, because it's much easier to track down that way.
	if err := runtime.CheckCodec(config.Codec, obj); err != nil {
		return nil, fmt.Errorf("storage codec doesn't seem to match given type: %v", err)
	}

	var indexedTrigger *indexedTriggerFunc
	if config.IndexerFuncs != nil {
		// For now, we don't support multiple trigger functions defined
		// for a given resource.
		if len(config.IndexerFuncs) > 1 {
			return nil, fmt.Errorf("cacher %s doesn't support more than one IndexerFunc: ", reflect.TypeOf(obj).String())
		}
		for key, value := range config.IndexerFuncs {
			if value != nil {
				indexedTrigger = &indexedTriggerFunc{
					indexName:   key,
					indexerFunc: value,
				}
			}
		}
	}

	if config.Clock == nil {
		config.Clock = clock.RealClock{}
	}
	objType := reflect.TypeOf(obj)
	cacher := &Cacher{
		resourcePrefix: config.ResourcePrefix,
		ready:          newReady(),
		storage:        config.Storage,
		objectType:     objType,
		groupResource:  config.GroupResource,
		versioner:      config.Versioner,
		newFunc:        config.NewFunc,
		indexedTrigger: indexedTrigger,
		watcherIdx:     0,
		watchers: indexedWatchers{
			allWatchers:   make(map[int]*cacheWatcher),
			valueWatchers: make(map[string]watchersMap),
		},
		// TODO: Figure out the correct value for the buffer size.
		incoming:              make(chan watchCacheEvent, 100),
		dispatchTimeoutBudget: newTimeBudget(),
		// We need to (potentially) stop both:
		// - wait.Until go-routine
		// - reflector.ListAndWatch
		// and there are no guarantees on the order that they will stop.
		// So we will be simply closing the channel, and synchronizing on the WaitGroup.
		stopCh:           stopCh,
		clock:            config.Clock,
		timer:            time.NewTimer(time.Duration(0)),
		bookmarkWatchers: newTimeBucketWatchers(config.Clock, defaultBookmarkFrequency),
	}

	// Ensure that timer is stopped.
	if !cacher.timer.Stop() {
		// Consume triggered (but not yet received) timer event
		// so that future reuse does not get a spurious timeout.
		<-cacher.timer.C
	}

	watchCache := newWatchCache(
		config.KeyFunc, cacher.processEvent, config.GetAttrsFunc, config.Versioner, config.Indexers, config.Clock, config.GroupResource)
	listerWatcher := NewCacherListerWatcher(config.Storage, config.ResourcePrefix, config.NewListFunc)
	reflectorName := "storage/cacher.go:" + config.ResourcePrefix

	reflector := cache.NewNamedReflector(reflectorName, listerWatcher, obj, watchCache, 0)
	// Configure reflector's pager to for an appropriate pagination chunk size for fetching data from
	// storage. The pager falls back to full list if paginated list calls fail due to an "Expired" error.
	reflector.WatchListPageSize = storageWatchListPageSize
	// When etcd loses leader for 3 cycles, it returns error "no leader".
	// We don't want to terminate all watchers as recreating all watchers puts high load on api-server.
	// In most of the cases, leader is reelected within few cycles.
	reflector.MaxInternalErrorRetryDuration = time.Second * 30

	cacher.watchCache = watchCache
	cacher.reflector = reflector

	go cacher.dispatchEvents()

	cacher.stopWg.Add(1)
	go func() {
		defer cacher.stopWg.Done()
		defer cacher.terminateAllWatchers()
		wait.Until(
			func() {
				if !cacher.isStopped() {
					cacher.startCaching(stopCh)
				}
			}, time.Second, stopCh,
		)
	}()

	return cacher, nil
}

func (c *Cacher) startCaching(stopChannel <-chan struct{}) {
	// The 'usable' lock is always 'RLock'able when it is safe to use the cache.
	// It is safe to use the cache after a successful list until a disconnection.
	// We start with usable (write) locked. The below OnReplace function will
	// unlock it after a successful list. The below defer will then re-lock
	// it when this function exits (always due to disconnection), only if
	// we actually got a successful list. This cycle will repeat as needed.
	successfulList := false
	c.watchCache.SetOnReplace(func() {
		successfulList = true
		c.ready.set(true)
		klog.V(1).Infof("cacher (%v): initialized", c.groupResource.String())
		metrics.WatchCacheInitializations.WithLabelValues(c.groupResource.String()).Inc()
	})
	defer func() {
		if successfulList {
			c.ready.set(false)
		}
	}()

	c.terminateAllWatchers()
	// Note that since onReplace may be not called due to errors, we explicitly
	// need to retry it on errors under lock.
	// Also note that startCaching is called in a loop, so there's no need
	// to have another loop here.
	if err := c.reflector.ListAndWatch(stopChannel); err != nil {
		klog.Errorf("cacher (%v): unexpected ListAndWatch error: %v; reinitializing...", c.groupResource.String(), err)
	}
}

// Versioner implements storage.Interface.
func (c *Cacher) Versioner() storage.Versioner {
	return c.storage.Versioner()
}

// Create implements storage.Interface.
func (c *Cacher) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	return c.storage.Create(ctx, key, obj, out, ttl)
}

// Delete implements storage.Interface.
func (c *Cacher) Delete(
	ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions,
	validateDeletion storage.ValidateObjectFunc, _ runtime.Object) error {
	// Ignore the suggestion and try to pass down the current version of the object
	// read from cache.
	if elem, exists, err := c.watchCache.GetByKey(key); err != nil {
		klog.Errorf("GetByKey returned error: %v", err)
	} else if exists {
		// DeepCopy the object since we modify resource version when serializing the
		// current object.
		currObj := elem.(*storeElement).Object.DeepCopyObject()
		return c.storage.Delete(ctx, key, out, preconditions, validateDeletion, currObj)
	}
	// If we couldn't get the object, fallback to no-suggestion.
	return c.storage.Delete(ctx, key, out, preconditions, validateDeletion, nil)
}

// Watch implements storage.Interface.
func (c *Cacher) Watch(ctx context.Context, key string, opts storage.ListOptions) (watch.Interface, error) {
	pred := opts.Predicate
	watchRV, err := c.versioner.ParseResourceVersion(opts.ResourceVersion)
	if err != nil {
		return nil, err
	}

	if err := c.ready.wait(); err != nil {
		return nil, errors.NewServiceUnavailable(err.Error())
	}

	triggerValue, triggerSupported := "", false
	if c.indexedTrigger != nil {
		for _, field := range pred.IndexFields {
			if field == c.indexedTrigger.indexName {
				if value, ok := pred.Field.RequiresExactMatch(field); ok {
					triggerValue, triggerSupported = value, true
				}
			}
		}
	}

	// It boils down to a tradeoff between:
	// - having it as small as possible to reduce memory usage
	// - having it large enough to ensure that watchers that need to process
	//   a bunch of changes have enough buffer to avoid from blocking other
	//   watchers on our watcher having a processing hiccup
	chanSize := c.watchCache.suggestedWatchChannelSize(c.indexedTrigger != nil, triggerSupported)

	// Determine watch timeout('0' means deadline is not set, ignore checking)
	deadline, _ := ctx.Deadline()

	identifier := fmt.Sprintf("key: %q, labels: %q, fields: %q", key, pred.Label, pred.Field)

	// Create a watcher here to reduce memory allocations under lock,
	// given that memory allocation may trigger GC and block the thread.
	// Also note that emptyFunc is a placeholder, until we will be able
	// to compute watcher.forget function (which has to happen under lock).
	watcher := newCacheWatcher(
		chanSize,
		filterWithAttrsFunction(key, pred),
		emptyFunc,
		c.versioner,
		deadline,
		pred.AllowWatchBookmarks,
		c.groupResource,
		identifier,
	)

	// We explicitly use thread unsafe version and do locking ourself to ensure that
	// no new events will be processed in the meantime. The watchCache will be unlocked
	// on return from this function.
	// Note that we cannot do it under Cacher lock, to avoid a deadlock, since the
	// underlying watchCache is calling processEvent under its lock.
	c.watchCache.RLock()
	defer c.watchCache.RUnlock()
	cacheInterval, err := c.watchCache.getAllEventsSinceLocked(watchRV)
	if err != nil {
		// To match the uncached watch implementation, once we have passed authn/authz/admission,
		// and successfully parsed a resource version, other errors must fail with a watch event of type ERROR,
		// rather than a directly returned error.
		return newErrWatcher(err), nil
	}

	func() {
		c.Lock()
		defer c.Unlock()
		// Update watcher.forget function once we can compute it.
		watcher.forget = forgetWatcher(c, watcher, c.watcherIdx, triggerValue, triggerSupported)
		c.watchers.addWatcher(watcher, c.watcherIdx, triggerValue, triggerSupported)

		// Add it to the queue only when the client support watch bookmarks.
		if watcher.allowWatchBookmarks {
			c.bookmarkWatchers.addWatcher(watcher)
		}
		c.watcherIdx++
	}()

	go watcher.processInterval(ctx, cacheInterval, watchRV)
	return watcher, nil
}

// Get implements storage.Interface.
func (c *Cacher) Get(ctx context.Context, key string, opts storage.GetOptions, objPtr runtime.Object) error {
	if opts.ResourceVersion == "" {
		// If resourceVersion is not specified, serve it from underlying
		// storage (for backward compatibility).
		return c.storage.Get(ctx, key, opts, objPtr)
	}

	// If resourceVersion is specified, serve it from cache.
	// It's guaranteed that the returned value is at least that
	// fresh as the given resourceVersion.
	getRV, err := c.versioner.ParseResourceVersion(opts.ResourceVersion)
	if err != nil {
		return err
	}

	if getRV == 0 && !c.ready.check() {
		// If Cacher is not yet initialized and we don't require any specific
		// minimal resource version, simply forward the request to storage.
		return c.storage.Get(ctx, key, opts, objPtr)
	}

	// Do not create a trace - it's not for free and there are tons
	// of Get requests. We can add it if it will be really needed.
	if err := c.ready.wait(); err != nil {
		return errors.NewServiceUnavailable(err.Error())
	}

	objVal, err := conversion.EnforcePtr(objPtr)
	if err != nil {
		return err
	}

	obj, exists, readResourceVersion, err := c.watchCache.WaitUntilFreshAndGet(ctx, getRV, key)
	if err != nil {
		return err
	}

	if exists {
		elem, ok := obj.(*storeElement)
		if !ok {
			return fmt.Errorf("non *storeElement returned from storage: %v", obj)
		}
		objVal.Set(reflect.ValueOf(elem.Object).Elem())
	} else {
		objVal.Set(reflect.Zero(objVal.Type()))
		if !opts.IgnoreNotFound {
			return storage.NewKeyNotFoundError(key, int64(readResourceVersion))
		}
	}
	return nil
}

// NOTICE: Keep in sync with shouldListFromStorage function in
//
//	staging/src/k8s.io/apiserver/pkg/util/flowcontrol/request/list_work_estimator.go
func shouldDelegateList(opts storage.ListOptions) bool {
	resourceVersion := opts.ResourceVersion
	pred := opts.Predicate
	pagingEnabled := utilfeature.DefaultFeatureGate.Enabled(features.APIListChunking)
	hasContinuation := pagingEnabled && len(pred.Continue) > 0
	hasLimit := pagingEnabled && pred.Limit > 0 && resourceVersion != "0"

	// If resourceVersion is not specified, serve it from underlying
	// storage (for backward compatibility). If a continuation is
	// requested, serve it from the underlying storage as well.
	// Limits are only sent to storage when resourceVersion is non-zero
	// since the watch cache isn't able to perform continuations, and
	// limits are ignored when resource version is zero
	return resourceVersion == "" || hasContinuation || hasLimit || opts.ResourceVersionMatch == metav1.ResourceVersionMatchExact
}

func (c *Cacher) listItems(ctx context.Context, listRV uint64, key string, pred storage.SelectionPredicate, recursive bool) ([]interface{}, uint64, string, error) {
	if !recursive {
		obj, exists, readResourceVersion, err := c.watchCache.WaitUntilFreshAndGet(ctx, listRV, key)
		if err != nil {
			return nil, 0, "", err
		}
		if exists {
			return []interface{}{obj}, readResourceVersion, "", nil
		}
		return nil, readResourceVersion, "", nil
	}
	return c.watchCache.WaitUntilFreshAndList(ctx, listRV, pred.MatcherIndex())
}

// GetList implements storage.Interface
func (c *Cacher) GetList(ctx context.Context, key string, opts storage.ListOptions, listObj runtime.Object) error {
	recursive := opts.Recursive
	resourceVersion := opts.ResourceVersion
	pred := opts.Predicate
	if shouldDelegateList(opts) {
		return c.storage.GetList(ctx, key, opts, listObj)
	}

	// If resourceVersion is specified, serve it from cache.
	// It's guaranteed that the returned value is at least that
	// fresh as the given resourceVersion.
	listRV, err := c.versioner.ParseResourceVersion(resourceVersion)
	if err != nil {
		return err
	}

	if listRV == 0 && !c.ready.check() {
		// If Cacher is not yet initialized and we don't require any specific
		// minimal resource version, simply forward the request to storage.
		return c.storage.GetList(ctx, key, opts, listObj)
	}

	ctx, span := tracing.Start(ctx, "cacher list",
		attribute.String("audit-id", audit.GetAuditIDTruncated(ctx)),
		attribute.Stringer("type", c.groupResource))
	defer span.End(500 * time.Millisecond)

	if err := c.ready.wait(); err != nil {
		return errors.NewServiceUnavailable(err.Error())
	}
	span.AddEvent("Ready")

	// List elements with at least 'listRV' from cache.
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return err
	}
	listVal, err := conversion.EnforcePtr(listPtr)
	if err != nil {
		return err
	}
	if listVal.Kind() != reflect.Slice {
		return fmt.Errorf("need a pointer to slice, got %v", listVal.Kind())
	}
	filter := filterWithAttrsFunction(key, pred)

	objs, readResourceVersion, indexUsed, err := c.listItems(ctx, listRV, key, pred, recursive)
	if err != nil {
		return err
	}
	span.AddEvent("Listed items from cache", attribute.Int("count", len(objs)))
	if len(objs) > listVal.Cap() && pred.Label.Empty() && pred.Field.Empty() {
		// Resize the slice appropriately, since we already know that none
		// of the elements will be filtered out.
		listVal.Set(reflect.MakeSlice(reflect.SliceOf(c.objectType.Elem()), 0, len(objs)))
		span.AddEvent("Resized result")
	}
	for _, obj := range objs {
		elem, ok := obj.(*storeElement)
		if !ok {
			return fmt.Errorf("non *storeElement returned from storage: %v", obj)
		}
		if filter(elem.Key, elem.Labels, elem.Fields) {
			listVal.Set(reflect.Append(listVal, reflect.ValueOf(elem.Object).Elem()))
		}
	}
	span.AddEvent("Filtered items", attribute.Int("count", listVal.Len()))
	if c.versioner != nil {
		if err := c.versioner.UpdateList(listObj, readResourceVersion, "", nil); err != nil {
			return err
		}
	}
	metrics.RecordListCacheMetrics(c.resourcePrefix, indexUsed, len(objs), listVal.Len())
	return nil
}

// GuaranteedUpdate implements storage.Interface.
func (c *Cacher) GuaranteedUpdate(
	ctx context.Context, key string, destination runtime.Object, ignoreNotFound bool,
	preconditions *storage.Preconditions, tryUpdate storage.UpdateFunc, _ runtime.Object) error {
	// Ignore the suggestion and try to pass down the current version of the object
	// read from cache.
	if elem, exists, err := c.watchCache.GetByKey(key); err != nil {
		klog.Errorf("GetByKey returned error: %v", err)
	} else if exists {
		// DeepCopy the object since we modify resource version when serializing the
		// current object.
		currObj := elem.(*storeElement).Object.DeepCopyObject()
		return c.storage.GuaranteedUpdate(ctx, key, destination, ignoreNotFound, preconditions, tryUpdate, currObj)
	}
	// If we couldn't get the object, fallback to no-suggestion.
	return c.storage.GuaranteedUpdate(ctx, key, destination, ignoreNotFound, preconditions, tryUpdate, nil)
}

// Count implements storage.Interface.
func (c *Cacher) Count(pathPrefix string) (int64, error) {
	return c.storage.Count(pathPrefix)
}

// baseObjectThreadUnsafe omits locking for cachingObject.
func baseObjectThreadUnsafe(object runtime.Object) runtime.Object {
	if co, ok := object.(*cachingObject); ok {
		return co.object
	}
	return object
}

func (c *Cacher) triggerValuesThreadUnsafe(event *watchCacheEvent) ([]string, bool) {
	if c.indexedTrigger == nil {
		return nil, false
	}

	result := make([]string, 0, 2)
	result = append(result, c.indexedTrigger.indexerFunc(baseObjectThreadUnsafe(event.Object)))
	if event.PrevObject == nil {
		return result, true
	}
	prevTriggerValue := c.indexedTrigger.indexerFunc(baseObjectThreadUnsafe(event.PrevObject))
	if result[0] != prevTriggerValue {
		result = append(result, prevTriggerValue)
	}
	return result, true
}

func (c *Cacher) processEvent(event *watchCacheEvent) {
	if curLen := int64(len(c.incoming)); c.incomingHWM.Update(curLen) {
		// Monitor if this gets backed up, and how much.
		klog.V(1).Infof("cacher (%v): %v objects queued in incoming channel.", c.groupResource.String(), curLen)
	}
	c.incoming <- *event
}

func (c *Cacher) dispatchEvents() {
	// Jitter to help level out any aggregate load.
	bookmarkTimer := c.clock.NewTimer(wait.Jitter(time.Second, 0.25))
	defer bookmarkTimer.Stop()

	lastProcessedResourceVersion := uint64(0)
	for {
		select {
		case event, ok := <-c.incoming:
			if !ok {
				return
			}
			// Don't dispatch bookmarks coming from the storage layer.
			// They can be very frequent (even to the level of subseconds)
			// to allow efficient watch resumption on kube-apiserver restarts,
			// and propagating them down may overload the whole system.
			//
			// TODO: If at some point we decide the performance and scalability
			// footprint is acceptable, this is the place to hook them in.
			// However, we then need to check if this was called as a result
			// of a bookmark event or regular Add/Update/Delete operation by
			// checking if resourceVersion here has changed.
			if event.Type != watch.Bookmark {
				c.dispatchEvent(&event)
			}
			lastProcessedResourceVersion = event.ResourceVersion
			metrics.EventsCounter.WithLabelValues(c.groupResource.String()).Inc()
		case <-bookmarkTimer.C():
			bookmarkTimer.Reset(wait.Jitter(time.Second, 0.25))
			// Never send a bookmark event if we did not see an event here, this is fine
			// because we don't provide any guarantees on sending bookmarks.
			if lastProcessedResourceVersion == 0 {
				// pop expired watchers in case there has been no update
				c.bookmarkWatchers.popExpiredWatchers()
				continue
			}
			bookmarkEvent := &watchCacheEvent{
				Type:            watch.Bookmark,
				Object:          c.newFunc(),
				ResourceVersion: lastProcessedResourceVersion,
			}
			if err := c.versioner.UpdateObject(bookmarkEvent.Object, bookmarkEvent.ResourceVersion); err != nil {
				klog.Errorf("failure to set resourceVersion to %d on bookmark event %+v", bookmarkEvent.ResourceVersion, bookmarkEvent.Object)
				continue
			}
			c.dispatchEvent(bookmarkEvent)
		case <-c.stopCh:
			return
		}
	}
}

func setCachingObjects(event *watchCacheEvent, versioner storage.Versioner) {
	switch event.Type {
	case watch.Added, watch.Modified:
		if object, err := newCachingObject(event.Object); err == nil {
			event.Object = object
		} else {
			klog.Errorf("couldn't create cachingObject from: %#v", event.Object)
		}
		// Don't wrap PrevObject for update event (for create events it is nil).
		// We only encode those to deliver DELETE watch events, so if
		// event.Object is not nil it can be used only for watchers for which
		// selector was satisfied for its previous version and is no longer
		// satisfied for the current version.
		// This is rare enough that it doesn't justify making deep-copy of the
		// object (done by newCachingObject) every time.
	case watch.Deleted:
		// Don't wrap Object for delete events - these are not to deliver any
		// events. Only wrap PrevObject.
		if object, err := newCachingObject(event.PrevObject); err == nil {
			// Update resource version of the object.
			// event.PrevObject is used to deliver DELETE watch events and
			// for them, we set resourceVersion to <current> instead of
			// the resourceVersion of the last modification of the object.
			updateResourceVersion(object, versioner, event.ResourceVersion)
			event.PrevObject = object
		} else {
			klog.Errorf("couldn't create cachingObject from: %#v", event.Object)
		}
	}
}

func (c *Cacher) dispatchEvent(event *watchCacheEvent) {
	c.startDispatching(event)
	defer c.finishDispatching()
	// Watchers stopped after startDispatching will be delayed to finishDispatching,

	// Since add() can block, we explicitly add when cacher is unlocked.
	// Dispatching event in nonblocking way first, which make faster watchers
	// not be blocked by slower ones.
	if event.Type == watch.Bookmark {
		for _, watcher := range c.watchersBuffer {
			watcher.nonblockingAdd(event)
		}
	} else {
		// Set up caching of object serializations only for dispatching this event.
		//
		// Storing serializations in memory would result in increased memory usage,
		// but it would help for caching encodings for watches started from old
		// versions. However, we still don't have a convincing data that the gain
		// from it justifies increased memory usage, so for now we drop the cached
		// serializations after dispatching this event.
		//
		// Given that CachingObject is just wrapping the object and not perfoming
		// deep-copying (until some field is explicitly being modified), we create
		// it unconditionally to ensure safety and reduce deep-copying.
		//
		// Make a shallow copy to allow overwriting Object and PrevObject.
		wcEvent := *event
		setCachingObjects(&wcEvent, c.versioner)
		event = &wcEvent

		c.blockedWatchers = c.blockedWatchers[:0]
		for _, watcher := range c.watchersBuffer {
			if !watcher.nonblockingAdd(event) {
				c.blockedWatchers = append(c.blockedWatchers, watcher)
			}
		}

		if len(c.blockedWatchers) > 0 {
			// dispatchEvent is called very often, so arrange
			// to reuse timers instead of constantly allocating.
			startTime := time.Now()
			timeout := c.dispatchTimeoutBudget.takeAvailable()
			c.timer.Reset(timeout)

			// Send event to all blocked watchers. As long as timer is running,
			// `add` will wait for the watcher to unblock. After timeout,
			// `add` will not wait, but immediately close a still blocked watcher.
			// Hence, every watcher gets the chance to unblock itself while timer
			// is running, not only the first ones in the list.
			timer := c.timer
			for _, watcher := range c.blockedWatchers {
				if !watcher.add(event, timer) {
					// fired, clean the timer by set it to nil.
					timer = nil
				}
			}

			// Stop the timer if it is not fired
			if timer != nil && !timer.Stop() {
				// Consume triggered (but not yet received) timer event
				// so that future reuse does not get a spurious timeout.
				<-timer.C
			}

			c.dispatchTimeoutBudget.returnUnused(timeout - time.Since(startTime))
		}
	}
}

func (c *Cacher) startDispatchingBookmarkEventsLocked() {
	// Pop already expired watchers. However, explicitly ignore stopped ones,
	// as we don't delete watcher from bookmarkWatchers when it is stopped.
	for _, watchers := range c.bookmarkWatchers.popExpiredWatchers() {
		for _, watcher := range watchers {
			// c.Lock() is held here.
			// watcher.stopThreadUnsafe() is protected by c.Lock()
			if watcher.stopped {
				continue
			}
			c.watchersBuffer = append(c.watchersBuffer, watcher)
			c.expiredBookmarkWatchers = append(c.expiredBookmarkWatchers, watcher)
		}
	}
}

// startDispatching chooses watchers potentially interested in a given event
// a marks dispatching as true.
func (c *Cacher) startDispatching(event *watchCacheEvent) {
	// It is safe to call triggerValuesThreadUnsafe here, because at this
	// point only this thread can access this event (we create a separate
	// watchCacheEvent for every dispatch).
	triggerValues, supported := c.triggerValuesThreadUnsafe(event)

	c.Lock()
	defer c.Unlock()

	c.dispatching = true
	// We are reusing the slice to avoid memory reallocations in every
	// dispatchEvent() call. That may prevent Go GC from freeing items
	// from previous phases that are sitting behind the current length
	// of the slice, but there is only a limited number of those and the
	// gain from avoiding memory allocations is much bigger.
	c.watchersBuffer = c.watchersBuffer[:0]

	if event.Type == watch.Bookmark {
		c.startDispatchingBookmarkEventsLocked()
		// return here to reduce following code indentation and diff
		return
	}

	// Iterate over "allWatchers" no matter what the trigger function is.
	for _, watcher := range c.watchers.allWatchers {
		c.watchersBuffer = append(c.watchersBuffer, watcher)
	}
	if supported {
		// Iterate over watchers interested in the given values of the trigger.
		for _, triggerValue := range triggerValues {
			for _, watcher := range c.watchers.valueWatchers[triggerValue] {
				c.watchersBuffer = append(c.watchersBuffer, watcher)
			}
		}
	} else {
		// supported equal to false generally means that trigger function
		// is not defined (or not aware of any indexes). In this case,
		// watchers filters should generally also don't generate any
		// trigger values, but can cause problems in case of some
		// misconfiguration. Thus we paranoidly leave this branch.

		// Iterate over watchers interested in exact values for all values.
		for _, watchers := range c.watchers.valueWatchers {
			for _, watcher := range watchers {
				c.watchersBuffer = append(c.watchersBuffer, watcher)
			}
		}
	}
}

// finishDispatching stops all the watchers that were supposed to be
// stopped in the meantime, but it was deferred to avoid closing input
// channels of watchers, as add() may still have writing to it.
// It also marks dispatching as false.
func (c *Cacher) finishDispatching() {
	c.Lock()
	defer c.Unlock()
	c.dispatching = false
	for _, watcher := range c.watchersToStop {
		watcher.stopLocked()
	}
	c.watchersToStop = c.watchersToStop[:0]

	for _, watcher := range c.expiredBookmarkWatchers {
		if watcher.stopped {
			continue
		}
		// requeue the watcher for the next bookmark if needed.
		c.bookmarkWatchers.addWatcher(watcher)
	}
	c.expiredBookmarkWatchers = c.expiredBookmarkWatchers[:0]
}

func (c *Cacher) terminateAllWatchers() {
	c.Lock()
	defer c.Unlock()
	c.watchers.terminateAll(c.groupResource, c.stopWatcherLocked)
}

func (c *Cacher) stopWatcherLocked(watcher *cacheWatcher) {
	if c.dispatching {
		c.watchersToStop = append(c.watchersToStop, watcher)
	} else {
		watcher.stopLocked()
	}
}

func (c *Cacher) isStopped() bool {
	c.stopLock.RLock()
	defer c.stopLock.RUnlock()
	return c.stopped
}

// Stop implements the graceful termination.
func (c *Cacher) Stop() {
	c.stopLock.Lock()
	if c.stopped {
		// avoid stopping twice (note: cachers are shared with subresources)
		c.stopLock.Unlock()
		return
	}
	c.stopped = true
	c.ready.stop()
	c.stopLock.Unlock()
	close(c.stopCh)
	c.stopWg.Wait()
}

func forgetWatcher(c *Cacher, w *cacheWatcher, index int, triggerValue string, triggerSupported bool) func(bool) {
	return func(drainWatcher bool) {
		c.Lock()
		defer c.Unlock()

		w.setDrainInputBufferLocked(drainWatcher)

		// It's possible that the watcher is already not in the structure (e.g. in case of
		// simultaneous Stop() and terminateAllWatchers(), but it is safe to call stopLocked()
		// on a watcher multiple times.
		c.watchers.deleteWatcher(index, triggerValue, triggerSupported, c.stopWatcherLocked)
	}
}

func filterWithAttrsFunction(key string, p storage.SelectionPredicate) filterWithAttrsFunc {
	filterFunc := func(objKey string, label labels.Set, field fields.Set) bool {
		if !hasPathPrefix(objKey, key) {
			return false
		}
		return p.MatchesObjectAttributes(label, field)
	}
	return filterFunc
}

// LastSyncResourceVersion returns resource version to which the underlying cache is synced.
func (c *Cacher) LastSyncResourceVersion() (uint64, error) {
	if err := c.ready.wait(); err != nil {
		return 0, errors.NewServiceUnavailable(err.Error())
	}

	resourceVersion := c.reflector.LastSyncResourceVersion()
	return c.versioner.ParseResourceVersion(resourceVersion)
}

// cacherListerWatcher opaques storage.Interface to expose cache.ListerWatcher.
type cacherListerWatcher struct {
	storage        storage.Interface
	resourcePrefix string
	newListFunc    func() runtime.Object
}

// NewCacherListerWatcher returns a storage.Interface backed ListerWatcher.
func NewCacherListerWatcher(storage storage.Interface, resourcePrefix string, newListFunc func() runtime.Object) cache.ListerWatcher {
	return &cacherListerWatcher{
		storage:        storage,
		resourcePrefix: resourcePrefix,
		newListFunc:    newListFunc,
	}
}

// Implements cache.ListerWatcher interface.
func (lw *cacherListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	list := lw.newListFunc()
	pred := storage.SelectionPredicate{
		Label:    labels.Everything(),
		Field:    fields.Everything(),
		Limit:    options.Limit,
		Continue: options.Continue,
	}

	storageOpts := storage.ListOptions{
		ResourceVersionMatch: options.ResourceVersionMatch,
		Predicate:            pred,
		Recursive:            true,
	}
	if err := lw.storage.GetList(context.TODO(), lw.resourcePrefix, storageOpts, list); err != nil {
		return nil, err
	}
	return list, nil
}

// Implements cache.ListerWatcher interface.
func (lw *cacherListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	opts := storage.ListOptions{
		ResourceVersion: options.ResourceVersion,
		Predicate:       storage.Everything,
		Recursive:       true,
		ProgressNotify:  true,
	}
	return lw.storage.Watch(context.TODO(), lw.resourcePrefix, opts)
}

// errWatcher implements watch.Interface to return a single error
type errWatcher struct {
	result chan watch.Event
}

func newErrWatcher(err error) *errWatcher {
	// Create an error event
	errEvent := watch.Event{Type: watch.Error}
	switch err := err.(type) {
	case runtime.Object:
		errEvent.Object = err
	case *errors.StatusError:
		errEvent.Object = &err.ErrStatus
	default:
		errEvent.Object = &metav1.Status{
			Status:  metav1.StatusFailure,
			Message: err.Error(),
			Reason:  metav1.StatusReasonInternalError,
			Code:    http.StatusInternalServerError,
		}
	}

	// Create a watcher with room for a single event, populate it, and close the channel
	watcher := &errWatcher{result: make(chan watch.Event, 1)}
	watcher.result <- errEvent
	close(watcher.result)

	return watcher
}

// Implements watch.Interface.
func (c *errWatcher) ResultChan() <-chan watch.Event {
	return c.result
}

// Implements watch.Interface.
func (c *errWatcher) Stop() {
	// no-op
}

// cacheWatcher implements watch.Interface
// this is not thread-safe
type cacheWatcher struct {
	input     chan *watchCacheEvent
	result    chan watch.Event
	done      chan struct{}
	filter    filterWithAttrsFunc
	stopped   bool
	forget    func(bool)
	versioner storage.Versioner
	// The watcher will be closed by server after the deadline,
	// save it here to send bookmark events before that.
	deadline            time.Time
	allowWatchBookmarks bool
	groupResource       schema.GroupResource

	// human readable identifier that helps assigning cacheWatcher
	// instance with request
	identifier string

	// drainInputBuffer indicates whether we should delay closing this watcher
	// and send all event in the input buffer.
	drainInputBuffer bool
}

func newCacheWatcher(
	chanSize int,
	filter filterWithAttrsFunc,
	forget func(bool),
	versioner storage.Versioner,
	deadline time.Time,
	allowWatchBookmarks bool,
	groupResource schema.GroupResource,
	identifier string,
) *cacheWatcher {
	return &cacheWatcher{
		input:               make(chan *watchCacheEvent, chanSize),
		result:              make(chan watch.Event, chanSize),
		done:                make(chan struct{}),
		filter:              filter,
		stopped:             false,
		forget:              forget,
		versioner:           versioner,
		deadline:            deadline,
		allowWatchBookmarks: allowWatchBookmarks,
		groupResource:       groupResource,
		identifier:          identifier,
	}
}

// Implements watch.Interface.
func (c *cacheWatcher) ResultChan() <-chan watch.Event {
	return c.result
}

// Implements watch.Interface.
func (c *cacheWatcher) Stop() {
	c.forget(false)
}

// we rely on the fact that stopLocked is actually protected by Cacher.Lock()
func (c *cacheWatcher) stopLocked() {
	if !c.stopped {
		c.stopped = true
		// stop without draining the input channel was requested.
		if !c.drainInputBuffer {
			close(c.done)
		}
		close(c.input)
	}

	// Even if the watcher was already stopped, if it previously was
	// using draining mode and it's not using it now we need to
	// close the done channel now. Otherwise we could leak the
	// processing goroutine if it will be trying to put more objects
	// into result channel, the channel will be full and there will
	// already be noone on the processing the events on the receiving end.
	if !c.drainInputBuffer && !c.isDoneChannelClosedLocked() {
		close(c.done)
	}
}

func (c *cacheWatcher) nonblockingAdd(event *watchCacheEvent) bool {
	select {
	case c.input <- event:
		return true
	default:
		return false
	}
}

// Nil timer means that add will not block (if it can't send event immediately, it will break the watcher)
func (c *cacheWatcher) add(event *watchCacheEvent, timer *time.Timer) bool {
	// Try to send the event immediately, without blocking.
	if c.nonblockingAdd(event) {
		return true
	}

	closeFunc := func() {
		// This means that we couldn't send event to that watcher.
		// Since we don't want to block on it infinitely,
		// we simply terminate it.
		klog.V(1).Infof("Forcing %v watcher close due to unresponsiveness: %v. len(c.input) = %v, len(c.result) = %v", c.groupResource.String(), c.identifier, len(c.input), len(c.result))
		metrics.TerminatedWatchersCounter.WithLabelValues(c.groupResource.String()).Inc()
		c.forget(false)
	}

	if timer == nil {
		closeFunc()
		return false
	}

	// OK, block sending, but only until timer fires.
	select {
	case c.input <- event:
		return true
	case <-timer.C:
		closeFunc()
		return false
	}
}

func (c *cacheWatcher) nextBookmarkTime(now time.Time, bookmarkFrequency time.Duration) (time.Time, bool) {
	// We try to send bookmarks:
	//
	// (a) right before the watcher timeout - for now we simply set it 2s before
	//     the deadline
	//
	// (b) roughly every minute
	//
	// (b) gives us periodicity if the watch breaks due to unexpected
	// conditions, (a) ensures that on timeout the watcher is as close to
	// now as possible - this covers 99% of cases.

	heartbeatTime := now.Add(bookmarkFrequency)
	if c.deadline.IsZero() {
		// Timeout is set by our client libraries (e.g. reflector) as well as defaulted by
		// apiserver if properly configured. So this shoudln't happen in practice.
		return heartbeatTime, true
	}
	if pretimeoutTime := c.deadline.Add(-2 * time.Second); pretimeoutTime.Before(heartbeatTime) {
		heartbeatTime = pretimeoutTime
	}

	if heartbeatTime.Before(now) {
		return time.Time{}, false
	}
	return heartbeatTime, true
}

// setDrainInputBufferLocked if set to true indicates that we should delay closing this watcher
// until we send all events residing in the input buffer.
func (c *cacheWatcher) setDrainInputBufferLocked(drain bool) {
	c.drainInputBuffer = drain
}

// isDoneChannelClosed checks if c.done channel is closed
func (c *cacheWatcher) isDoneChannelClosedLocked() bool {
	select {
	case <-c.done:
		return true
	default:
	}
	return false
}

func getMutableObject(object runtime.Object) runtime.Object {
	if _, ok := object.(*cachingObject); ok {
		// It is safe to return without deep-copy, because the underlying
		// object will lazily perform deep-copy on the first try to change
		// any of its fields.
		return object
	}
	return object.DeepCopyObject()
}

func updateResourceVersion(object runtime.Object, versioner storage.Versioner, resourceVersion uint64) {
	if err := versioner.UpdateObject(object, resourceVersion); err != nil {
		utilruntime.HandleError(fmt.Errorf("failure to version api object (%d) %#v: %v", resourceVersion, object, err))
	}
}

func (c *cacheWatcher) convertToWatchEvent(event *watchCacheEvent) *watch.Event {
	if event.Type == watch.Bookmark {
		return &watch.Event{Type: watch.Bookmark, Object: event.Object.DeepCopyObject()}
	}

	curObjPasses := event.Type != watch.Deleted && c.filter(event.Key, event.ObjLabels, event.ObjFields)
	oldObjPasses := false
	if event.PrevObject != nil {
		oldObjPasses = c.filter(event.Key, event.PrevObjLabels, event.PrevObjFields)
	}
	if !curObjPasses && !oldObjPasses {
		// Watcher is not interested in that object.
		return nil
	}

	switch {
	case curObjPasses && !oldObjPasses:
		return &watch.Event{Type: watch.Added, Object: getMutableObject(event.Object)}
	case curObjPasses && oldObjPasses:
		return &watch.Event{Type: watch.Modified, Object: getMutableObject(event.Object)}
	case !curObjPasses && oldObjPasses:
		// return a delete event with the previous object content, but with the event's resource version
		oldObj := getMutableObject(event.PrevObject)
		// We know that if oldObj is cachingObject (which can only be set via
		// setCachingObjects), its resourceVersion is already set correctly and
		// we don't need to update it. However, since cachingObject efficiently
		// handles noop updates, we avoid this microoptimization here.
		updateResourceVersion(oldObj, c.versioner, event.ResourceVersion)
		return &watch.Event{Type: watch.Deleted, Object: oldObj}
	}

	return nil
}

// NOTE: sendWatchCacheEvent is assumed to not modify <event> !!!
func (c *cacheWatcher) sendWatchCacheEvent(event *watchCacheEvent) {
	watchEvent := c.convertToWatchEvent(event)
	if watchEvent == nil {
		// Watcher is not interested in that object.
		return
	}

	// We need to ensure that if we put event X to the c.result, all
	// previous events were already put into it before, no matter whether
	// c.done is close or not.
	// Thus we cannot simply select from c.done and c.result and this
	// would give us non-determinism.
	// At the same time, we don't want to block infinitely on putting
	// to c.result, when c.done is already closed.
	//
	// This ensures that with c.done already close, we at most once go
	// into the next select after this. With that, no matter which
	// statement we choose there, we will deliver only consecutive
	// events.
	select {
	case <-c.done:
		return
	default:
	}

	select {
	case c.result <- *watchEvent:
	case <-c.done:
	}
}

func (c *cacheWatcher) processInterval(ctx context.Context, cacheInterval *watchCacheInterval, resourceVersion uint64) {
	defer utilruntime.HandleCrash()
	defer close(c.result)
	defer c.Stop()

	// Check how long we are processing initEvents.
	// As long as these are not processed, we are not processing
	// any incoming events, so if it takes long, we may actually
	// block all watchers for some time.
	// TODO: From the logs it seems that there happens processing
	// times even up to 1s which is very long. However, this doesn't
	// depend that much on the number of initEvents. E.g. from the
	// 2000-node Kubemark run we have logs like this, e.g.:
	// ... processing 13862 initEvents took 66.808689ms
	// ... processing 14040 initEvents took 993.532539ms
	// We should understand what is blocking us in those cases (e.g.
	// is it lack of CPU, network, or sth else) and potentially
	// consider increase size of result buffer in those cases.
	const initProcessThreshold = 500 * time.Millisecond
	startTime := time.Now()

	initEventCount := 0
	for {
		event, err := cacheInterval.Next()
		if err != nil {
			// An error indicates that the cache interval
			// has been invalidated and can no longer serve
			// events.
			//
			// Initially we considered sending an "out-of-history"
			// Error event in this case, but because historically
			// such events weren't sent out of the watchCache, we
			// decided not to. This is still ok, because on watch
			// closure, the watcher will try to re-instantiate the
			// watch and then will get an explicit "out-of-history"
			// window. There is potential for optimization, but for
			// now, in order to be on the safe side and not break
			// custom clients, the cost of it is something that we
			// are fully accepting.
			klog.Warningf("couldn't retrieve watch event to serve: %#v", err)
			return
		}
		if event == nil {
			break
		}
		c.sendWatchCacheEvent(event)
		// With some events already sent, update resourceVersion so that
		// events that were buffered and not yet processed won't be delivered
		// to this watcher second time causing going back in time.
		resourceVersion = event.ResourceVersion
		initEventCount++
	}

	if initEventCount > 0 {
		metrics.InitCounter.WithLabelValues(c.groupResource.String()).Add(float64(initEventCount))
	}
	processingTime := time.Since(startTime)
	if processingTime > initProcessThreshold {
		klog.V(2).Infof("processing %d initEvents of %s (%s) took %v", initEventCount, c.groupResource, c.identifier, processingTime)
	}

	c.process(ctx, resourceVersion)
}

func (c *cacheWatcher) process(ctx context.Context, resourceVersion uint64) {
	// At this point we already start processing incoming watch events.
	// However, the init event can still be processed because their serialization
	// and sending to the client happens asynchrnously.
	// TODO: As describe in the KEP, we would like to estimate that by delaying
	//   the initialization signal proportionally to the number of events to
	//   process, but we're leaving this to the tuning phase.
	utilflowcontrol.WatchInitialized(ctx)

	for {
		select {
		case event, ok := <-c.input:
			if !ok {
				return
			}
			// only send events newer than resourceVersion
			if event.ResourceVersion > resourceVersion {
				c.sendWatchCacheEvent(event)
			}
		case <-ctx.Done():
			return
		}
	}
}
