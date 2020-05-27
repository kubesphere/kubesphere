package kubernetes

import (
	"errors"
	"fmt"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batch_v1 "k8s.io/api/batch/v1"
	batch_v1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/kiali/kiali/log"
)

type (
	// Inspired/reused from istio code:
	// https://github.com/istio/istio/blob/master/mixer/adapter/kubernetesenv/cache.go
	cacheController interface {
		// Control Cache
		Start()
		HasSynced() bool
		WaitForSync() bool
		Stop()

		// Business methods
		GetCronJobs(namespace string) ([]batch_v1beta1.CronJob, error)
		GetDeployment(namespace string, name string) (*appsv1.Deployment, error)
		GetDeployments(namespace string) ([]appsv1.Deployment, error)
		GetEndpoints(namespace, name string) (*v1.Endpoints, error)
		GetJobs(namespace string) ([]batch_v1.Job, error)
		GetPods(namespace string) ([]v1.Pod, error)
		GetReplicationControllers(namespace string) ([]v1.ReplicationController, error)
		GetReplicaSets(namespace string) ([]appsv1.ReplicaSet, error)
		GetService(namespace string, name string) (*v1.Service, error)
		GetServices(namespace string) ([]v1.Service, error)
		GetStatefulSet(namespace string, name string) (*appsv1.StatefulSet, error)
		GetStatefulSets(namespace string) ([]appsv1.StatefulSet, error)
	}

	controllerImpl struct {
		clientset       kube.Interface
		refreshDuration time.Duration
		stopChan        chan struct{}
		syncCount       int
		maxSyncCount    int
		isErrorState    bool
		lastError       error
		lastErrorLock   sync.Mutex
		controllers     map[string]cache.SharedIndexInformer
	}
)

var (
	lastCacheErrorLock sync.Mutex
	errorCallbacks     []func(error)
)

func init() {
	setupErrorHandlers()
	errorCallbacks = make([]func(error), 0)
}

func setupErrorHandlers() {
	nErrFunc := len(utilruntime.ErrorHandlers)
	customErrorHandler := make([]func(error), nErrFunc+1)
	for i, errorFunc := range utilruntime.ErrorHandlers {
		customErrorHandler[i] = errorFunc
	}
	customErrorHandler[nErrFunc] = func(err error) {
		for _, callback := range errorCallbacks {
			callback(err)
		}
	}
	utilruntime.ErrorHandlers = customErrorHandler
}

func registerErrorCallback(callback func(error)) {
	defer lastCacheErrorLock.Unlock()
	lastCacheErrorLock.Lock()
	errorCallbacks = append(errorCallbacks, callback)
}

func newCacheController(clientset kube.Interface, refreshDuration time.Duration) cacheController {
	newControllerImpl := controllerImpl{
		clientset:       clientset,
		refreshDuration: refreshDuration,
		stopChan:        nil,
		controllers:     initControllers(clientset, refreshDuration),
		syncCount:       0,
		maxSyncCount:    20, // Move this to config ? or this constant is good enough ?
	}
	registerErrorCallback(newControllerImpl.ErrorCallback)

	return &newControllerImpl
}

func initControllers(clientset kube.Interface, refreshDuration time.Duration) map[string]cache.SharedIndexInformer {
	sharedInformers := informers.NewSharedInformerFactory(clientset, refreshDuration)
	controllers := make(map[string]cache.SharedIndexInformer)
	controllers["Pod"] = sharedInformers.Core().V1().Pods().Informer()
	controllers["ReplicationController"] = sharedInformers.Core().V1().ReplicationControllers().Informer()
	controllers["Deployment"] = sharedInformers.Apps().V1().Deployments().Informer()
	controllers["ReplicaSet"] = sharedInformers.Apps().V1().ReplicaSets().Informer()
	controllers["StatefulSet"] = sharedInformers.Apps().V1().StatefulSets().Informer()
	controllers["Job"] = sharedInformers.Batch().V1().Jobs().Informer()
	controllers["CronJob"] = sharedInformers.Batch().V1beta1().CronJobs().Informer()
	controllers["Service"] = sharedInformers.Core().V1().Services().Informer()
	controllers["Endpoints"] = sharedInformers.Core().V1().Endpoints().Informer()

	return controllers
}

func (c *controllerImpl) Start() {
	if c.stopChan == nil {
		c.stopChan = make(chan struct{})
		go c.run(c.stopChan)
		log.Infof("K8S cache started")
	} else {
		log.Warningf("K8S cache is already running")
	}
}

func (c *controllerImpl) run(stop <-chan struct{}) {
	for _, cn := range c.controllers {
		go cn.Run(stop)
	}
	<-stop
	log.Infof("K8S cache stopped")
}

