/*
Copyright 2017 The Kubernetes Authors.

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

package options

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	auditv1alpha1 "k8s.io/apiserver/pkg/apis/audit/v1alpha1"
	auditv1beta1 "k8s.io/apiserver/pkg/apis/audit/v1beta1"
	"k8s.io/apiserver/pkg/audit"
	"k8s.io/apiserver/pkg/audit/policy"
	"k8s.io/apiserver/pkg/features"
	"k8s.io/apiserver/pkg/server"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	pluginbuffered "k8s.io/apiserver/plugin/pkg/audit/buffered"
	plugindynamic "k8s.io/apiserver/plugin/pkg/audit/dynamic"
	pluginenforced "k8s.io/apiserver/plugin/pkg/audit/dynamic/enforced"
	pluginlog "k8s.io/apiserver/plugin/pkg/audit/log"
	plugintruncate "k8s.io/apiserver/plugin/pkg/audit/truncate"
	pluginwebhook "k8s.io/apiserver/plugin/pkg/audit/webhook"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
)

const (
	// Default configuration values for ModeBatch.
	defaultBatchBufferSize = 10000 // Buffer up to 10000 events before starting discarding.
	// These batch parameters are only used by the webhook backend.
	defaultBatchMaxSize       = 400              // Only send up to 400 events at a time.
	defaultBatchMaxWait       = 30 * time.Second // Send events at least twice a minute.
	defaultBatchThrottleQPS   = 10               // Limit the send rate by 10 QPS.
	defaultBatchThrottleBurst = 15               // Allow up to 15 QPS burst.
)

func appendBackend(existing, newBackend audit.Backend) audit.Backend {
	if existing == nil {
		return newBackend
	}
	if newBackend == nil {
		return existing
	}
	return audit.Union(existing, newBackend)
}

type AuditOptions struct {
	// Policy configuration file for filtering audit events that are captured.
	// If unspecified, a default is provided.
	PolicyFile string

	// Plugin options
	LogOptions     AuditLogOptions
	WebhookOptions AuditWebhookOptions
	DynamicOptions AuditDynamicOptions
}

const (
	// ModeBatch indicates that the audit backend should buffer audit events
	// internally, sending batch updates either once a certain number of
	// events have been received or a certain amount of time has passed.
	ModeBatch = "batch"
	// ModeBlocking causes the audit backend to block on every attempt to process
	// a set of events. This causes requests to the API server to wait for the
	// flush before sending a response.
	ModeBlocking = "blocking"
	// ModeBlockingStrict is the same as ModeBlocking, except when there is
	// a failure during audit logging at RequestReceived stage, the whole
	// request to apiserver will fail.
	ModeBlockingStrict = "blocking-strict"
)

// AllowedModes is the modes known for audit backends.
var AllowedModes = []string{
	ModeBatch,
	ModeBlocking,
	ModeBlockingStrict,
}

type AuditBatchOptions struct {
	// Should the backend asynchronous batch events to the webhook backend or
	// should the backend block responses?
	//
	// Defaults to asynchronous batch events.
	Mode string
	// Configuration for batching backend. Only used in batch mode.
	BatchConfig pluginbuffered.BatchConfig
}

type AuditTruncateOptions struct {
	// Whether truncating is enabled or not.
	Enabled bool

	// Truncating configuration.
	TruncateConfig plugintruncate.Config
}

// AuditLogOptions determines the output of the structured audit log by default.
type AuditLogOptions struct {
	Path       string
	MaxAge     int
	MaxBackups int
	MaxSize    int
	Format     string

	BatchOptions    AuditBatchOptions
	TruncateOptions AuditTruncateOptions

	// API group version used for serializing audit events.
	GroupVersionString string
}

// AuditWebhookOptions control the webhook configuration for audit events.
type AuditWebhookOptions struct {
	ConfigFile     string
	InitialBackoff time.Duration

	BatchOptions    AuditBatchOptions
	TruncateOptions AuditTruncateOptions

	// API group version used for serializing audit events.
	GroupVersionString string
}

// AuditDynamicOptions control the configuration of dynamic backends for audit events
type AuditDynamicOptions struct {
	// Enabled tells whether the dynamic audit capability is enabled.
	Enabled bool

	// Configuration for batching backend. This is currently only used as an override
	// for integration tests
	BatchConfig *pluginbuffered.BatchConfig
}

func NewAuditOptions() *AuditOptions {
	return &AuditOptions{
		WebhookOptions: AuditWebhookOptions{
			InitialBackoff: pluginwebhook.DefaultInitialBackoff,
			BatchOptions: AuditBatchOptions{
				Mode:        ModeBatch,
				BatchConfig: defaultWebhookBatchConfig(),
			},
			TruncateOptions:    NewAuditTruncateOptions(),
			GroupVersionString: "audit.k8s.io/v1",
		},
		LogOptions: AuditLogOptions{
			Format: pluginlog.FormatJson,
			BatchOptions: AuditBatchOptions{
				Mode:        ModeBlocking,
				BatchConfig: defaultLogBatchConfig(),
			},
			TruncateOptions:    NewAuditTruncateOptions(),
			GroupVersionString: "audit.k8s.io/v1",
		},
		DynamicOptions: AuditDynamicOptions{
			Enabled:     false,
			BatchConfig: plugindynamic.NewDefaultWebhookBatchConfig(),
		},
	}
}

func NewAuditTruncateOptions() AuditTruncateOptions {
	return AuditTruncateOptions{
		Enabled: false,
		TruncateConfig: plugintruncate.Config{
			MaxBatchSize: 10 * 1024 * 1024, // 10MB
			MaxEventSize: 100 * 1024,       // 100KB
		},
	}
}

// Validate checks invalid config combination
func (o *AuditOptions) Validate() []error {
	if o == nil {
		return nil
	}

	var allErrors []error
	allErrors = append(allErrors, o.LogOptions.Validate()...)
	allErrors = append(allErrors, o.WebhookOptions.Validate()...)
	allErrors = append(allErrors, o.DynamicOptions.Validate()...)

	return allErrors
}

func validateBackendMode(pluginName string, mode string) error {
	for _, m := range AllowedModes {
		if m == mode {
			return nil
		}
	}
	return fmt.Errorf("invalid audit %s mode %s, allowed modes are %q", pluginName, mode, strings.Join(AllowedModes, ","))
}

func validateBackendBatchOptions(pluginName string, options AuditBatchOptions) error {
	if err := validateBackendMode(pluginName, options.Mode); err != nil {
		return err
	}
	if options.Mode != ModeBatch {
		// Don't validate the unused options.
		return nil
	}
	config := options.BatchConfig
	if config.BufferSize <= 0 {
		return fmt.Errorf("invalid audit batch %s buffer size %v, must be a positive number", pluginName, config.BufferSize)
	}
	if config.MaxBatchSize <= 0 {
		return fmt.Errorf("invalid audit batch %s max batch size %v, must be a positive number", pluginName, config.MaxBatchSize)
	}
	if config.ThrottleEnable {
		if config.ThrottleQPS <= 0 {
			return fmt.Errorf("invalid audit batch %s throttle QPS %v, must be a positive number", pluginName, config.ThrottleQPS)
		}
		if config.ThrottleBurst <= 0 {
			return fmt.Errorf("invalid audit batch %s throttle burst %v, must be a positive number", pluginName, config.ThrottleBurst)
		}
	}
	return nil
}

var knownGroupVersions = []schema.GroupVersion{
	auditv1alpha1.SchemeGroupVersion,
	auditv1beta1.SchemeGroupVersion,
	auditv1.SchemeGroupVersion,
}

func validateGroupVersionString(groupVersion string) error {
	gv, err := schema.ParseGroupVersion(groupVersion)
	if err != nil {
		return err
	}
	if !knownGroupVersion(gv) {
		return fmt.Errorf("invalid group version, allowed versions are %q", knownGroupVersions)
	}
	return nil
}

func knownGroupVersion(gv schema.GroupVersion) bool {
	for _, knownGv := range knownGroupVersions {
		if gv == knownGv {
			return true
		}
	}
	return false
}

func (o *AuditOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.PolicyFile, "audit-policy-file", o.PolicyFile,
		"Path to the file that defines the audit policy configuration.")

	o.LogOptions.AddFlags(fs)
	o.LogOptions.BatchOptions.AddFlags(pluginlog.PluginName, fs)
	o.LogOptions.TruncateOptions.AddFlags(pluginlog.PluginName, fs)
	o.WebhookOptions.AddFlags(fs)
	o.WebhookOptions.BatchOptions.AddFlags(pluginwebhook.PluginName, fs)
	o.WebhookOptions.TruncateOptions.AddFlags(pluginwebhook.PluginName, fs)
	o.DynamicOptions.AddFlags(fs)
}

func (o *AuditOptions) ApplyTo(
	c *server.Config,
	kubeClientConfig *restclient.Config,
	informers informers.SharedInformerFactory,
	processInfo *ProcessInfo,
	webhookOptions *WebhookOptions,
) error {
	if o == nil {
		return nil
	}
	if c == nil {
		return fmt.Errorf("server config must be non-nil")
	}

	// 1. Build policy checker
	checker, err := o.newPolicyChecker()
	if err != nil {
		return err
	}

	// 2. Build log backend
	var logBackend audit.Backend
	if w := o.LogOptions.getWriter(); w != nil {
		if checker == nil {
			klog.V(2).Info("No audit policy file provided, no events will be recorded for log backend")
		} else {
			logBackend = o.LogOptions.newBackend(w)
		}
	}

	// 3. Build webhook backend
	var webhookBackend audit.Backend
	if o.WebhookOptions.enabled() {
		if checker == nil {
			klog.V(2).Info("No audit policy file provided, no events will be recorded for webhook backend")
		} else {
			webhookBackend, err = o.WebhookOptions.newUntruncatedBackend()
			if err != nil {
				return err
			}
		}
	}

	groupVersion, err := schema.ParseGroupVersion(o.WebhookOptions.GroupVersionString)
	if err != nil {
		return err
	}

	// 4. Apply dynamic options.
	var dynamicBackend audit.Backend
	if o.DynamicOptions.enabled() {
		// if dynamic is enabled the webhook and log backends need to be wrapped in an enforced backend with the static policy
		if webhookBackend != nil {
			webhookBackend = pluginenforced.NewBackend(webhookBackend, checker)
		}
		if logBackend != nil {
			logBackend = pluginenforced.NewBackend(logBackend, checker)
		}
		// build dynamic backend
		dynamicBackend, checker, err = o.DynamicOptions.newBackend(c.ExternalAddress, kubeClientConfig, informers, processInfo, webhookOptions)
		if err != nil {
			return err
		}
		// union dynamic and webhook backends so that truncate options can be applied to both
		dynamicBackend = appendBackend(webhookBackend, dynamicBackend)
		dynamicBackend = o.WebhookOptions.TruncateOptions.wrapBackend(dynamicBackend, groupVersion)
	} else if webhookBackend != nil {
		// if only webhook is enabled wrap it in the truncate options
		dynamicBackend = o.WebhookOptions.TruncateOptions.wrapBackend(webhookBackend, groupVersion)
	}

	// 5. Set the policy checker
	c.AuditPolicyChecker = checker

	// 6. Join the log backend with the webhooks
	c.AuditBackend = appendBackend(logBackend, dynamicBackend)

	if c.AuditBackend != nil {
		klog.V(2).Infof("Using audit backend: %s", c.AuditBackend)
	}
	return nil
}

func (o *AuditOptions) newPolicyChecker() (policy.Checker, error) {
	if o.PolicyFile == "" {
		return nil, nil
	}

	p, err := policy.LoadPolicyFromFile(o.PolicyFile)
	if err != nil {
		return nil, fmt.Errorf("loading audit policy file: %v", err)
	}
	return policy.NewChecker(p), nil
}

func (o *AuditBatchOptions) AddFlags(pluginName string, fs *pflag.FlagSet) {
	fs.StringVar(&o.Mode, fmt.Sprintf("audit-%s-mode", pluginName), o.Mode,
		"Strategy for sending audit events. Blocking indicates sending events should block"+
			" server responses. Batch causes the backend to buffer and write events"+
			" asynchronously. Known modes are "+strings.Join(AllowedModes, ",")+".")
	fs.IntVar(&o.BatchConfig.BufferSize, fmt.Sprintf("audit-%s-batch-buffer-size", pluginName),
		o.BatchConfig.BufferSize, "The size of the buffer to store events before "+
			"batching and writing. Only used in batch mode.")
	fs.IntVar(&o.BatchConfig.MaxBatchSize, fmt.Sprintf("audit-%s-batch-max-size", pluginName),
		o.BatchConfig.MaxBatchSize, "The maximum size of a batch. Only used in batch mode.")
	fs.DurationVar(&o.BatchConfig.MaxBatchWait, fmt.Sprintf("audit-%s-batch-max-wait", pluginName),
		o.BatchConfig.MaxBatchWait, "The amount of time to wait before force writing the "+
			"batch that hadn't reached the max size. Only used in batch mode.")
	fs.BoolVar(&o.BatchConfig.ThrottleEnable, fmt.Sprintf("audit-%s-batch-throttle-enable", pluginName),
		o.BatchConfig.ThrottleEnable, "Whether batching throttling is enabled. Only used in batch mode.")
	fs.Float32Var(&o.BatchConfig.ThrottleQPS, fmt.Sprintf("audit-%s-batch-throttle-qps", pluginName),
		o.BatchConfig.ThrottleQPS, "Maximum average number of batches per second. "+
			"Only used in batch mode.")
	fs.IntVar(&o.BatchConfig.ThrottleBurst, fmt.Sprintf("audit-%s-batch-throttle-burst", pluginName),
		o.BatchConfig.ThrottleBurst, "Maximum number of requests sent at the same "+
			"moment if ThrottleQPS was not utilized before. Only used in batch mode.")
}

type ignoreErrorsBackend struct {
	audit.Backend
}

func (i *ignoreErrorsBackend) ProcessEvents(ev ...*auditinternal.Event) bool {
	i.Backend.ProcessEvents(ev...)
	return true
}

func (i *ignoreErrorsBackend) String() string {
	return fmt.Sprintf("ignoreErrors<%s>", i.Backend)
}

func (o *AuditBatchOptions) wrapBackend(delegate audit.Backend) audit.Backend {
	if o.Mode == ModeBlockingStrict {
		return delegate
	}
	if o.Mode == ModeBlocking {
		return &ignoreErrorsBackend{Backend: delegate}
	}
	return pluginbuffered.NewBackend(delegate, o.BatchConfig)
}

func (o *AuditTruncateOptions) Validate(pluginName string) error {
	config := o.TruncateConfig
	if config.MaxEventSize <= 0 {
		return fmt.Errorf("invalid audit truncate %s max event size %v, must be a positive number", pluginName, config.MaxEventSize)
	}
	if config.MaxBatchSize < config.MaxEventSize {
		return fmt.Errorf("invalid audit truncate %s max batch size %v, must be greater than "+
			"max event size (%v)", pluginName, config.MaxBatchSize, config.MaxEventSize)
	}
	return nil
}

func (o *AuditTruncateOptions) AddFlags(pluginName string, fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, fmt.Sprintf("audit-%s-truncate-enabled", pluginName),
		o.Enabled, "Whether event and batch truncating is enabled.")
	fs.Int64Var(&o.TruncateConfig.MaxBatchSize, fmt.Sprintf("audit-%s-truncate-max-batch-size", pluginName),
		o.TruncateConfig.MaxBatchSize, "Maximum size of the batch sent to the underlying backend. "+
			"Actual serialized size can be several hundreds of bytes greater. If a batch exceeds this limit, "+
			"it is split into several batches of smaller size.")
	fs.Int64Var(&o.TruncateConfig.MaxEventSize, fmt.Sprintf("audit-%s-truncate-max-event-size", pluginName),
		o.TruncateConfig.MaxEventSize, "Maximum size of the audit event sent to the underlying backend. "+
			"If the size of an event is greater than this number, first request and response are removed, and "+
			"if this doesn't reduce the size enough, event is discarded.")
}

func (o *AuditTruncateOptions) wrapBackend(delegate audit.Backend, gv schema.GroupVersion) audit.Backend {
	if !o.Enabled {
		return delegate
	}
	return plugintruncate.NewBackend(delegate, o.TruncateConfig, gv)
}

func (o *AuditLogOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Path, "audit-log-path", o.Path,
		"If set, all requests coming to the apiserver will be logged to this file.  '-' means standard out.")
	fs.IntVar(&o.MaxAge, "audit-log-maxage", o.MaxAge,
		"The maximum number of days to retain old audit log files based on the timestamp encoded in their filename.")
	fs.IntVar(&o.MaxBackups, "audit-log-maxbackup", o.MaxBackups,
		"The maximum number of old audit log files to retain.")
	fs.IntVar(&o.MaxSize, "audit-log-maxsize", o.MaxSize,
		"The maximum size in megabytes of the audit log file before it gets rotated.")
	fs.StringVar(&o.Format, "audit-log-format", o.Format,
		"Format of saved audits. \"legacy\" indicates 1-line text format for each event."+
			" \"json\" indicates structured json format. Known formats are "+
			strings.Join(pluginlog.AllowedFormats, ",")+".")
	fs.StringVar(&o.GroupVersionString, "audit-log-version", o.GroupVersionString,
		"API group and version used for serializing audit events written to log.")
}

func (o *AuditLogOptions) Validate() []error {
	// Check whether the log backend is enabled based on the options.
	if !o.enabled() {
		return nil
	}

	var allErrors []error

	if err := validateBackendBatchOptions(pluginlog.PluginName, o.BatchOptions); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := o.TruncateOptions.Validate(pluginlog.PluginName); err != nil {
		allErrors = append(allErrors, err)
	}

	if err := validateGroupVersionString(o.GroupVersionString); err != nil {
		allErrors = append(allErrors, err)
	}

	// Check log format
	validFormat := false
	for _, f := range pluginlog.AllowedFormats {
		if f == o.Format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		allErrors = append(allErrors, fmt.Errorf("invalid audit log format %s, allowed formats are %q", o.Format, strings.Join(pluginlog.AllowedFormats, ",")))
	}

	// Check validities of MaxAge, MaxBackups and MaxSize of log options, if file log backend is enabled.
	if o.MaxAge < 0 {
		allErrors = append(allErrors, fmt.Errorf("--audit-log-maxage %v can't be a negative number", o.MaxAge))
	}
	if o.MaxBackups < 0 {
		allErrors = append(allErrors, fmt.Errorf("--audit-log-maxbackup %v can't be a negative number", o.MaxBackups))
	}
	if o.MaxSize < 0 {
		allErrors = append(allErrors, fmt.Errorf("--audit-log-maxsize %v can't be a negative number", o.MaxSize))
	}

	return allErrors
}

// Check whether the log backend is enabled based on the options.
func (o *AuditLogOptions) enabled() bool {
	return o != nil && o.Path != ""
}

func (o *AuditLogOptions) getWriter() io.Writer {
	if !o.enabled() {
		return nil
	}

	var w io.Writer = os.Stdout
	if o.Path != "-" {
		w = &lumberjack.Logger{
			Filename:   o.Path,
			MaxAge:     o.MaxAge,
			MaxBackups: o.MaxBackups,
			MaxSize:    o.MaxSize,
		}
	}
	return w
}

func (o *AuditLogOptions) newBackend(w io.Writer) audit.Backend {
	groupVersion, _ := schema.ParseGroupVersion(o.GroupVersionString)
	log := pluginlog.NewBackend(w, o.Format, groupVersion)
	log = o.BatchOptions.wrapBackend(log)
	log = o.TruncateOptions.wrapBackend(log, groupVersion)
	return log
}

func (o *AuditWebhookOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigFile, "audit-webhook-config-file", o.ConfigFile,
		"Path to a kubeconfig formatted file that defines the audit webhook configuration.")
	fs.DurationVar(&o.InitialBackoff, "audit-webhook-initial-backoff",
		o.InitialBackoff, "The amount of time to wait before retrying the first failed request.")
	fs.DurationVar(&o.InitialBackoff, "audit-webhook-batch-initial-backoff",
		o.InitialBackoff, "The amount of time to wait before retrying the first failed request.")
	fs.MarkDeprecated("audit-webhook-batch-initial-backoff",
		"Deprecated, use --audit-webhook-initial-backoff instead.")
	fs.StringVar(&o.GroupVersionString, "audit-webhook-version", o.GroupVersionString,
		"API group and version used for serializing audit events written to webhook.")
}

func (o *AuditWebhookOptions) Validate() []error {
	if !o.enabled() {
		return nil
	}

	var allErrors []error
	if err := validateBackendBatchOptions(pluginwebhook.PluginName, o.BatchOptions); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := o.TruncateOptions.Validate(pluginwebhook.PluginName); err != nil {
		allErrors = append(allErrors, err)
	}

	if err := validateGroupVersionString(o.GroupVersionString); err != nil {
		allErrors = append(allErrors, err)
	}
	return allErrors
}

func (o *AuditWebhookOptions) enabled() bool {
	return o != nil && o.ConfigFile != ""
}

// newUntruncatedBackend returns a webhook backend without the truncate options applied
// this is done so that the same trucate backend can wrap both the webhook and dynamic backends
func (o *AuditWebhookOptions) newUntruncatedBackend() (audit.Backend, error) {
	groupVersion, _ := schema.ParseGroupVersion(o.GroupVersionString)
	webhook, err := pluginwebhook.NewBackend(o.ConfigFile, groupVersion, o.InitialBackoff)
	if err != nil {
		return nil, fmt.Errorf("initializing audit webhook: %v", err)
	}
	webhook = o.BatchOptions.wrapBackend(webhook)
	return webhook, nil
}

func (o *AuditDynamicOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "audit-dynamic-configuration", o.Enabled,
		"Enables dynamic audit configuration. This feature also requires the DynamicAuditing feature flag")
}

func (o *AuditDynamicOptions) enabled() bool {
	return o.Enabled && utilfeature.DefaultFeatureGate.Enabled(features.DynamicAuditing)
}

func (o *AuditDynamicOptions) Validate() []error {
	var allErrors []error
	if o.Enabled && !utilfeature.DefaultFeatureGate.Enabled(features.DynamicAuditing) {
		allErrors = append(allErrors, fmt.Errorf("--audit-dynamic-configuration set, but DynamicAuditing feature gate is not enabled"))
	}
	return allErrors
}

func (o *AuditDynamicOptions) newBackend(
	hostname string,
	kubeClientConfig *restclient.Config,
	informers informers.SharedInformerFactory,
	processInfo *ProcessInfo,
	webhookOptions *WebhookOptions,
) (audit.Backend, policy.Checker, error) {
	if err := validateProcessInfo(processInfo); err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, nil, err
	}
	if webhookOptions == nil {
		webhookOptions = NewWebhookOptions()
	}
	checker := policy.NewDynamicChecker()
	informer := informers.Auditregistration().V1alpha1().AuditSinks()
	eventSink := &v1core.EventSinkImpl{Interface: clientset.CoreV1().Events(processInfo.Namespace)}

	dc := &plugindynamic.Config{
		Informer:       informer,
		BufferedConfig: o.BatchConfig,
		EventConfig: plugindynamic.EventConfig{
			Sink: eventSink,
			Source: corev1.EventSource{
				Component: processInfo.Name,
				Host:      hostname,
			},
		},
		WebhookConfig: plugindynamic.WebhookConfig{
			AuthInfoResolverWrapper: webhookOptions.AuthInfoResolverWrapper,
			ServiceResolver:         webhookOptions.ServiceResolver,
		},
	}
	backend, err := plugindynamic.NewBackend(dc)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create dynamic audit backend: %v", err)
	}
	return backend, checker, nil
}

// defaultWebhookBatchConfig returns the default BatchConfig used by the Webhook backend.
func defaultWebhookBatchConfig() pluginbuffered.BatchConfig {
	return pluginbuffered.BatchConfig{
		BufferSize:   defaultBatchBufferSize,
		MaxBatchSize: defaultBatchMaxSize,
		MaxBatchWait: defaultBatchMaxWait,

		ThrottleEnable: true,
		ThrottleQPS:    defaultBatchThrottleQPS,
		ThrottleBurst:  defaultBatchThrottleBurst,

		AsyncDelegate: true,
	}
}

// defaultLogBatchConfig returns the default BatchConfig used by the Log backend.
func defaultLogBatchConfig() pluginbuffered.BatchConfig {
	return pluginbuffered.BatchConfig{
		BufferSize: defaultBatchBufferSize,
		// Batching is not useful for the log-file backend.
		// MaxBatchWait ignored.
		MaxBatchSize:   1,
		ThrottleEnable: false,
		// Asynchronous log threads just create lock contention.
		AsyncDelegate: false,
	}
}