func (c *controllerImpl) HasSynced() bool {
	if c.syncCount > c.maxSyncCount {
		log.Errorf("Max attempts reached syncing cache. Error connecting to k8s API: %d > %d", c.syncCount, c.maxSyncCount)
		c.Stop()
		return false
	}
	hasSynced := true
	for _, cn := range c.controllers {
		hasSynced = hasSynced && cn.HasSynced()
	}
	if hasSynced {
		c.syncCount = 0
	} else {
		c.syncCount = c.syncCount + 1
	}
	return hasSynced
}

func (c *controllerImpl) WaitForSync() bool {
	return cache.WaitForCacheSync(c.stopChan, c.HasSynced)
}

func (c *controllerImpl) Stop() {
	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}
}

func (c *controllerImpl) ErrorCallback(err error) {
	if c.isErrorState == false {
		log.Warningf("Error callback received: %s", err)
		c.lastErrorLock.Lock()
		c.isErrorState = true
		c.lastError = err
		c.lastErrorLock.Unlock()
		c.Stop()
	}
}

func (c *controllerImpl) checkStateAndRetry() error {
	if c.isErrorState == false {
		return nil
	}

	// Retry of the cache is hold by one single goroutine
	c.lastErrorLock.Lock()
	if c.isErrorState == true {
		// ping to check if backend is still unavailable (used namespace endpoint)
		_, err := c.clientset.CoreV1().Namespaces().List(emptyListOptions)
		if err != nil {
			c.lastError = fmt.Errorf("Error retrying to connect to K8S API backend. %s", err)
		} else {
			c.lastError = nil
			c.isErrorState = false
			c.Start()
			c.WaitForSync()
		}
	}
	c.lastErrorLock.Unlock()
	return c.lastError
}

func (c *controllerImpl) GetCronJobs(namespace string) ([]batch_v1beta1.CronJob, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []batch_v1beta1.CronJob{}, err
	}
	indexer := c.controllers["CronJob"].GetIndexer()
	cronjobs, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []batch_v1beta1.CronJob{}, err
	}
	if len(cronjobs) > 0 {
		_, ok := cronjobs[0].(*batch_v1beta1.CronJob)
		if !ok {
			return []batch_v1beta1.CronJob{}, errors.New("Bad CronJob type found in cache")
		}
		nsCronjobs := make([]batch_v1beta1.CronJob, len(cronjobs))
		for i, cronjob := range cronjobs {
			nsCronjobs[i] = *(cronjob.(*batch_v1beta1.CronJob))
		}
		return nsCronjobs, nil
	}
	return []batch_v1beta1.CronJob{}, nil
}

func (c *controllerImpl) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return nil, err
	}
	indexer := c.controllers["Deployment"].GetIndexer()
	deps, exist, err := indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if exist {
		dep, ok := deps.(*appsv1.Deployment)
		if !ok {
			return nil, errors.New("Bad Deployment type found in cache")
		}
		return dep, nil
	}
	return nil, NewNotFound(name, "apps/v1", "Deployment")
}

func (c *controllerImpl) GetDeployments(namespace string) ([]appsv1.Deployment, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []appsv1.Deployment{}, err
	}
	indexer := c.controllers["Deployment"].GetIndexer()
	deps, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []appsv1.Deployment{}, err
	}
	if len(deps) > 0 {
		_, ok := deps[0].(*appsv1.Deployment)
		if !ok {
			return nil, errors.New("Bad Deployment type found in cache")
		}
		nsDeps := make([]appsv1.Deployment, len(deps))
		for i, dep := range deps {
			nsDeps[i] = *(dep.(*appsv1.Deployment))
		}
		return nsDeps, nil
	}
	return []appsv1.Deployment{}, nil
}

func (c *controllerImpl) GetEndpoints(namespace, name string) (*v1.Endpoints, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return nil, err
	}
	indexer := c.controllers["Endpoints"].GetIndexer()
	endpoints, exist, err := indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if exist {
		endpoint, ok := endpoints.(*v1.Endpoints)
		if !ok {
			return nil, errors.New("Bad Endpoints type found in cache")
		}
		return endpoint, nil
	}
	return nil, NewNotFound(name, "core/v1", "Endpoints")
}

func (c *controllerImpl) GetJobs(namespace string) ([]batch_v1.Job, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []batch_v1.Job{}, err
	}
	indexer := c.controllers["Job"].GetIndexer()
	jobs, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []batch_v1.Job{}, err
	}
	if len(jobs) > 0 {
		_, ok := jobs[0].(*batch_v1.Job)
		if !ok {
			return []batch_v1.Job{}, errors.New("Bad Job type found in cache")
		}
		nsJobs := make([]batch_v1.Job, len(jobs))
		for i, job := range jobs {
			nsJobs[i] = *(job.(*batch_v1.Job))
		}
		return nsJobs, nil
	}
	return []batch_v1.Job{}, nil
}

func (c *controllerImpl) GetPods(namespace string) ([]v1.Pod, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []v1.Pod{}, err
	}
	indexer := c.controllers["Pod"].GetIndexer()
	pods, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []v1.Pod{}, err
	}
	if len(pods) > 0 {
		_, ok := pods[0].(*v1.Pod)
		if !ok {
			return []v1.Pod{}, errors.New("Bad Pod type found in cache")
		}
		nsPods := make([]v1.Pod, len(pods))
		for i, pod := range pods {
			nsPods[i] = *(pod.(*v1.Pod))
		}
		return nsPods, nil
	}
	return []v1.Pod{}, nil
}

func (c *controllerImpl) GetReplicationControllers(namespace string) ([]v1.ReplicationController, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []v1.ReplicationController{}, err
	}
	indexer := c.controllers["ReplicationController"].GetIndexer()
	repcons, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []v1.ReplicationController{}, err
	}
	if len(repcons) > 0 {
		_, ok := repcons[0].(*v1.ReplicationController)
		if !ok {
			return []v1.ReplicationController{}, errors.New("Bad ReplicationController type found in cache")
		}
		nsRepcons := make([]v1.ReplicationController, len(repcons))
		for i, repcon := range repcons {
			nsRepcons[i] = *(repcon.(*v1.ReplicationController))
		}
		return nsRepcons, nil
	}
	return []v1.ReplicationController{}, nil
}

func (c *controllerImpl) GetReplicaSets(namespace string) ([]appsv1.ReplicaSet, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []appsv1.ReplicaSet{}, err
	}
	indexer := c.controllers["ReplicaSet"].GetIndexer()
	repsets, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []appsv1.ReplicaSet{}, err
	}
	if len(repsets) > 0 {
		_, ok := repsets[0].(*appsv1.ReplicaSet)
		if !ok {
			return []appsv1.ReplicaSet{}, errors.New("Bad ReplicaSet type found in cache")
		}
		nsRepsets := make([]appsv1.ReplicaSet, len(repsets))
		for i, repset := range repsets {
			nsRepsets[i] = *(repset.(*appsv1.ReplicaSet))
		}
		return nsRepsets, nil
	}
	return []appsv1.ReplicaSet{}, nil
}

func (c *controllerImpl) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return nil, err
	}
	indexer := c.controllers["StatefulSet"].GetIndexer()
	fulsets, exist, err := indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if exist {
		fulset, ok := fulsets.(*appsv1.StatefulSet)
		if !ok {
			return nil, errors.New("Bad StatefulSet type found in cache")
		}
		return fulset, nil
	}
	return nil, NewNotFound(name, "apps/v1", "StatefulSet")
}

func (c *controllerImpl) GetStatefulSets(namespace string) ([]appsv1.StatefulSet, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []appsv1.StatefulSet{}, err
	}
	indexer := c.controllers["StatefulSet"].GetIndexer()
	fulsets, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []appsv1.StatefulSet{}, err
	}
	if len(fulsets) > 0 {
		_, ok := fulsets[0].(*appsv1.StatefulSet)
		if !ok {
			return []appsv1.StatefulSet{}, errors.New("Bad StatefulSet type found in cache")
		}
		nsFulsets := make([]appsv1.StatefulSet, len(fulsets))
		for i, fulset := range fulsets {
			nsFulsets[i] = *(fulset.(*appsv1.StatefulSet))
		}
		return nsFulsets, nil
	}
	return []appsv1.StatefulSet{}, nil
}

func (c *controllerImpl) GetService(namespace, name string) (*v1.Service, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return nil, err
	}
	indexer := c.controllers["Service"].GetIndexer()
	services, exist, err := indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if exist {
		service, ok := services.(*v1.Service)
		if !ok {
			return nil, errors.New("Bad Service type found in cache")
		}
		return service, nil
	}
	return nil, NewNotFound(name, "core/v1", "Service")
}

func (c *controllerImpl) GetServices(namespace string) ([]v1.Service, error) {
	if err := c.checkStateAndRetry(); err != nil {
		return []v1.Service{}, err
	}
	indexer := c.controllers["Service"].GetIndexer()
	services, err := indexer.ByIndex("namespace", namespace)
	if err != nil {
		return []v1.Service{}, err
	}
	if len(services) > 0 {
		_, ok := services[0].(*v1.Service)
		if !ok {
			return []v1.Service{}, errors.New("Bad Service type found in cache")
		}
		nsServices := make([]v1.Service, len(services))
		for i, service := range services {
			nsServices[i] = *(service.(*v1.Service))
		}
		return nsServices, nil
	}
	return []v1.Service{}, nil
}
