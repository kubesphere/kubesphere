// Package monitoring provides access to the Stackdriver Monitoring API.
//
// See https://cloud.google.com/monitoring/api/
//
// Usage example:
//
//   import "google.golang.org/api/monitoring/v3"
//   ...
//   monitoringService, err := monitoring.New(oauthHttpClient)
package monitoring

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	context "golang.org/x/net/context"
	ctxhttp "golang.org/x/net/context/ctxhttp"
	gensupport "google.golang.org/api/gensupport"
	googleapi "google.golang.org/api/googleapi"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = gensupport.MarshalJSON
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "monitoring:v3"
const apiName = "monitoring"
const apiVersion = "v3"
const basePath = "https://monitoring.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

	// View and write monitoring data for all of your Google and third-party
	// Cloud and API projects
	MonitoringScope = "https://www.googleapis.com/auth/monitoring"

	// View monitoring data for all of your Google Cloud and third-party
	// projects
	MonitoringReadScope = "https://www.googleapis.com/auth/monitoring.read"

	// Publish metric data to your Google Cloud projects
	MonitoringWriteScope = "https://www.googleapis.com/auth/monitoring.write"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Projects = NewProjectsService(s)
	s.UptimeCheckIps = NewUptimeCheckIpsService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Projects *ProjectsService

	UptimeCheckIps *UptimeCheckIpsService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewProjectsService(s *Service) *ProjectsService {
	rs := &ProjectsService{s: s}
	rs.CollectdTimeSeries = NewProjectsCollectdTimeSeriesService(s)
	rs.Groups = NewProjectsGroupsService(s)
	rs.MetricDescriptors = NewProjectsMetricDescriptorsService(s)
	rs.MonitoredResourceDescriptors = NewProjectsMonitoredResourceDescriptorsService(s)
	rs.TimeSeries = NewProjectsTimeSeriesService(s)
	rs.UptimeCheckConfigs = NewProjectsUptimeCheckConfigsService(s)
	return rs
}

type ProjectsService struct {
	s *Service

	CollectdTimeSeries *ProjectsCollectdTimeSeriesService

	Groups *ProjectsGroupsService

	MetricDescriptors *ProjectsMetricDescriptorsService

	MonitoredResourceDescriptors *ProjectsMonitoredResourceDescriptorsService

	TimeSeries *ProjectsTimeSeriesService

	UptimeCheckConfigs *ProjectsUptimeCheckConfigsService
}

func NewProjectsCollectdTimeSeriesService(s *Service) *ProjectsCollectdTimeSeriesService {
	rs := &ProjectsCollectdTimeSeriesService{s: s}
	return rs
}

type ProjectsCollectdTimeSeriesService struct {
	s *Service
}

func NewProjectsGroupsService(s *Service) *ProjectsGroupsService {
	rs := &ProjectsGroupsService{s: s}
	rs.Members = NewProjectsGroupsMembersService(s)
	return rs
}

type ProjectsGroupsService struct {
	s *Service

	Members *ProjectsGroupsMembersService
}

func NewProjectsGroupsMembersService(s *Service) *ProjectsGroupsMembersService {
	rs := &ProjectsGroupsMembersService{s: s}
	return rs
}

type ProjectsGroupsMembersService struct {
	s *Service
}

func NewProjectsMetricDescriptorsService(s *Service) *ProjectsMetricDescriptorsService {
	rs := &ProjectsMetricDescriptorsService{s: s}
	return rs
}

type ProjectsMetricDescriptorsService struct {
	s *Service
}

func NewProjectsMonitoredResourceDescriptorsService(s *Service) *ProjectsMonitoredResourceDescriptorsService {
	rs := &ProjectsMonitoredResourceDescriptorsService{s: s}
	return rs
}

type ProjectsMonitoredResourceDescriptorsService struct {
	s *Service
}

func NewProjectsTimeSeriesService(s *Service) *ProjectsTimeSeriesService {
	rs := &ProjectsTimeSeriesService{s: s}
	return rs
}

type ProjectsTimeSeriesService struct {
	s *Service
}

func NewProjectsUptimeCheckConfigsService(s *Service) *ProjectsUptimeCheckConfigsService {
	rs := &ProjectsUptimeCheckConfigsService{s: s}
	return rs
}

type ProjectsUptimeCheckConfigsService struct {
	s *Service
}

func NewUptimeCheckIpsService(s *Service) *UptimeCheckIpsService {
	rs := &UptimeCheckIpsService{s: s}
	return rs
}

type UptimeCheckIpsService struct {
	s *Service
}

// BasicAuthentication: A type of authentication to perform against the
// specified resource or URL that uses username and password. Currently,
// only Basic authentication is supported in Uptime Monitoring.
type BasicAuthentication struct {
	// Password: The password to authenticate.
	Password string `json:"password,omitempty"`

	// Username: The username to authenticate.
	Username string `json:"username,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Password") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Password") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *BasicAuthentication) MarshalJSON() ([]byte, error) {
	type NoMethod BasicAuthentication
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// BucketOptions: BucketOptions describes the bucket boundaries used to
// create a histogram for the distribution. The buckets can be in a
// linear sequence, an exponential sequence, or each bucket can be
// specified explicitly. BucketOptions does not include the number of
// values in each bucket.A bucket has an inclusive lower bound and
// exclusive upper bound for the values that are counted for that
// bucket. The upper bound of a bucket must be strictly greater than the
// lower bound. The sequence of N buckets for a distribution consists of
// an underflow bucket (number 0), zero or more finite buckets (number 1
// through N - 2) and an overflow bucket (number N - 1). The buckets are
// contiguous: the lower bound of bucket i (i > 0) is the same as the
// upper bound of bucket i - 1. The buckets span the whole range of
// finite values: lower bound of the underflow bucket is -infinity and
// the upper bound of the overflow bucket is +infinity. The finite
// buckets are so-called because both bounds are finite.
type BucketOptions struct {
	// ExplicitBuckets: The explicit buckets.
	ExplicitBuckets *Explicit `json:"explicitBuckets,omitempty"`

	// ExponentialBuckets: The exponential buckets.
	ExponentialBuckets *Exponential `json:"exponentialBuckets,omitempty"`

	// LinearBuckets: The linear bucket.
	LinearBuckets *Linear `json:"linearBuckets,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ExplicitBuckets") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ExplicitBuckets") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *BucketOptions) MarshalJSON() ([]byte, error) {
	type NoMethod BucketOptions
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CollectdPayload: A collection of data points sent from a
// collectd-based plugin. See the collectd documentation for more
// information.
type CollectdPayload struct {
	// EndTime: The end time of the interval.
	EndTime string `json:"endTime,omitempty"`

	// Metadata: The measurement metadata. Example: "process_id" -> 12345
	Metadata map[string]TypedValue `json:"metadata,omitempty"`

	// Plugin: The name of the plugin. Example: "disk".
	Plugin string `json:"plugin,omitempty"`

	// PluginInstance: The instance name of the plugin Example: "hdcl".
	PluginInstance string `json:"pluginInstance,omitempty"`

	// StartTime: The start time of the interval.
	StartTime string `json:"startTime,omitempty"`

	// Type: The measurement type. Example: "memory".
	Type string `json:"type,omitempty"`

	// TypeInstance: The measurement type instance. Example: "used".
	TypeInstance string `json:"typeInstance,omitempty"`

	// Values: The measured values during this time interval. Each value
	// must have a different dataSourceName.
	Values []*CollectdValue `json:"values,omitempty"`

	// ForceSendFields is a list of field names (e.g. "EndTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "EndTime") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CollectdPayload) MarshalJSON() ([]byte, error) {
	type NoMethod CollectdPayload
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CollectdPayloadError: Describes the error status for payloads that
// were not written.
type CollectdPayloadError struct {
	// Error: Records the error status for the payload. If this field is
	// present, the partial errors for nested values won't be populated.
	Error *Status `json:"error,omitempty"`

	// Index: The zero-based index in
	// CreateCollectdTimeSeriesRequest.collectd_payloads.
	Index int64 `json:"index,omitempty"`

	// ValueErrors: Records the error status for values that were not
	// written due to an error.Failed payloads for which nothing is written
	// will not include partial value errors.
	ValueErrors []*CollectdValueError `json:"valueErrors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Error") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Error") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CollectdPayloadError) MarshalJSON() ([]byte, error) {
	type NoMethod CollectdPayloadError
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CollectdValue: A single data point from a collectd-based plugin.
type CollectdValue struct {
	// DataSourceName: The data source for the collectd value. For example
	// there are two data sources for network measurements: "rx" and "tx".
	DataSourceName string `json:"dataSourceName,omitempty"`

	// DataSourceType: The type of measurement.
	//
	// Possible values:
	//   "UNSPECIFIED_DATA_SOURCE_TYPE" - An unspecified data source type.
	// This corresponds to
	// google.api.MetricDescriptor.MetricKind.METRIC_KIND_UNSPECIFIED.
	//   "GAUGE" - An instantaneous measurement of a varying quantity. This
	// corresponds to google.api.MetricDescriptor.MetricKind.GAUGE.
	//   "COUNTER" - A cumulative value over time. This corresponds to
	// google.api.MetricDescriptor.MetricKind.CUMULATIVE.
	//   "DERIVE" - A rate of change of the measurement.
	//   "ABSOLUTE" - An amount of change since the last measurement
	// interval. This corresponds to
	// google.api.MetricDescriptor.MetricKind.DELTA.
	DataSourceType string `json:"dataSourceType,omitempty"`

	// Value: The measurement value.
	Value *TypedValue `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DataSourceName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DataSourceName") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *CollectdValue) MarshalJSON() ([]byte, error) {
	type NoMethod CollectdValue
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CollectdValueError: Describes the error status for values that were
// not written.
type CollectdValueError struct {
	// Error: Records the error status for the value.
	Error *Status `json:"error,omitempty"`

	// Index: The zero-based index in CollectdPayload.values within the
	// parent CreateCollectdTimeSeriesRequest.collectd_payloads.
	Index int64 `json:"index,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Error") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Error") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CollectdValueError) MarshalJSON() ([]byte, error) {
	type NoMethod CollectdValueError
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ContentMatcher: Used to perform string matching. Currently, this
// matches on the exact content. In the future, it can be expanded to
// allow for regular expressions and more complex matching.
type ContentMatcher struct {
	// Content: String content to match (max 1024 bytes)
	Content string `json:"content,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Content") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Content") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ContentMatcher) MarshalJSON() ([]byte, error) {
	type NoMethod ContentMatcher
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CreateCollectdTimeSeriesRequest: The CreateCollectdTimeSeries
// request.
type CreateCollectdTimeSeriesRequest struct {
	// CollectdPayloads: The collectd payloads representing the time series
	// data. You must not include more than a single point for each time
	// series, so no two payloads can have the same values for all of the
	// fields plugin, plugin_instance, type, and type_instance.
	CollectdPayloads []*CollectdPayload `json:"collectdPayloads,omitempty"`

	// CollectdVersion: The version of collectd that collected the data.
	// Example: "5.3.0-192.el6".
	CollectdVersion string `json:"collectdVersion,omitempty"`

	// Resource: The monitored resource associated with the time series.
	Resource *MonitoredResource `json:"resource,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CollectdPayloads") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CollectdPayloads") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *CreateCollectdTimeSeriesRequest) MarshalJSON() ([]byte, error) {
	type NoMethod CreateCollectdTimeSeriesRequest
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CreateCollectdTimeSeriesResponse: The CreateCollectdTimeSeries
// response.
type CreateCollectdTimeSeriesResponse struct {
	// PayloadErrors: Records the error status for points that were not
	// written due to an error.Failed requests for which nothing is written
	// will return an error response instead.
	PayloadErrors []*CollectdPayloadError `json:"payloadErrors,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "PayloadErrors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "PayloadErrors") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CreateCollectdTimeSeriesResponse) MarshalJSON() ([]byte, error) {
	type NoMethod CreateCollectdTimeSeriesResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CreateTimeSeriesRequest: The CreateTimeSeries request.
type CreateTimeSeriesRequest struct {
	// TimeSeries: The new data to be added to a list of time series. Adds
	// at most one data point to each of several time series. The new data
	// point must be more recent than any other point in its time series.
	// Each TimeSeries value must fully specify a unique time series by
	// supplying all label values for the metric and the monitored resource.
	TimeSeries []*TimeSeries `json:"timeSeries,omitempty"`

	// ForceSendFields is a list of field names (e.g. "TimeSeries") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "TimeSeries") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CreateTimeSeriesRequest) MarshalJSON() ([]byte, error) {
	type NoMethod CreateTimeSeriesRequest
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Distribution: Distribution contains summary statistics for a
// population of values. It optionally contains a histogram representing
// the distribution of those values across a set of buckets.The summary
// statistics are the count, mean, sum of the squared deviation from the
// mean, the minimum, and the maximum of the set of population of
// values. The histogram is based on a sequence of buckets and gives a
// count of values that fall into each bucket. The boundaries of the
// buckets are given either explicitly or by formulas for buckets of
// fixed or exponentially increasing widths.Although it is not
// forbidden, it is generally a bad idea to include non-finite values
// (infinities or NaNs) in the population of values, as this will render
// the mean and sum_of_squared_deviation fields meaningless.
type Distribution struct {
	// BucketCounts: Required in the Stackdriver Monitoring API v3. The
	// values for each bucket specified in bucket_options. The sum of the
	// values in bucketCounts must equal the value in the count field of the
	// Distribution object. The order of the bucket counts follows the
	// numbering schemes described for the three bucket types. The underflow
	// bucket has number 0; the finite buckets, if any, have numbers 1
	// through N-2; and the overflow bucket has number N-1. The size of
	// bucket_counts must not be greater than N. If the size is less than N,
	// then the remaining buckets are assigned values of zero.
	BucketCounts googleapi.Int64s `json:"bucketCounts,omitempty"`

	// BucketOptions: Required in the Stackdriver Monitoring API v3. Defines
	// the histogram bucket boundaries.
	BucketOptions *BucketOptions `json:"bucketOptions,omitempty"`

	// Count: The number of values in the population. Must be non-negative.
	// This value must equal the sum of the values in bucket_counts if a
	// histogram is provided.
	Count int64 `json:"count,omitempty,string"`

	// Mean: The arithmetic mean of the values in the population. If count
	// is zero then this field must be zero.
	Mean float64 `json:"mean,omitempty"`

	// Range: If specified, contains the range of the population values. The
	// field must not be present if the count is zero. This field is
	// presently ignored by the Stackdriver Monitoring API v3.
	Range *Range `json:"range,omitempty"`

	// SumOfSquaredDeviation: The sum of squared deviations from the mean of
	// the values in the population. For values x_i this
	// is:
	// Sum[i=1..n]((x_i - mean)^2)
	// Knuth, "The Art of Computer Programming", Vol. 2, page 323, 3rd
	// edition describes Welford's method for accumulating this sum in one
	// pass.If count is zero then this field must be zero.
	SumOfSquaredDeviation float64 `json:"sumOfSquaredDeviation,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BucketCounts") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BucketCounts") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Distribution) MarshalJSON() ([]byte, error) {
	type NoMethod Distribution
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Distribution) UnmarshalJSON(data []byte) error {
	type NoMethod Distribution
	var s1 struct {
		Mean                  gensupport.JSONFloat64 `json:"mean"`
		SumOfSquaredDeviation gensupport.JSONFloat64 `json:"sumOfSquaredDeviation"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Mean = float64(s1.Mean)
	s.SumOfSquaredDeviation = float64(s1.SumOfSquaredDeviation)
	return nil
}

// Empty: A generic empty message that you can re-use to avoid defining
// duplicated empty messages in your APIs. A typical example is to use
// it as the request or the response type of an API method. For
// instance:
// service Foo {
//   rpc Bar(google.protobuf.Empty) returns
// (google.protobuf.Empty);
// }
// The JSON representation for Empty is empty JSON object {}.
type Empty struct {
	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`
}

// Explicit: Specifies a set of buckets with arbitrary widths.There are
// size(bounds) + 1 (= N) buckets. Bucket i has the following
// boundaries:Upper bound (0 <= i < N-1): boundsi  Lower bound (1 <= i <
// N); boundsi - 1The bounds field must contain at least one element. If
// bounds has only one element, then there are no finite buckets, and
// that single element is the common boundary of the overflow and
// underflow buckets.
type Explicit struct {
	// Bounds: The values must be monotonically increasing.
	Bounds []float64 `json:"bounds,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Bounds") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Bounds") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Explicit) MarshalJSON() ([]byte, error) {
	type NoMethod Explicit
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Exponential: Specifies an exponential sequence of buckets that have a
// width that is proportional to the value of the lower bound. Each
// bucket represents a constant relative uncertainty on a specific value
// in the bucket.There are num_finite_buckets + 2 (= N) buckets. Bucket
// i has the following boundaries:Upper bound (0 <= i < N-1): scale *
// (growth_factor ^ i).  Lower bound (1 <= i < N): scale *
// (growth_factor ^ (i - 1)).
type Exponential struct {
	// GrowthFactor: Must be greater than 1.
	GrowthFactor float64 `json:"growthFactor,omitempty"`

	// NumFiniteBuckets: Must be greater than 0.
	NumFiniteBuckets int64 `json:"numFiniteBuckets,omitempty"`

	// Scale: Must be greater than 0.
	Scale float64 `json:"scale,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GrowthFactor") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "GrowthFactor") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Exponential) MarshalJSON() ([]byte, error) {
	type NoMethod Exponential
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Exponential) UnmarshalJSON(data []byte) error {
	type NoMethod Exponential
	var s1 struct {
		GrowthFactor gensupport.JSONFloat64 `json:"growthFactor"`
		Scale        gensupport.JSONFloat64 `json:"scale"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.GrowthFactor = float64(s1.GrowthFactor)
	s.Scale = float64(s1.Scale)
	return nil
}

// Field: A single field of a message type.
type Field struct {
	// Cardinality: The field cardinality.
	//
	// Possible values:
	//   "CARDINALITY_UNKNOWN" - For fields with unknown cardinality.
	//   "CARDINALITY_OPTIONAL" - For optional fields.
	//   "CARDINALITY_REQUIRED" - For required fields. Proto2 syntax only.
	//   "CARDINALITY_REPEATED" - For repeated fields.
	Cardinality string `json:"cardinality,omitempty"`

	// DefaultValue: The string value of the default value of this field.
	// Proto2 syntax only.
	DefaultValue string `json:"defaultValue,omitempty"`

	// JsonName: The field JSON name.
	JsonName string `json:"jsonName,omitempty"`

	// Kind: The field type.
	//
	// Possible values:
	//   "TYPE_UNKNOWN" - Field type unknown.
	//   "TYPE_DOUBLE" - Field type double.
	//   "TYPE_FLOAT" - Field type float.
	//   "TYPE_INT64" - Field type int64.
	//   "TYPE_UINT64" - Field type uint64.
	//   "TYPE_INT32" - Field type int32.
	//   "TYPE_FIXED64" - Field type fixed64.
	//   "TYPE_FIXED32" - Field type fixed32.
	//   "TYPE_BOOL" - Field type bool.
	//   "TYPE_STRING" - Field type string.
	//   "TYPE_GROUP" - Field type group. Proto2 syntax only, and
	// deprecated.
	//   "TYPE_MESSAGE" - Field type message.
	//   "TYPE_BYTES" - Field type bytes.
	//   "TYPE_UINT32" - Field type uint32.
	//   "TYPE_ENUM" - Field type enum.
	//   "TYPE_SFIXED32" - Field type sfixed32.
	//   "TYPE_SFIXED64" - Field type sfixed64.
	//   "TYPE_SINT32" - Field type sint32.
	//   "TYPE_SINT64" - Field type sint64.
	Kind string `json:"kind,omitempty"`

	// Name: The field name.
	Name string `json:"name,omitempty"`

	// Number: The field number.
	Number int64 `json:"number,omitempty"`

	// OneofIndex: The index of the field type in Type.oneofs, for message
	// or enumeration types. The first type has index 1; zero means the type
	// is not in the list.
	OneofIndex int64 `json:"oneofIndex,omitempty"`

	// Options: The protocol buffer options.
	Options []*Option `json:"options,omitempty"`

	// Packed: Whether to use alternative packed wire representation.
	Packed bool `json:"packed,omitempty"`

	// TypeUrl: The field type URL, without the scheme, for message or
	// enumeration types. Example:
	// "type.googleapis.com/google.protobuf.Timestamp".
	TypeUrl string `json:"typeUrl,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Cardinality") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Cardinality") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Field) MarshalJSON() ([]byte, error) {
	type NoMethod Field
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Group: The description of a dynamic collection of monitored
// resources. Each group has a filter that is matched against monitored
// resources and their associated metadata. If a group's filter matches
// an available monitored resource, then that resource is a member of
// that group. Groups can contain any number of monitored resources, and
// each monitored resource can be a member of any number of
// groups.Groups can be nested in parent-child hierarchies. The
// parentName field identifies an optional parent for each group. If a
// group has a parent, then the only monitored resources available to be
// matched by the group's filter are the resources contained in the
// parent group. In other words, a group contains the monitored
// resources that match its filter and the filters of all the group's
// ancestors. A group without a parent can contain any monitored
// resource.For example, consider an infrastructure running a set of
// instances with two user-defined tags: "environment" and "role". A
// parent group has a filter, environment="production". A child of that
// parent group has a filter, role="transcoder". The parent group
// contains all instances in the production environment, regardless of
// their roles. The child group contains instances that have the
// transcoder role and are in the production environment.The monitored
// resources contained in a group can change at any moment, depending on
// what resources exist and what filters are associated with the group
// and its ancestors.
type Group struct {
	// DisplayName: A user-assigned name for this group, used only for
	// display purposes.
	DisplayName string `json:"displayName,omitempty"`

	// Filter: The filter used to determine which monitored resources belong
	// to this group.
	Filter string `json:"filter,omitempty"`

	// IsCluster: If true, the members of this group are considered to be a
	// cluster. The system can perform additional analysis on groups that
	// are clusters.
	IsCluster bool `json:"isCluster,omitempty"`

	// Name: Output only. The name of this group. The format is
	// "projects/{project_id_or_number}/groups/{group_id}". When creating a
	// group, this field is ignored and a new name is created consisting of
	// the project specified in the call to CreateGroup and a unique
	// {group_id} that is generated automatically.
	Name string `json:"name,omitempty"`

	// ParentName: The name of the group's parent, if it has one. The format
	// is "projects/{project_id_or_number}/groups/{group_id}". For groups
	// with no parent, parentName is the empty string, "".
	ParentName string `json:"parentName,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "DisplayName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DisplayName") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Group) MarshalJSON() ([]byte, error) {
	type NoMethod Group
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// HttpCheck: Information involved in an HTTP/HTTPS uptime check
// request.
type HttpCheck struct {
	// AuthInfo: The authentication information. Optional when creating an
	// HTTP check; defaults to empty.
	AuthInfo *BasicAuthentication `json:"authInfo,omitempty"`

	// Headers: The list of headers to send as part of the uptime check
	// request. If two headers have the same key and different values, they
	// should be entered as a single header, with the value being a
	// comma-separated list of all the desired values as described at
	// https://www.w3.org/Protocols/rfc2616/rfc2616.txt (page 31). Entering
	// two separate headers with the same key in a Create call will cause
	// the first to be overwritten by the second. The maximum number of
	// headers allowed is 100.
	Headers map[string]string `json:"headers,omitempty"`

	// MaskHeaders: Boolean specifiying whether to encrypt the header
	// information. Encryption should be specified for any headers related
	// to authentication that you do not wish to be seen when retrieving the
	// configuration. The server will be responsible for encrypting the
	// headers. On Get/List calls, if mask_headers is set to True then the
	// headers will be obscured with ******.
	MaskHeaders bool `json:"maskHeaders,omitempty"`

	// Path: The path to the page to run the check against. Will be combined
	// with the host (specified within the MonitoredResource) and port to
	// construct the full URL. Optional (defaults to "/").
	Path string `json:"path,omitempty"`

	// Port: The port to the page to run the check against. Will be combined
	// with host (specified within the MonitoredResource) and path to
	// construct the full URL. Optional (defaults to 80 without SSL, or 443
	// with SSL).
	Port int64 `json:"port,omitempty"`

	// UseSsl: If true, use HTTPS instead of HTTP to run the check.
	UseSsl bool `json:"useSsl,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AuthInfo") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "AuthInfo") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *HttpCheck) MarshalJSON() ([]byte, error) {
	type NoMethod HttpCheck
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// InternalChecker: Nimbus InternalCheckers.
type InternalChecker struct {
	// CheckerId: The checker ID.
	CheckerId string `json:"checkerId,omitempty"`

	// DisplayName: The checker's human-readable name.
	DisplayName string `json:"displayName,omitempty"`

	// GcpZone: The GCP zone the uptime check should egress from. Only
	// respected for internal uptime checks, where internal_network is
	// specified.
	GcpZone string `json:"gcpZone,omitempty"`

	// Network: The internal network to perform this uptime check on.
	Network string `json:"network,omitempty"`

	// ProjectId: The GCP project ID. Not necessarily the same as the
	// project_id for the config.
	ProjectId string `json:"projectId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CheckerId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CheckerId") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *InternalChecker) MarshalJSON() ([]byte, error) {
	type NoMethod InternalChecker
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// LabelDescriptor: A description of a label.
type LabelDescriptor struct {
	// Description: A human-readable description for the label.
	Description string `json:"description,omitempty"`

	// Key: The label key.
	Key string `json:"key,omitempty"`

	// ValueType: The type of data that can be assigned to the label.
	//
	// Possible values:
	//   "STRING" - A variable-length string. This is the default.
	//   "BOOL" - Boolean; true or false.
	//   "INT64" - A 64-bit signed integer.
	ValueType string `json:"valueType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Description") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *LabelDescriptor) MarshalJSON() ([]byte, error) {
	type NoMethod LabelDescriptor
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Linear: Specifies a linear sequence of buckets that all have the same
// width (except overflow and underflow). Each bucket represents a
// constant absolute uncertainty on the specific value in the
// bucket.There are num_finite_buckets + 2 (= N) buckets. Bucket i has
// the following boundaries:Upper bound (0 <= i < N-1): offset + (width
// * i).  Lower bound (1 <= i < N): offset + (width * (i - 1)).
type Linear struct {
	// NumFiniteBuckets: Must be greater than 0.
	NumFiniteBuckets int64 `json:"numFiniteBuckets,omitempty"`

	// Offset: Lower bound of the first bucket.
	Offset float64 `json:"offset,omitempty"`

	// Width: Must be greater than 0.
	Width float64 `json:"width,omitempty"`

	// ForceSendFields is a list of field names (e.g. "NumFiniteBuckets") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NumFiniteBuckets") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *Linear) MarshalJSON() ([]byte, error) {
	type NoMethod Linear
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Linear) UnmarshalJSON(data []byte) error {
	type NoMethod Linear
	var s1 struct {
		Offset gensupport.JSONFloat64 `json:"offset"`
		Width  gensupport.JSONFloat64 `json:"width"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Offset = float64(s1.Offset)
	s.Width = float64(s1.Width)
	return nil
}

// ListGroupMembersResponse: The ListGroupMembers response.
type ListGroupMembersResponse struct {
	// Members: A set of monitored resources in the group.
	Members []*MonitoredResource `json:"members,omitempty"`

	// NextPageToken: If there are more results than have been returned,
	// then this field is set to a non-empty value. To see the additional
	// results, use that value as pageToken in the next call to this method.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// TotalSize: The total number of elements matching this request.
	TotalSize int64 `json:"totalSize,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Members") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Members") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListGroupMembersResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListGroupMembersResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListGroupsResponse: The ListGroups response.
type ListGroupsResponse struct {
	// Group: The groups that match the specified filters.
	Group []*Group `json:"group,omitempty"`

	// NextPageToken: If there are more results than have been returned,
	// then this field is set to a non-empty value. To see the additional
	// results, use that value as pageToken in the next call to this method.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Group") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Group") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListGroupsResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListGroupsResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListMetricDescriptorsResponse: The ListMetricDescriptors response.
type ListMetricDescriptorsResponse struct {
	// MetricDescriptors: The metric descriptors that are available to the
	// project and that match the value of filter, if present.
	MetricDescriptors []*MetricDescriptor `json:"metricDescriptors,omitempty"`

	// NextPageToken: If there are more results than have been returned,
	// then this field is set to a non-empty value. To see the additional
	// results, use that value as pageToken in the next call to this method.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "MetricDescriptors")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "MetricDescriptors") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *ListMetricDescriptorsResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListMetricDescriptorsResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListMonitoredResourceDescriptorsResponse: The
// ListMonitoredResourceDescriptors response.
type ListMonitoredResourceDescriptorsResponse struct {
	// NextPageToken: If there are more results than have been returned,
	// then this field is set to a non-empty value. To see the additional
	// results, use that value as pageToken in the next call to this method.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ResourceDescriptors: The monitored resource descriptors that are
	// available to this project and that match filter, if present.
	ResourceDescriptors []*MonitoredResourceDescriptor `json:"resourceDescriptors,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NextPageToken") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListMonitoredResourceDescriptorsResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListMonitoredResourceDescriptorsResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListTimeSeriesResponse: The ListTimeSeries response.
type ListTimeSeriesResponse struct {
	// NextPageToken: If there are more results than have been returned,
	// then this field is set to a non-empty value. To see the additional
	// results, use that value as pageToken in the next call to this method.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// TimeSeries: One or more time series that match the filter included in
	// the request.
	TimeSeries []*TimeSeries `json:"timeSeries,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NextPageToken") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListTimeSeriesResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListTimeSeriesResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListUptimeCheckConfigsResponse: The protocol for the
// ListUptimeCheckConfigs response.
type ListUptimeCheckConfigsResponse struct {
	// NextPageToken: This field represents the pagination token to retrieve
	// the next page of results. If the value is empty, it means no further
	// results for the request. To retrieve the next page of results, the
	// value of the next_page_token is passed to the subsequent List method
	// call (in the request message's page_token field).
	NextPageToken string `json:"nextPageToken,omitempty"`

	// TotalSize: The total number of uptime check configurations for the
	// project, irrespective of any pagination.
	TotalSize int64 `json:"totalSize,omitempty"`

	// UptimeCheckConfigs: The returned uptime check configurations.
	UptimeCheckConfigs []*UptimeCheckConfig `json:"uptimeCheckConfigs,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NextPageToken") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListUptimeCheckConfigsResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListUptimeCheckConfigsResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListUptimeCheckIpsResponse: The protocol for the ListUptimeCheckIps
// response.
type ListUptimeCheckIpsResponse struct {
	// NextPageToken: This field represents the pagination token to retrieve
	// the next page of results. If the value is empty, it means no further
	// results for the request. To retrieve the next page of results, the
	// value of the next_page_token is passed to the subsequent List method
	// call (in the request message's page_token field). NOTE: this field is
	// not yet implemented
	NextPageToken string `json:"nextPageToken,omitempty"`

	// UptimeCheckIps: The returned list of IP addresses (including region
	// and location) that the checkers run from.
	UptimeCheckIps []*UptimeCheckIp `json:"uptimeCheckIps,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NextPageToken") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListUptimeCheckIpsResponse) MarshalJSON() ([]byte, error) {
	type NoMethod ListUptimeCheckIpsResponse
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Metric: A specific metric, identified by specifying values for all of
// the labels of a MetricDescriptor.
type Metric struct {
	// Labels: The set of label values that uniquely identify this metric.
	// All labels listed in the MetricDescriptor must be assigned values.
	Labels map[string]string `json:"labels,omitempty"`

	// Type: An existing metric type, see google.api.MetricDescriptor. For
	// example, custom.googleapis.com/invoice/paid/amount.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Labels") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Labels") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Metric) MarshalJSON() ([]byte, error) {
	type NoMethod Metric
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// MetricDescriptor: Defines a metric type and its schema. Once a metric
// descriptor is created, deleting or altering it stops data collection
// and makes the metric type's existing data unusable.
type MetricDescriptor struct {
	// Description: A detailed description of the metric, which can be used
	// in documentation.
	Description string `json:"description,omitempty"`

	// DisplayName: A concise name for the metric, which can be displayed in
	// user interfaces. Use sentence case without an ending period, for
	// example "Request count". This field is optional but it is recommended
	// to be set for any metrics associated with user-visible concepts, such
	// as Quota.
	DisplayName string `json:"displayName,omitempty"`

	// Labels: The set of labels that can be used to describe a specific
	// instance of this metric type. For example, the
	// appengine.googleapis.com/http/server/response_latencies metric type
	// has a label for the HTTP response code, response_code, so you can
	// look at latencies for successful responses or just for responses that
	// failed.
	Labels []*LabelDescriptor `json:"labels,omitempty"`

	// MetricKind: Whether the metric records instantaneous values, changes
	// to a value, etc. Some combinations of metric_kind and value_type
	// might not be supported.
	//
	// Possible values:
	//   "METRIC_KIND_UNSPECIFIED" - Do not use this default value.
	//   "GAUGE" - An instantaneous measurement of a value.
	//   "DELTA" - The change in a value during a time interval.
	//   "CUMULATIVE" - A value accumulated over a time interval. Cumulative
	// measurements in a time series should have the same start time and
	// increasing end times, until an event resets the cumulative value to
	// zero and sets a new start time for the following points.
	MetricKind string `json:"metricKind,omitempty"`

	// Name: The resource name of the metric descriptor.
	Name string `json:"name,omitempty"`

	// Type: The metric type, including its DNS name prefix. The type is not
	// URL-encoded. All user-defined custom metric types have the DNS name
	// custom.googleapis.com. Metric types should use a natural hierarchical
	// grouping. For
	// example:
	// "custom.googleapis.com/invoice/paid/amount"
	// "appengine.google
	// apis.com/http/server/response_latencies"
	//
	Type string `json:"type,omitempty"`

	// Unit: Optional. The unit in which the metric value is reported. For
	// example, kBy/s means kilobytes/sec, and 1 is the dimensionless unit.
	// The supported units are a subset of The Unified Code for Units of
	// Measure standard (http://unitsofmeasure.org/ucum.html).<br><br> This
	// field is part of the metric's documentation, but it is ignored by
	// Stackdriver.
	Unit string `json:"unit,omitempty"`

	// ValueType: Whether the measurement is an integer, a floating-point
	// number, etc. Some combinations of metric_kind and value_type might
	// not be supported.
	//
	// Possible values:
	//   "VALUE_TYPE_UNSPECIFIED" - Do not use this default value.
	//   "BOOL" - The value is a boolean. This value type can be used only
	// if the metric kind is GAUGE.
	//   "INT64" - The value is a signed 64-bit integer.
	//   "DOUBLE" - The value is a double precision floating point number.
	//   "STRING" - The value is a text string. This value type can be used
	// only if the metric kind is GAUGE.
	//   "DISTRIBUTION" - The value is a Distribution.
	//   "MONEY" - The value is money.
	ValueType string `json:"valueType,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Description") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *MetricDescriptor) MarshalJSON() ([]byte, error) {
	type NoMethod MetricDescriptor
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// MonitoredResource: An object representing a resource that can be used
// for monitoring, logging, billing, or other purposes. Examples include
// virtual machine instances, databases, and storage devices such as
// disks. The type field identifies a MonitoredResourceDescriptor object
// that describes the resource's schema. Information in the labels field
// identifies the actual resource and its attributes according to the
// schema. For example, a particular Compute Engine VM instance could be
// represented by the following object, because the
// MonitoredResourceDescriptor for "gce_instance" has labels
// "instance_id" and "zone":
// { "type": "gce_instance",
//   "labels": { "instance_id": "12345678901234",
//               "zone": "us-central1-a" }}
//
type MonitoredResource struct {
	// Labels: Required. Values for all of the labels listed in the
	// associated monitored resource descriptor. For example, Compute Engine
	// VM instances use the labels "project_id", "instance_id", and "zone".
	Labels map[string]string `json:"labels,omitempty"`

	// Type: Required. The monitored resource type. This field must match
	// the type field of a MonitoredResourceDescriptor object. For example,
	// the type of a Compute Engine VM instance is gce_instance.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Labels") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Labels") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *MonitoredResource) MarshalJSON() ([]byte, error) {
	type NoMethod MonitoredResource
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// MonitoredResourceDescriptor: An object that describes the schema of a
// MonitoredResource object using a type name and a set of labels. For
// example, the monitored resource descriptor for Google Compute Engine
// VM instances has a type of "gce_instance" and specifies the use of
// the labels "instance_id" and "zone" to identify particular VM
// instances.Different APIs can support different monitored resource
// types. APIs generally provide a list method that returns the
// monitored resource descriptors used by the API.
type MonitoredResourceDescriptor struct {
	// Description: Optional. A detailed description of the monitored
	// resource type that might be used in documentation.
	Description string `json:"description,omitempty"`

	// DisplayName: Optional. A concise name for the monitored resource type
	// that might be displayed in user interfaces. It should be a Title
	// Cased Noun Phrase, without any article or other determiners. For
	// example, "Google Cloud SQL Database".
	DisplayName string `json:"displayName,omitempty"`

	// Labels: Required. A set of labels used to describe instances of this
	// monitored resource type. For example, an individual Google Cloud SQL
	// database is identified by values for the labels "database_id" and
	// "zone".
	Labels []*LabelDescriptor `json:"labels,omitempty"`

	// Name: Optional. The resource name of the monitored resource
	// descriptor:
	// "projects/{project_id}/monitoredResourceDescriptors/{type}" where
	// {type} is the value of the type field in this object and {project_id}
	// is a project ID that provides API-specific context for accessing the
	// type. APIs that do not use project information can use the resource
	// name format "monitoredResourceDescriptors/{type}".
	Name string `json:"name,omitempty"`

	// Type: Required. The monitored resource type. For example, the type
	// "cloudsql_database" represents databases in Google Cloud SQL. The
	// maximum length of this value is 256 characters.
	Type string `json:"type,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Description") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *MonitoredResourceDescriptor) MarshalJSON() ([]byte, error) {
	type NoMethod MonitoredResourceDescriptor
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// MonitoredResourceMetadata: Auxiliary metadata for a MonitoredResource
// object. MonitoredResource objects contain the minimum set of
// information to uniquely identify a monitored resource instance. There
// is some other useful auxiliary metadata. Google Stackdriver
// Monitoring & Logging uses an ingestion pipeline to extract metadata
// for cloud resources of all types , and stores the metadata in this
// message.
type MonitoredResourceMetadata struct {
	// SystemLabels: Output only. Values for predefined system metadata
	// labels. System labels are a kind of metadata extracted by Google
	// Stackdriver. Stackdriver determines what system labels are useful and
	// how to obtain their values. Some examples: "machine_image", "vpc",
	// "subnet_id", "security_group", "name", etc. System label values can
	// be only strings, Boolean values, or a list of strings. For example:
	// { "name": "my-test-instance",
	//   "security_group": ["a", "b", "c"],
	//   "spot_instance": false }
	//
	SystemLabels googleapi.RawMessage `json:"systemLabels,omitempty"`

	// UserLabels: Output only. A map of user-defined metadata labels.
	UserLabels map[string]string `json:"userLabels,omitempty"`

	// ForceSendFields is a list of field names (e.g. "SystemLabels") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "SystemLabels") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *MonitoredResourceMetadata) MarshalJSON() ([]byte, error) {
	type NoMethod MonitoredResourceMetadata
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Option: A protocol buffer option, which can be attached to a message,
// field, enumeration, etc.
type Option struct {
	// Name: The option's name. For protobuf built-in options (options
	// defined in descriptor.proto), this is the short name. For example,
	// "map_entry". For custom options, it should be the fully-qualified
	// name. For example, "google.api.http".
	Name string `json:"name,omitempty"`

	// Value: The option's value packed in an Any message. If the value is a
	// primitive, the corresponding wrapper type defined in
	// google/protobuf/wrappers.proto should be used. If the value is an
	// enum, it should be stored as an int32 value using the
	// google.protobuf.Int32Value type.
	Value googleapi.RawMessage `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Name") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Option) MarshalJSON() ([]byte, error) {
	type NoMethod Option
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Point: A single data point in a time series.
type Point struct {
	// Interval: The time interval to which the data point applies. For
	// GAUGE metrics, only the end time of the interval is used. For DELTA
	// metrics, the start and end time should specify a non-zero interval,
	// with subsequent points specifying contiguous and non-overlapping
	// intervals. For CUMULATIVE metrics, the start and end time should
	// specify a non-zero interval, with subsequent points specifying the
	// same start time and increasing end times, until an event resets the
	// cumulative value to zero and sets a new start time for the following
	// points.
	Interval *TimeInterval `json:"interval,omitempty"`

	// Value: The value of the data point.
	Value *TypedValue `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Interval") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Interval") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Point) MarshalJSON() ([]byte, error) {
	type NoMethod Point
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Range: The range of the population values.
type Range struct {
	// Max: The maximum of the population values.
	Max float64 `json:"max,omitempty"`

	// Min: The minimum of the population values.
	Min float64 `json:"min,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Max") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Max") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Range) MarshalJSON() ([]byte, error) {
	type NoMethod Range
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *Range) UnmarshalJSON(data []byte) error {
	type NoMethod Range
	var s1 struct {
		Max gensupport.JSONFloat64 `json:"max"`
		Min gensupport.JSONFloat64 `json:"min"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	s.Max = float64(s1.Max)
	s.Min = float64(s1.Min)
	return nil
}

// ResourceGroup: The resource submessage for group checks. It can be
// used instead of a monitored resource, when multiple resources are
// being monitored.
type ResourceGroup struct {
	// GroupId: The group of resources being monitored. Should be only the
	// group_id, not projects/<project_id>/groups/<group_id>.
	GroupId string `json:"groupId,omitempty"`

	// ResourceType: The resource type of the group members.
	//
	// Possible values:
	//   "RESOURCE_TYPE_UNSPECIFIED" - Default value (not valid).
	//   "INSTANCE" - A group of instances from Google Cloud Platform (GCP)
	// or Amazon Web Services (AWS).
	//   "AWS_ELB_LOAD_BALANCER" - A group of Amazon ELB load balancers.
	ResourceType string `json:"resourceType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GroupId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "GroupId") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ResourceGroup) MarshalJSON() ([]byte, error) {
	type NoMethod ResourceGroup
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// SourceContext: SourceContext represents information about the source
// of a protobuf element, like the file in which it is defined.
type SourceContext struct {
	// FileName: The path-qualified name of the .proto file that contained
	// the associated protobuf element. For example:
	// "google/protobuf/source_context.proto".
	FileName string `json:"fileName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FileName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "FileName") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *SourceContext) MarshalJSON() ([]byte, error) {
	type NoMethod SourceContext
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Status: The Status type defines a logical error model that is
// suitable for different programming environments, including REST APIs
// and RPC APIs. It is used by gRPC (https://github.com/grpc). The error
// model is designed to be:
// Simple to use and understand for most users
// Flexible enough to meet unexpected needsOverviewThe Status message
// contains three pieces of data: error code, error message, and error
// details. The error code should be an enum value of google.rpc.Code,
// but it may accept additional error codes if needed. The error message
// should be a developer-facing English message that helps developers
// understand and resolve the error. If a localized user-facing error
// message is needed, put the localized message in the error details or
// localize it in the client. The optional error details may contain
// arbitrary information about the error. There is a predefined set of
// error detail types in the package google.rpc that can be used for
// common error conditions.Language mappingThe Status message is the
// logical representation of the error model, but it is not necessarily
// the actual wire format. When the Status message is exposed in
// different client libraries and different wire protocols, it can be
// mapped differently. For example, it will likely be mapped to some
// exceptions in Java, but more likely mapped to some error codes in
// C.Other usesThe error model and the Status message can be used in a
// variety of environments, either with or without APIs, to provide a
// consistent developer experience across different environments.Example
// uses of this error model include:
// Partial errors. If a service needs to return partial errors to the
// client, it may embed the Status in the normal response to indicate
// the partial errors.
// Workflow errors. A typical workflow has multiple steps. Each step may
// have a Status message for error reporting.
// Batch operations. If a client uses batch request and batch response,
// the Status message should be used directly inside batch response, one
// for each error sub-response.
// Asynchronous operations. If an API call embeds asynchronous operation
// results in its response, the status of those operations should be
// represented directly using the Status message.
// Logging. If some API errors are stored in logs, the message Status
// could be used directly after any stripping needed for
// security/privacy reasons.
type Status struct {
	// Code: The status code, which should be an enum value of
	// google.rpc.Code.
	Code int64 `json:"code,omitempty"`

	// Details: A list of messages that carry the error details. There is a
	// common set of message types for APIs to use.
	Details []googleapi.RawMessage `json:"details,omitempty"`

	// Message: A developer-facing error message, which should be in
	// English. Any user-facing error message should be localized and sent
	// in the google.rpc.Status.details field, or localized by the client.
	Message string `json:"message,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Code") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Code") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Status) MarshalJSON() ([]byte, error) {
	type NoMethod Status
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// TcpCheck: Information required for a TCP uptime check request.
type TcpCheck struct {
	// Port: The port to the page to run the check against. Will be combined
	// with host (specified within the MonitoredResource) to construct the
	// full URL. Required.
	Port int64 `json:"port,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Port") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Port") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *TcpCheck) MarshalJSON() ([]byte, error) {
	type NoMethod TcpCheck
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// TimeInterval: A time interval extending just after a start time
// through an end time. If the start time is the same as the end time,
// then the interval represents a single point in time.
type TimeInterval struct {
	// EndTime: Required. The end of the time interval.
	EndTime string `json:"endTime,omitempty"`

	// StartTime: Optional. The beginning of the time interval. The default
	// value for the start time is the end time. The start time must not be
	// later than the end time.
	StartTime string `json:"startTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "EndTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "EndTime") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *TimeInterval) MarshalJSON() ([]byte, error) {
	type NoMethod TimeInterval
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// TimeSeries: A collection of data points that describes the
// time-varying values of a metric. A time series is identified by a
// combination of a fully-specified monitored resource and a
// fully-specified metric. This type is used for both listing and
// creating time series.
type TimeSeries struct {
	// Metadata: Output only. The associated monitored resource metadata.
	// When reading a a timeseries, this field will include metadata labels
	// that are explicitly named in the reduction. When creating a
	// timeseries, this field is ignored.
	Metadata *MonitoredResourceMetadata `json:"metadata,omitempty"`

	// Metric: The associated metric. A fully-specified metric used to
	// identify the time series.
	Metric *Metric `json:"metric,omitempty"`

	// MetricKind: The metric kind of the time series. When listing time
	// series, this metric kind might be different from the metric kind of
	// the associated metric if this time series is an alignment or
	// reduction of other time series.When creating a time series, this
	// field is optional. If present, it must be the same as the metric kind
	// of the associated metric. If the associated metric's descriptor must
	// be auto-created, then this field specifies the metric kind of the new
	// descriptor and must be either GAUGE (the default) or CUMULATIVE.
	//
	// Possible values:
	//   "METRIC_KIND_UNSPECIFIED" - Do not use this default value.
	//   "GAUGE" - An instantaneous measurement of a value.
	//   "DELTA" - The change in a value during a time interval.
	//   "CUMULATIVE" - A value accumulated over a time interval. Cumulative
	// measurements in a time series should have the same start time and
	// increasing end times, until an event resets the cumulative value to
	// zero and sets a new start time for the following points.
	MetricKind string `json:"metricKind,omitempty"`

	// Points: The data points of this time series. When listing time
	// series, points are returned in reverse time order.When creating a
	// time series, this field must contain exactly one point and the
	// point's type must be the same as the value type of the associated
	// metric. If the associated metric's descriptor must be auto-created,
	// then the value type of the descriptor is determined by the point's
	// type, which must be BOOL, INT64, DOUBLE, or DISTRIBUTION.
	Points []*Point `json:"points,omitempty"`

	// Resource: The associated monitored resource. Custom metrics can use
	// only certain monitored resource types in their time series data.
	Resource *MonitoredResource `json:"resource,omitempty"`

	// ValueType: The value type of the time series. When listing time
	// series, this value type might be different from the value type of the
	// associated metric if this time series is an alignment or reduction of
	// other time series.When creating a time series, this field is
	// optional. If present, it must be the same as the type of the data in
	// the points field.
	//
	// Possible values:
	//   "VALUE_TYPE_UNSPECIFIED" - Do not use this default value.
	//   "BOOL" - The value is a boolean. This value type can be used only
	// if the metric kind is GAUGE.
	//   "INT64" - The value is a signed 64-bit integer.
	//   "DOUBLE" - The value is a double precision floating point number.
	//   "STRING" - The value is a text string. This value type can be used
	// only if the metric kind is GAUGE.
	//   "DISTRIBUTION" - The value is a Distribution.
	//   "MONEY" - The value is money.
	ValueType string `json:"valueType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Metadata") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *TimeSeries) MarshalJSON() ([]byte, error) {
	type NoMethod TimeSeries
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Type: A protocol buffer message type.
type Type struct {
	// Fields: The list of fields.
	Fields []*Field `json:"fields,omitempty"`

	// Name: The fully qualified message name.
	Name string `json:"name,omitempty"`

	// Oneofs: The list of types appearing in oneof definitions in this
	// type.
	Oneofs []string `json:"oneofs,omitempty"`

	// Options: The protocol buffer options.
	Options []*Option `json:"options,omitempty"`

	// SourceContext: The source context.
	SourceContext *SourceContext `json:"sourceContext,omitempty"`

	// Syntax: The source syntax.
	//
	// Possible values:
	//   "SYNTAX_PROTO2" - Syntax proto2.
	//   "SYNTAX_PROTO3" - Syntax proto3.
	Syntax string `json:"syntax,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Fields") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Fields") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Type) MarshalJSON() ([]byte, error) {
	type NoMethod Type
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// TypedValue: A single strongly-typed value.
type TypedValue struct {
	// BoolValue: A Boolean value: true or false.
	BoolValue *bool `json:"boolValue,omitempty"`

	// DistributionValue: A distribution value.
	DistributionValue *Distribution `json:"distributionValue,omitempty"`

	// DoubleValue: A 64-bit double-precision floating-point number. Its
	// magnitude is approximately &plusmn;10<sup>&plusmn;300</sup> and it
	// has 16 significant digits of precision.
	DoubleValue *float64 `json:"doubleValue,omitempty"`

	// Int64Value: A 64-bit integer. Its range is approximately
	// &plusmn;9.2x10<sup>18</sup>.
	Int64Value *int64 `json:"int64Value,omitempty,string"`

	// StringValue: A variable-length string value.
	StringValue *string `json:"stringValue,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoolValue") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BoolValue") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *TypedValue) MarshalJSON() ([]byte, error) {
	type NoMethod TypedValue
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

func (s *TypedValue) UnmarshalJSON(data []byte) error {
	type NoMethod TypedValue
	var s1 struct {
		DoubleValue *gensupport.JSONFloat64 `json:"doubleValue"`
		*NoMethod
	}
	s1.NoMethod = (*NoMethod)(s)
	if err := json.Unmarshal(data, &s1); err != nil {
		return err
	}
	if s1.DoubleValue != nil {
		s.DoubleValue = (*float64)(s1.DoubleValue)
	}
	return nil
}

// UptimeCheckConfig: This message configures which resources and
// services to monitor for availability.
type UptimeCheckConfig struct {
	// ContentMatchers: The expected content on the page the check is run
	// against. Currently, only the first entry in the list is supported,
	// and other entries will be ignored. The server will look for an exact
	// match of the string in the page response's content. This field is
	// optional and should only be specified if a content match is required.
	ContentMatchers []*ContentMatcher `json:"contentMatchers,omitempty"`

	// DisplayName: A human-friendly name for the uptime check
	// configuration. The display name should be unique within a Stackdriver
	// Account in order to make it easier to identify; however, uniqueness
	// is not enforced. Required.
	DisplayName string `json:"displayName,omitempty"`

	// HttpCheck: Contains information needed to make an HTTP or HTTPS
	// check.
	HttpCheck *HttpCheck `json:"httpCheck,omitempty"`

	// InternalCheckers: The internal checkers that this check will egress
	// from. If is_internal is true and this list is empty, the check will
	// egress from all InternalCheckers configured for the project that owns
	// this CheckConfig.
	InternalCheckers []*InternalChecker `json:"internalCheckers,omitempty"`

	// IsInternal: Denotes whether this is a check that egresses from
	// InternalCheckers.
	IsInternal bool `json:"isInternal,omitempty"`

	// MonitoredResource: The monitored resource
	// (https://cloud.google.com/monitoring/api/resources) associated with
	// the configuration. The following monitored resource types are
	// supported for uptime checks:  uptime_url  gce_instance  gae_app
	// aws_ec2_instance  aws_elb_load_balancer
	MonitoredResource *MonitoredResource `json:"monitoredResource,omitempty"`

	// Name: A unique resource name for this UptimeCheckConfig. The format
	// is:projects/[PROJECT_ID]/uptimeCheckConfigs/[UPTIME_CHECK_ID].This
	// field should be omitted when creating the uptime check configuration;
	// on create, the resource name is assigned by the server and included
	// in the response.
	Name string `json:"name,omitempty"`

	// Period: How often, in seconds, the uptime check is performed.
	// Currently, the only supported values are 60s (1 minute), 300s (5
	// minutes), 600s (10 minutes), and 900s (15 minutes). Required.
	Period string `json:"period,omitempty"`

	// ResourceGroup: The group resource associated with the configuration.
	ResourceGroup *ResourceGroup `json:"resourceGroup,omitempty"`

	// SelectedRegions: The list of regions from which the check will be
	// run. If this field is specified, enough regions to include a minimum
	// of 3 locations must be provided, or an error message is returned. Not
	// specifying this field will result in uptime checks running from all
	// regions.
	//
	// Possible values:
	//   "REGION_UNSPECIFIED" - Default value if no region is specified.
	// Will result in uptime checks running from all regions.
	//   "USA" - Allows checks to run from locations within the United
	// States of America.
	//   "EUROPE" - Allows checks to run from locations within the continent
	// of Europe.
	//   "SOUTH_AMERICA" - Allows checks to run from locations within the
	// continent of South America.
	//   "ASIA_PACIFIC" - Allows checks to run from locations within the
	// Asia Pacific area (ex: Singapore).
	SelectedRegions []string `json:"selectedRegions,omitempty"`

	// TcpCheck: Contains information needed to make a TCP check.
	TcpCheck *TcpCheck `json:"tcpCheck,omitempty"`

	// Timeout: The maximum amount of time to wait for the request to
	// complete (must be between 1 and 60 seconds). Required.
	Timeout string `json:"timeout,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ContentMatchers") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ContentMatchers") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *UptimeCheckConfig) MarshalJSON() ([]byte, error) {
	type NoMethod UptimeCheckConfig
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// UptimeCheckIp: Contains the region, location, and list of IP
// addresses where checkers in the location run from.
type UptimeCheckIp struct {
	// IpAddress: The IP address from which the uptime check originates.
	// This is a full IP address (not an IP address range). Most IP
	// addresses, as of this publication, are in IPv4 format; however, one
	// should not rely on the IP addresses being in IPv4 format indefinitely
	// and should support interpreting this field in either IPv4 or IPv6
	// format.
	IpAddress string `json:"ipAddress,omitempty"`

	// Location: A more specific location within the region that typically
	// encodes a particular city/town/metro (and its containing
	// state/province or country) within the broader umbrella region
	// category.
	Location string `json:"location,omitempty"`

	// Region: A broad region category in which the IP address is located.
	//
	// Possible values:
	//   "REGION_UNSPECIFIED" - Default value if no region is specified.
	// Will result in uptime checks running from all regions.
	//   "USA" - Allows checks to run from locations within the United
	// States of America.
	//   "EUROPE" - Allows checks to run from locations within the continent
	// of Europe.
	//   "SOUTH_AMERICA" - Allows checks to run from locations within the
	// continent of South America.
	//   "ASIA_PACIFIC" - Allows checks to run from locations within the
	// Asia Pacific area (ex: Singapore).
	Region string `json:"region,omitempty"`

	// ForceSendFields is a list of field names (e.g. "IpAddress") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "IpAddress") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *UptimeCheckIp) MarshalJSON() ([]byte, error) {
	type NoMethod UptimeCheckIp
	raw := NoMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// method id "monitoring.projects.collectdTimeSeries.create":

type ProjectsCollectdTimeSeriesCreateCall struct {
	s                               *Service
	name                            string
	createcollectdtimeseriesrequest *CreateCollectdTimeSeriesRequest
	urlParams_                      gensupport.URLParams
	ctx_                            context.Context
	header_                         http.Header
}

// Create: Stackdriver Monitoring Agent only: Creates a new time
// series.<aside class="caution">This method is only for use by the
// Stackdriver Monitoring Agent. Use projects.timeSeries.create
// instead.</aside>
func (r *ProjectsCollectdTimeSeriesService) Create(name string, createcollectdtimeseriesrequest *CreateCollectdTimeSeriesRequest) *ProjectsCollectdTimeSeriesCreateCall {
	c := &ProjectsCollectdTimeSeriesCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.createcollectdtimeseriesrequest = createcollectdtimeseriesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsCollectdTimeSeriesCreateCall) Fields(s ...googleapi.Field) *ProjectsCollectdTimeSeriesCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsCollectdTimeSeriesCreateCall) Context(ctx context.Context) *ProjectsCollectdTimeSeriesCreateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsCollectdTimeSeriesCreateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsCollectdTimeSeriesCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createcollectdtimeseriesrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/collectdTimeSeries")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.collectdTimeSeries.create" call.
// Exactly one of *CreateCollectdTimeSeriesResponse or error will be
// non-nil. Any non-2xx status code is an error. Response headers are in
// either *CreateCollectdTimeSeriesResponse.ServerResponse.Header or (if
// a response was returned at all) in error.(*googleapi.Error).Header.
// Use googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsCollectdTimeSeriesCreateCall) Do(opts ...googleapi.CallOption) (*CreateCollectdTimeSeriesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &CreateCollectdTimeSeriesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Stackdriver Monitoring Agent only: Creates a new time series.\u003caside class=\"caution\"\u003eThis method is only for use by the Stackdriver Monitoring Agent. Use projects.timeSeries.create instead.\u003c/aside\u003e",
	//   "flatPath": "v3/projects/{projectsId}/collectdTimeSeries",
	//   "httpMethod": "POST",
	//   "id": "monitoring.projects.collectdTimeSeries.create",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The project in which to create the time series. The format is \"projects/PROJECT_ID_OR_NUMBER\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/collectdTimeSeries",
	//   "request": {
	//     "$ref": "CreateCollectdTimeSeriesRequest"
	//   },
	//   "response": {
	//     "$ref": "CreateCollectdTimeSeriesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// method id "monitoring.projects.groups.create":

type ProjectsGroupsCreateCall struct {
	s          *Service
	name       string
	group      *Group
	urlParams_ gensupport.URLParams
	ctx_       context.Context
	header_    http.Header
}

// Create: Creates a new group.
func (r *ProjectsGroupsService) Create(name string, group *Group) *ProjectsGroupsCreateCall {
	c := &ProjectsGroupsCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.group = group
	return c
}

// ValidateOnly sets the optional parameter "validateOnly": If true,
// validate this request but do not create the group.
func (c *ProjectsGroupsCreateCall) ValidateOnly(validateOnly bool) *ProjectsGroupsCreateCall {
	c.urlParams_.Set("validateOnly", fmt.Sprint(validateOnly))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsGroupsCreateCall) Fields(s ...googleapi.Field) *ProjectsGroupsCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsGroupsCreateCall) Context(ctx context.Context) *ProjectsGroupsCreateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsGroupsCreateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsGroupsCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.group)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/groups")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.groups.create" call.
// Exactly one of *Group or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Group.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsGroupsCreateCall) Do(opts ...googleapi.CallOption) (*Group, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Group{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a new group.",
	//   "flatPath": "v3/projects/{projectsId}/groups",
	//   "httpMethod": "POST",
	//   "id": "monitoring.projects.groups.create",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The project in which to create the group. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "validateOnly": {
	//       "description": "If true, validate this request but do not create the group.",
	//       "location": "query",
	//       "type": "boolean"
	//     }
	//   },
	//   "path": "v3/{+name}/groups",
	//   "request": {
	//     "$ref": "Group"
	//   },
	//   "response": {
	//     "$ref": "Group"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.projects.groups.delete":

type ProjectsGroupsDeleteCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
	header_    http.Header
}

// Delete: Deletes an existing group.
func (r *ProjectsGroupsService) Delete(name string) *ProjectsGroupsDeleteCall {
	c := &ProjectsGroupsDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsGroupsDeleteCall) Fields(s ...googleapi.Field) *ProjectsGroupsDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsGroupsDeleteCall) Context(ctx context.Context) *ProjectsGroupsDeleteCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsGroupsDeleteCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsGroupsDeleteCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.groups.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsGroupsDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Deletes an existing group.",
	//   "flatPath": "v3/projects/{projectsId}/groups/{groupsId}",
	//   "httpMethod": "DELETE",
	//   "id": "monitoring.projects.groups.delete",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The group to delete. The format is \"projects/{project_id_or_number}/groups/{group_id}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/groups/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.projects.groups.get":

type ProjectsGroupsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// Get: Gets a single group.
func (r *ProjectsGroupsService) Get(name string) *ProjectsGroupsGetCall {
	c := &ProjectsGroupsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsGroupsGetCall) Fields(s ...googleapi.Field) *ProjectsGroupsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsGroupsGetCall) IfNoneMatch(entityTag string) *ProjectsGroupsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsGroupsGetCall) Context(ctx context.Context) *ProjectsGroupsGetCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsGroupsGetCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsGroupsGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.groups.get" call.
// Exactly one of *Group or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Group.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsGroupsGetCall) Do(opts ...googleapi.CallOption) (*Group, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Group{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a single group.",
	//   "flatPath": "v3/projects/{projectsId}/groups/{groupsId}",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.groups.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The group to retrieve. The format is \"projects/{project_id_or_number}/groups/{group_id}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/groups/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "Group"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// method id "monitoring.projects.groups.list":

type ProjectsGroupsListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Lists the existing groups.
func (r *ProjectsGroupsService) List(name string) *ProjectsGroupsListCall {
	c := &ProjectsGroupsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// AncestorsOfGroup sets the optional parameter "ancestorsOfGroup": A
// group name: "projects/{project_id_or_number}/groups/{group_id}".
// Returns groups that are ancestors of the specified group. The groups
// are returned in order, starting with the immediate parent and ending
// with the most distant ancestor. If the specified group has no
// immediate parent, the results are empty.
func (c *ProjectsGroupsListCall) AncestorsOfGroup(ancestorsOfGroup string) *ProjectsGroupsListCall {
	c.urlParams_.Set("ancestorsOfGroup", ancestorsOfGroup)
	return c
}

// ChildrenOfGroup sets the optional parameter "childrenOfGroup": A
// group name: "projects/{project_id_or_number}/groups/{group_id}".
// Returns groups whose parentName field contains the group name. If no
// groups have this parent, the results are empty.
func (c *ProjectsGroupsListCall) ChildrenOfGroup(childrenOfGroup string) *ProjectsGroupsListCall {
	c.urlParams_.Set("childrenOfGroup", childrenOfGroup)
	return c
}

// DescendantsOfGroup sets the optional parameter "descendantsOfGroup":
// A group name: "projects/{project_id_or_number}/groups/{group_id}".
// Returns the descendants of the specified group. This is a superset of
// the results returned by the childrenOfGroup filter, and includes
// children-of-children, and so forth.
func (c *ProjectsGroupsListCall) DescendantsOfGroup(descendantsOfGroup string) *ProjectsGroupsListCall {
	c.urlParams_.Set("descendantsOfGroup", descendantsOfGroup)
	return c
}

// PageSize sets the optional parameter "pageSize": A positive number
// that is the maximum number of results to return.
func (c *ProjectsGroupsListCall) PageSize(pageSize int64) *ProjectsGroupsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return additional results from the previous method call.
func (c *ProjectsGroupsListCall) PageToken(pageToken string) *ProjectsGroupsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsGroupsListCall) Fields(s ...googleapi.Field) *ProjectsGroupsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsGroupsListCall) IfNoneMatch(entityTag string) *ProjectsGroupsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsGroupsListCall) Context(ctx context.Context) *ProjectsGroupsListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsGroupsListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsGroupsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/groups")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.groups.list" call.
// Exactly one of *ListGroupsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListGroupsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsGroupsListCall) Do(opts ...googleapi.CallOption) (*ListGroupsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListGroupsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists the existing groups.",
	//   "flatPath": "v3/projects/{projectsId}/groups",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.groups.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "ancestorsOfGroup": {
	//       "description": "A group name: \"projects/{project_id_or_number}/groups/{group_id}\". Returns groups that are ancestors of the specified group. The groups are returned in order, starting with the immediate parent and ending with the most distant ancestor. If the specified group has no immediate parent, the results are empty.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "childrenOfGroup": {
	//       "description": "A group name: \"projects/{project_id_or_number}/groups/{group_id}\". Returns groups whose parentName field contains the group name. If no groups have this parent, the results are empty.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "descendantsOfGroup": {
	//       "description": "A group name: \"projects/{project_id_or_number}/groups/{group_id}\". Returns the descendants of the specified group. This is a superset of the results returned by the childrenOfGroup filter, and includes children-of-children, and so forth.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The project whose groups are to be listed. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "A positive number that is the maximum number of results to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return additional results from the previous method call.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/groups",
	//   "response": {
	//     "$ref": "ListGroupsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsGroupsListCall) Pages(ctx context.Context, f func(*ListGroupsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "monitoring.projects.groups.update":

type ProjectsGroupsUpdateCall struct {
	s          *Service
	name       string
	group      *Group
	urlParams_ gensupport.URLParams
	ctx_       context.Context
	header_    http.Header
}

// Update: Updates an existing group. You can change any group
// attributes except name.
func (r *ProjectsGroupsService) Update(name string, group *Group) *ProjectsGroupsUpdateCall {
	c := &ProjectsGroupsUpdateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.group = group
	return c
}

// ValidateOnly sets the optional parameter "validateOnly": If true,
// validate this request but do not update the existing group.
func (c *ProjectsGroupsUpdateCall) ValidateOnly(validateOnly bool) *ProjectsGroupsUpdateCall {
	c.urlParams_.Set("validateOnly", fmt.Sprint(validateOnly))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsGroupsUpdateCall) Fields(s ...googleapi.Field) *ProjectsGroupsUpdateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsGroupsUpdateCall) Context(ctx context.Context) *ProjectsGroupsUpdateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsGroupsUpdateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsGroupsUpdateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.group)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.groups.update" call.
// Exactly one of *Group or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Group.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsGroupsUpdateCall) Do(opts ...googleapi.CallOption) (*Group, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Group{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing group. You can change any group attributes except name.",
	//   "flatPath": "v3/projects/{projectsId}/groups/{groupsId}",
	//   "httpMethod": "PUT",
	//   "id": "monitoring.projects.groups.update",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "Output only. The name of this group. The format is \"projects/{project_id_or_number}/groups/{group_id}\". When creating a group, this field is ignored and a new name is created consisting of the project specified in the call to CreateGroup and a unique {group_id} that is generated automatically.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/groups/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "validateOnly": {
	//       "description": "If true, validate this request but do not update the existing group.",
	//       "location": "query",
	//       "type": "boolean"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "request": {
	//     "$ref": "Group"
	//   },
	//   "response": {
	//     "$ref": "Group"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.projects.groups.members.list":

type ProjectsGroupsMembersListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Lists the monitored resources that are members of a group.
func (r *ProjectsGroupsMembersService) List(name string) *ProjectsGroupsMembersListCall {
	c := &ProjectsGroupsMembersListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Filter sets the optional parameter "filter": An optional list filter
// describing the members to be returned. The filter may reference the
// type, labels, and metadata of monitored resources that comprise the
// group. For example, to return only resources representing Compute
// Engine VM instances, use this filter:
// resource.type = "gce_instance"
func (c *ProjectsGroupsMembersListCall) Filter(filter string) *ProjectsGroupsMembersListCall {
	c.urlParams_.Set("filter", filter)
	return c
}

// IntervalEndTime sets the optional parameter "interval.endTime":
// Required. The end of the time interval.
func (c *ProjectsGroupsMembersListCall) IntervalEndTime(intervalEndTime string) *ProjectsGroupsMembersListCall {
	c.urlParams_.Set("interval.endTime", intervalEndTime)
	return c
}

// IntervalStartTime sets the optional parameter "interval.startTime":
// The beginning of the time interval. The default value for the start
// time is the end time. The start time must not be later than the end
// time.
func (c *ProjectsGroupsMembersListCall) IntervalStartTime(intervalStartTime string) *ProjectsGroupsMembersListCall {
	c.urlParams_.Set("interval.startTime", intervalStartTime)
	return c
}

// PageSize sets the optional parameter "pageSize": A positive number
// that is the maximum number of results to return.
func (c *ProjectsGroupsMembersListCall) PageSize(pageSize int64) *ProjectsGroupsMembersListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return additional results from the previous method call.
func (c *ProjectsGroupsMembersListCall) PageToken(pageToken string) *ProjectsGroupsMembersListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsGroupsMembersListCall) Fields(s ...googleapi.Field) *ProjectsGroupsMembersListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsGroupsMembersListCall) IfNoneMatch(entityTag string) *ProjectsGroupsMembersListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsGroupsMembersListCall) Context(ctx context.Context) *ProjectsGroupsMembersListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsGroupsMembersListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsGroupsMembersListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/members")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.groups.members.list" call.
// Exactly one of *ListGroupMembersResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ListGroupMembersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsGroupsMembersListCall) Do(opts ...googleapi.CallOption) (*ListGroupMembersResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListGroupMembersResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists the monitored resources that are members of a group.",
	//   "flatPath": "v3/projects/{projectsId}/groups/{groupsId}/members",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.groups.members.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "filter": {
	//       "description": "An optional list filter describing the members to be returned. The filter may reference the type, labels, and metadata of monitored resources that comprise the group. For example, to return only resources representing Compute Engine VM instances, use this filter:\nresource.type = \"gce_instance\"\n",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "interval.endTime": {
	//       "description": "Required. The end of the time interval.",
	//       "format": "google-datetime",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "interval.startTime": {
	//       "description": "Optional. The beginning of the time interval. The default value for the start time is the end time. The start time must not be later than the end time.",
	//       "format": "google-datetime",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The group whose members are listed. The format is \"projects/{project_id_or_number}/groups/{group_id}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/groups/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "A positive number that is the maximum number of results to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return additional results from the previous method call.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/members",
	//   "response": {
	//     "$ref": "ListGroupMembersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsGroupsMembersListCall) Pages(ctx context.Context, f func(*ListGroupMembersResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "monitoring.projects.metricDescriptors.create":

type ProjectsMetricDescriptorsCreateCall struct {
	s                *Service
	name             string
	metricdescriptor *MetricDescriptor
	urlParams_       gensupport.URLParams
	ctx_             context.Context
	header_          http.Header
}

// Create: Creates a new metric descriptor. User-created metric
// descriptors define custom metrics.
func (r *ProjectsMetricDescriptorsService) Create(name string, metricdescriptor *MetricDescriptor) *ProjectsMetricDescriptorsCreateCall {
	c := &ProjectsMetricDescriptorsCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.metricdescriptor = metricdescriptor
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsMetricDescriptorsCreateCall) Fields(s ...googleapi.Field) *ProjectsMetricDescriptorsCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsMetricDescriptorsCreateCall) Context(ctx context.Context) *ProjectsMetricDescriptorsCreateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsMetricDescriptorsCreateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsMetricDescriptorsCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.metricdescriptor)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/metricDescriptors")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.metricDescriptors.create" call.
// Exactly one of *MetricDescriptor or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *MetricDescriptor.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsMetricDescriptorsCreateCall) Do(opts ...googleapi.CallOption) (*MetricDescriptor, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MetricDescriptor{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a new metric descriptor. User-created metric descriptors define custom metrics.",
	//   "flatPath": "v3/projects/{projectsId}/metricDescriptors",
	//   "httpMethod": "POST",
	//   "id": "monitoring.projects.metricDescriptors.create",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The project on which to execute the request. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/metricDescriptors",
	//   "request": {
	//     "$ref": "MetricDescriptor"
	//   },
	//   "response": {
	//     "$ref": "MetricDescriptor"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// method id "monitoring.projects.metricDescriptors.delete":

type ProjectsMetricDescriptorsDeleteCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
	header_    http.Header
}

// Delete: Deletes a metric descriptor. Only user-created custom metrics
// can be deleted.
func (r *ProjectsMetricDescriptorsService) Delete(name string) *ProjectsMetricDescriptorsDeleteCall {
	c := &ProjectsMetricDescriptorsDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsMetricDescriptorsDeleteCall) Fields(s ...googleapi.Field) *ProjectsMetricDescriptorsDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsMetricDescriptorsDeleteCall) Context(ctx context.Context) *ProjectsMetricDescriptorsDeleteCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsMetricDescriptorsDeleteCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsMetricDescriptorsDeleteCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.metricDescriptors.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsMetricDescriptorsDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Deletes a metric descriptor. Only user-created custom metrics can be deleted.",
	//   "flatPath": "v3/projects/{projectsId}/metricDescriptors/{metricDescriptorsId}",
	//   "httpMethod": "DELETE",
	//   "id": "monitoring.projects.metricDescriptors.delete",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The metric descriptor on which to execute the request. The format is \"projects/{project_id_or_number}/metricDescriptors/{metric_id}\". An example of {metric_id} is: \"custom.googleapis.com/my_test_metric\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/metricDescriptors/.+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.projects.metricDescriptors.get":

type ProjectsMetricDescriptorsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// Get: Gets a single metric descriptor. This method does not require a
// Stackdriver account.
func (r *ProjectsMetricDescriptorsService) Get(name string) *ProjectsMetricDescriptorsGetCall {
	c := &ProjectsMetricDescriptorsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsMetricDescriptorsGetCall) Fields(s ...googleapi.Field) *ProjectsMetricDescriptorsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsMetricDescriptorsGetCall) IfNoneMatch(entityTag string) *ProjectsMetricDescriptorsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsMetricDescriptorsGetCall) Context(ctx context.Context) *ProjectsMetricDescriptorsGetCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsMetricDescriptorsGetCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsMetricDescriptorsGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.metricDescriptors.get" call.
// Exactly one of *MetricDescriptor or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *MetricDescriptor.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsMetricDescriptorsGetCall) Do(opts ...googleapi.CallOption) (*MetricDescriptor, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MetricDescriptor{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a single metric descriptor. This method does not require a Stackdriver account.",
	//   "flatPath": "v3/projects/{projectsId}/metricDescriptors/{metricDescriptorsId}",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.metricDescriptors.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The metric descriptor on which to execute the request. The format is \"projects/{project_id_or_number}/metricDescriptors/{metric_id}\". An example value of {metric_id} is \"compute.googleapis.com/instance/disk/read_bytes_count\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/metricDescriptors/.+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "MetricDescriptor"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// method id "monitoring.projects.metricDescriptors.list":

type ProjectsMetricDescriptorsListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Lists metric descriptors that match a filter. This method does
// not require a Stackdriver account.
func (r *ProjectsMetricDescriptorsService) List(name string) *ProjectsMetricDescriptorsListCall {
	c := &ProjectsMetricDescriptorsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Filter sets the optional parameter "filter": If this field is empty,
// all custom and system-defined metric descriptors are returned.
// Otherwise, the filter specifies which metric descriptors are to be
// returned. For example, the following filter matches all custom
// metrics:
// metric.type = starts_with("custom.googleapis.com/")
func (c *ProjectsMetricDescriptorsListCall) Filter(filter string) *ProjectsMetricDescriptorsListCall {
	c.urlParams_.Set("filter", filter)
	return c
}

// PageSize sets the optional parameter "pageSize": A positive number
// that is the maximum number of results to return.
func (c *ProjectsMetricDescriptorsListCall) PageSize(pageSize int64) *ProjectsMetricDescriptorsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return additional results from the previous method call.
func (c *ProjectsMetricDescriptorsListCall) PageToken(pageToken string) *ProjectsMetricDescriptorsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsMetricDescriptorsListCall) Fields(s ...googleapi.Field) *ProjectsMetricDescriptorsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsMetricDescriptorsListCall) IfNoneMatch(entityTag string) *ProjectsMetricDescriptorsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsMetricDescriptorsListCall) Context(ctx context.Context) *ProjectsMetricDescriptorsListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsMetricDescriptorsListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsMetricDescriptorsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/metricDescriptors")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.metricDescriptors.list" call.
// Exactly one of *ListMetricDescriptorsResponse or error will be
// non-nil. Any non-2xx status code is an error. Response headers are in
// either *ListMetricDescriptorsResponse.ServerResponse.Header or (if a
// response was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsMetricDescriptorsListCall) Do(opts ...googleapi.CallOption) (*ListMetricDescriptorsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListMetricDescriptorsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists metric descriptors that match a filter. This method does not require a Stackdriver account.",
	//   "flatPath": "v3/projects/{projectsId}/metricDescriptors",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.metricDescriptors.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "filter": {
	//       "description": "If this field is empty, all custom and system-defined metric descriptors are returned. Otherwise, the filter specifies which metric descriptors are to be returned. For example, the following filter matches all custom metrics:\nmetric.type = starts_with(\"custom.googleapis.com/\")\n",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The project on which to execute the request. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "A positive number that is the maximum number of results to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return additional results from the previous method call.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/metricDescriptors",
	//   "response": {
	//     "$ref": "ListMetricDescriptorsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsMetricDescriptorsListCall) Pages(ctx context.Context, f func(*ListMetricDescriptorsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "monitoring.projects.monitoredResourceDescriptors.get":

type ProjectsMonitoredResourceDescriptorsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// Get: Gets a single monitored resource descriptor. This method does
// not require a Stackdriver account.
func (r *ProjectsMonitoredResourceDescriptorsService) Get(name string) *ProjectsMonitoredResourceDescriptorsGetCall {
	c := &ProjectsMonitoredResourceDescriptorsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsMonitoredResourceDescriptorsGetCall) Fields(s ...googleapi.Field) *ProjectsMonitoredResourceDescriptorsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsMonitoredResourceDescriptorsGetCall) IfNoneMatch(entityTag string) *ProjectsMonitoredResourceDescriptorsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsMonitoredResourceDescriptorsGetCall) Context(ctx context.Context) *ProjectsMonitoredResourceDescriptorsGetCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsMonitoredResourceDescriptorsGetCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsMonitoredResourceDescriptorsGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.monitoredResourceDescriptors.get" call.
// Exactly one of *MonitoredResourceDescriptor or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *MonitoredResourceDescriptor.ServerResponse.Header or (if a response
// was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsMonitoredResourceDescriptorsGetCall) Do(opts ...googleapi.CallOption) (*MonitoredResourceDescriptor, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &MonitoredResourceDescriptor{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a single monitored resource descriptor. This method does not require a Stackdriver account.",
	//   "flatPath": "v3/projects/{projectsId}/monitoredResourceDescriptors/{monitoredResourceDescriptorsId}",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.monitoredResourceDescriptors.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The monitored resource descriptor to get. The format is \"projects/{project_id_or_number}/monitoredResourceDescriptors/{resource_type}\". The {resource_type} is a predefined type, such as cloudsql_database.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/monitoredResourceDescriptors/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "MonitoredResourceDescriptor"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// method id "monitoring.projects.monitoredResourceDescriptors.list":

type ProjectsMonitoredResourceDescriptorsListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Lists monitored resource descriptors that match a filter. This
// method does not require a Stackdriver account.
func (r *ProjectsMonitoredResourceDescriptorsService) List(name string) *ProjectsMonitoredResourceDescriptorsListCall {
	c := &ProjectsMonitoredResourceDescriptorsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Filter sets the optional parameter "filter": An optional filter
// describing the descriptors to be returned. The filter can reference
// the descriptor's type and labels. For example, the following filter
// returns only Google Compute Engine descriptors that have an id
// label:
// resource.type = starts_with("gce_") AND resource.label:id
func (c *ProjectsMonitoredResourceDescriptorsListCall) Filter(filter string) *ProjectsMonitoredResourceDescriptorsListCall {
	c.urlParams_.Set("filter", filter)
	return c
}

// PageSize sets the optional parameter "pageSize": A positive number
// that is the maximum number of results to return.
func (c *ProjectsMonitoredResourceDescriptorsListCall) PageSize(pageSize int64) *ProjectsMonitoredResourceDescriptorsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return additional results from the previous method call.
func (c *ProjectsMonitoredResourceDescriptorsListCall) PageToken(pageToken string) *ProjectsMonitoredResourceDescriptorsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsMonitoredResourceDescriptorsListCall) Fields(s ...googleapi.Field) *ProjectsMonitoredResourceDescriptorsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsMonitoredResourceDescriptorsListCall) IfNoneMatch(entityTag string) *ProjectsMonitoredResourceDescriptorsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsMonitoredResourceDescriptorsListCall) Context(ctx context.Context) *ProjectsMonitoredResourceDescriptorsListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsMonitoredResourceDescriptorsListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsMonitoredResourceDescriptorsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/monitoredResourceDescriptors")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.monitoredResourceDescriptors.list" call.
// Exactly one of *ListMonitoredResourceDescriptorsResponse or error
// will be non-nil. Any non-2xx status code is an error. Response
// headers are in either
// *ListMonitoredResourceDescriptorsResponse.ServerResponse.Header or
// (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsMonitoredResourceDescriptorsListCall) Do(opts ...googleapi.CallOption) (*ListMonitoredResourceDescriptorsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListMonitoredResourceDescriptorsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists monitored resource descriptors that match a filter. This method does not require a Stackdriver account.",
	//   "flatPath": "v3/projects/{projectsId}/monitoredResourceDescriptors",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.monitoredResourceDescriptors.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "filter": {
	//       "description": "An optional filter describing the descriptors to be returned. The filter can reference the descriptor's type and labels. For example, the following filter returns only Google Compute Engine descriptors that have an id label:\nresource.type = starts_with(\"gce_\") AND resource.label:id\n",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The project on which to execute the request. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "A positive number that is the maximum number of results to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return additional results from the previous method call.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/monitoredResourceDescriptors",
	//   "response": {
	//     "$ref": "ListMonitoredResourceDescriptorsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsMonitoredResourceDescriptorsListCall) Pages(ctx context.Context, f func(*ListMonitoredResourceDescriptorsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "monitoring.projects.timeSeries.create":

type ProjectsTimeSeriesCreateCall struct {
	s                       *Service
	name                    string
	createtimeseriesrequest *CreateTimeSeriesRequest
	urlParams_              gensupport.URLParams
	ctx_                    context.Context
	header_                 http.Header
}

// Create: Creates or adds data to one or more time series. The response
// is empty if all time series in the request were written. If any time
// series could not be written, a corresponding failure message is
// included in the error response.
func (r *ProjectsTimeSeriesService) Create(name string, createtimeseriesrequest *CreateTimeSeriesRequest) *ProjectsTimeSeriesCreateCall {
	c := &ProjectsTimeSeriesCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.createtimeseriesrequest = createtimeseriesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsTimeSeriesCreateCall) Fields(s ...googleapi.Field) *ProjectsTimeSeriesCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsTimeSeriesCreateCall) Context(ctx context.Context) *ProjectsTimeSeriesCreateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsTimeSeriesCreateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsTimeSeriesCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createtimeseriesrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/timeSeries")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.timeSeries.create" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsTimeSeriesCreateCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates or adds data to one or more time series. The response is empty if all time series in the request were written. If any time series could not be written, a corresponding failure message is included in the error response.",
	//   "flatPath": "v3/projects/{projectsId}/timeSeries",
	//   "httpMethod": "POST",
	//   "id": "monitoring.projects.timeSeries.create",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The project on which to execute the request. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/timeSeries",
	//   "request": {
	//     "$ref": "CreateTimeSeriesRequest"
	//   },
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.write"
	//   ]
	// }

}

// method id "monitoring.projects.timeSeries.list":

type ProjectsTimeSeriesListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Lists time series that match a filter. This method does not
// require a Stackdriver account.
func (r *ProjectsTimeSeriesService) List(name string) *ProjectsTimeSeriesListCall {
	c := &ProjectsTimeSeriesListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// AggregationAlignmentPeriod sets the optional parameter
// "aggregation.alignmentPeriod": The alignment period for per-time
// series alignment. If present, alignmentPeriod must be at least 60
// seconds. After per-time series alignment, each time series will
// contain data points only on the period boundaries. If
// perSeriesAligner is not specified or equals ALIGN_NONE, then this
// field is ignored. If perSeriesAligner is specified and does not equal
// ALIGN_NONE, then this field must be defined; otherwise an error is
// returned.
func (c *ProjectsTimeSeriesListCall) AggregationAlignmentPeriod(aggregationAlignmentPeriod string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("aggregation.alignmentPeriod", aggregationAlignmentPeriod)
	return c
}

// AggregationCrossSeriesReducer sets the optional parameter
// "aggregation.crossSeriesReducer": The approach to be used to combine
// time series. Not all reducer functions may be applied to all time
// series, depending on the metric type and the value type of the
// original time series. Reduction may change the metric type of value
// type of the time series.Time series data must be aligned in order to
// perform cross-time series reduction. If crossSeriesReducer is
// specified, then perSeriesAligner must be specified and not equal
// ALIGN_NONE and alignmentPeriod must be specified; otherwise, an error
// is returned.
//
// Possible values:
//   "REDUCE_NONE"
//   "REDUCE_MEAN"
//   "REDUCE_MIN"
//   "REDUCE_MAX"
//   "REDUCE_SUM"
//   "REDUCE_STDDEV"
//   "REDUCE_COUNT"
//   "REDUCE_COUNT_TRUE"
//   "REDUCE_COUNT_FALSE"
//   "REDUCE_FRACTION_TRUE"
//   "REDUCE_PERCENTILE_99"
//   "REDUCE_PERCENTILE_95"
//   "REDUCE_PERCENTILE_50"
//   "REDUCE_PERCENTILE_05"
func (c *ProjectsTimeSeriesListCall) AggregationCrossSeriesReducer(aggregationCrossSeriesReducer string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("aggregation.crossSeriesReducer", aggregationCrossSeriesReducer)
	return c
}

// AggregationGroupByFields sets the optional parameter
// "aggregation.groupByFields": The set of fields to preserve when
// crossSeriesReducer is specified. The groupByFields determine how the
// time series are partitioned into subsets prior to applying the
// aggregation function. Each subset contains time series that have the
// same value for each of the grouping fields. Each individual time
// series is a member of exactly one subset. The crossSeriesReducer is
// applied to each subset of time series. It is not possible to reduce
// across different resource types, so this field implicitly contains
// resource.type. Fields not specified in groupByFields are aggregated
// away. If groupByFields is not specified and all the time series have
// the same resource type, then the time series are aggregated into a
// single output time series. If crossSeriesReducer is not defined, this
// field is ignored.
func (c *ProjectsTimeSeriesListCall) AggregationGroupByFields(aggregationGroupByFields ...string) *ProjectsTimeSeriesListCall {
	c.urlParams_.SetMulti("aggregation.groupByFields", append([]string{}, aggregationGroupByFields...))
	return c
}

// AggregationPerSeriesAligner sets the optional parameter
// "aggregation.perSeriesAligner": The approach to be used to align
// individual time series. Not all alignment functions may be applied to
// all time series, depending on the metric type and value type of the
// original time series. Alignment may change the metric type or the
// value type of the time series.Time series data must be aligned in
// order to perform cross-time series reduction. If crossSeriesReducer
// is specified, then perSeriesAligner must be specified and not equal
// ALIGN_NONE and alignmentPeriod must be specified; otherwise, an error
// is returned.
//
// Possible values:
//   "ALIGN_NONE"
//   "ALIGN_DELTA"
//   "ALIGN_RATE"
//   "ALIGN_INTERPOLATE"
//   "ALIGN_NEXT_OLDER"
//   "ALIGN_MIN"
//   "ALIGN_MAX"
//   "ALIGN_MEAN"
//   "ALIGN_COUNT"
//   "ALIGN_SUM"
//   "ALIGN_STDDEV"
//   "ALIGN_COUNT_TRUE"
//   "ALIGN_COUNT_FALSE"
//   "ALIGN_FRACTION_TRUE"
//   "ALIGN_PERCENTILE_99"
//   "ALIGN_PERCENTILE_95"
//   "ALIGN_PERCENTILE_50"
//   "ALIGN_PERCENTILE_05"
//   "ALIGN_PERCENT_CHANGE"
func (c *ProjectsTimeSeriesListCall) AggregationPerSeriesAligner(aggregationPerSeriesAligner string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("aggregation.perSeriesAligner", aggregationPerSeriesAligner)
	return c
}

// Filter sets the optional parameter "filter": A monitoring filter that
// specifies which time series should be returned. The filter must
// specify a single metric type, and can additionally specify metric
// labels and other information. For example:
// metric.type = "compute.googleapis.com/instance/cpu/usage_time" AND
//     metric.label.instance_name = "my-instance-name"
func (c *ProjectsTimeSeriesListCall) Filter(filter string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("filter", filter)
	return c
}

// IntervalEndTime sets the optional parameter "interval.endTime":
// Required. The end of the time interval.
func (c *ProjectsTimeSeriesListCall) IntervalEndTime(intervalEndTime string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("interval.endTime", intervalEndTime)
	return c
}

// IntervalStartTime sets the optional parameter "interval.startTime":
// The beginning of the time interval. The default value for the start
// time is the end time. The start time must not be later than the end
// time.
func (c *ProjectsTimeSeriesListCall) IntervalStartTime(intervalStartTime string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("interval.startTime", intervalStartTime)
	return c
}

// OrderBy sets the optional parameter "orderBy": Unsupported: must be
// left blank. The points in each time series are returned in reverse
// time order.
func (c *ProjectsTimeSeriesListCall) OrderBy(orderBy string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("orderBy", orderBy)
	return c
}

// PageSize sets the optional parameter "pageSize": A positive number
// that is the maximum number of results to return. When view field sets
// to FULL, it limits the number of Points server will return; if view
// field is HEADERS, it limits the number of TimeSeries server will
// return.
func (c *ProjectsTimeSeriesListCall) PageSize(pageSize int64) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return additional results from the previous method call.
func (c *ProjectsTimeSeriesListCall) PageToken(pageToken string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// View sets the optional parameter "view": Specifies which information
// is returned about the time series.
//
// Possible values:
//   "FULL"
//   "HEADERS"
func (c *ProjectsTimeSeriesListCall) View(view string) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("view", view)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsTimeSeriesListCall) Fields(s ...googleapi.Field) *ProjectsTimeSeriesListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsTimeSeriesListCall) IfNoneMatch(entityTag string) *ProjectsTimeSeriesListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsTimeSeriesListCall) Context(ctx context.Context) *ProjectsTimeSeriesListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsTimeSeriesListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsTimeSeriesListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}/timeSeries")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.timeSeries.list" call.
// Exactly one of *ListTimeSeriesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListTimeSeriesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsTimeSeriesListCall) Do(opts ...googleapi.CallOption) (*ListTimeSeriesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListTimeSeriesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists time series that match a filter. This method does not require a Stackdriver account.",
	//   "flatPath": "v3/projects/{projectsId}/timeSeries",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.timeSeries.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "aggregation.alignmentPeriod": {
	//       "description": "The alignment period for per-time series alignment. If present, alignmentPeriod must be at least 60 seconds. After per-time series alignment, each time series will contain data points only on the period boundaries. If perSeriesAligner is not specified or equals ALIGN_NONE, then this field is ignored. If perSeriesAligner is specified and does not equal ALIGN_NONE, then this field must be defined; otherwise an error is returned.",
	//       "format": "google-duration",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "aggregation.crossSeriesReducer": {
	//       "description": "The approach to be used to combine time series. Not all reducer functions may be applied to all time series, depending on the metric type and the value type of the original time series. Reduction may change the metric type of value type of the time series.Time series data must be aligned in order to perform cross-time series reduction. If crossSeriesReducer is specified, then perSeriesAligner must be specified and not equal ALIGN_NONE and alignmentPeriod must be specified; otherwise, an error is returned.",
	//       "enum": [
	//         "REDUCE_NONE",
	//         "REDUCE_MEAN",
	//         "REDUCE_MIN",
	//         "REDUCE_MAX",
	//         "REDUCE_SUM",
	//         "REDUCE_STDDEV",
	//         "REDUCE_COUNT",
	//         "REDUCE_COUNT_TRUE",
	//         "REDUCE_COUNT_FALSE",
	//         "REDUCE_FRACTION_TRUE",
	//         "REDUCE_PERCENTILE_99",
	//         "REDUCE_PERCENTILE_95",
	//         "REDUCE_PERCENTILE_50",
	//         "REDUCE_PERCENTILE_05"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "aggregation.groupByFields": {
	//       "description": "The set of fields to preserve when crossSeriesReducer is specified. The groupByFields determine how the time series are partitioned into subsets prior to applying the aggregation function. Each subset contains time series that have the same value for each of the grouping fields. Each individual time series is a member of exactly one subset. The crossSeriesReducer is applied to each subset of time series. It is not possible to reduce across different resource types, so this field implicitly contains resource.type. Fields not specified in groupByFields are aggregated away. If groupByFields is not specified and all the time series have the same resource type, then the time series are aggregated into a single output time series. If crossSeriesReducer is not defined, this field is ignored.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "aggregation.perSeriesAligner": {
	//       "description": "The approach to be used to align individual time series. Not all alignment functions may be applied to all time series, depending on the metric type and value type of the original time series. Alignment may change the metric type or the value type of the time series.Time series data must be aligned in order to perform cross-time series reduction. If crossSeriesReducer is specified, then perSeriesAligner must be specified and not equal ALIGN_NONE and alignmentPeriod must be specified; otherwise, an error is returned.",
	//       "enum": [
	//         "ALIGN_NONE",
	//         "ALIGN_DELTA",
	//         "ALIGN_RATE",
	//         "ALIGN_INTERPOLATE",
	//         "ALIGN_NEXT_OLDER",
	//         "ALIGN_MIN",
	//         "ALIGN_MAX",
	//         "ALIGN_MEAN",
	//         "ALIGN_COUNT",
	//         "ALIGN_SUM",
	//         "ALIGN_STDDEV",
	//         "ALIGN_COUNT_TRUE",
	//         "ALIGN_COUNT_FALSE",
	//         "ALIGN_FRACTION_TRUE",
	//         "ALIGN_PERCENTILE_99",
	//         "ALIGN_PERCENTILE_95",
	//         "ALIGN_PERCENTILE_50",
	//         "ALIGN_PERCENTILE_05",
	//         "ALIGN_PERCENT_CHANGE"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "filter": {
	//       "description": "A monitoring filter that specifies which time series should be returned. The filter must specify a single metric type, and can additionally specify metric labels and other information. For example:\nmetric.type = \"compute.googleapis.com/instance/cpu/usage_time\" AND\n    metric.label.instance_name = \"my-instance-name\"\n",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "interval.endTime": {
	//       "description": "Required. The end of the time interval.",
	//       "format": "google-datetime",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "interval.startTime": {
	//       "description": "Optional. The beginning of the time interval. The default value for the start time is the end time. The start time must not be later than the end time.",
	//       "format": "google-datetime",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The project on which to execute the request. The format is \"projects/{project_id_or_number}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "orderBy": {
	//       "description": "Unsupported: must be left blank. The points in each time series are returned in reverse time order.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "A positive number that is the maximum number of results to return. When view field sets to FULL, it limits the number of Points server will return; if view field is HEADERS, it limits the number of TimeSeries server will return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return additional results from the previous method call.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "view": {
	//       "description": "Specifies which information is returned about the time series.",
	//       "enum": [
	//         "FULL",
	//         "HEADERS"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}/timeSeries",
	//   "response": {
	//     "$ref": "ListTimeSeriesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsTimeSeriesListCall) Pages(ctx context.Context, f func(*ListTimeSeriesResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "monitoring.projects.uptimeCheckConfigs.create":

type ProjectsUptimeCheckConfigsCreateCall struct {
	s                 *Service
	parent            string
	uptimecheckconfig *UptimeCheckConfig
	urlParams_        gensupport.URLParams
	ctx_              context.Context
	header_           http.Header
}

// Create: Creates a new uptime check configuration.
func (r *ProjectsUptimeCheckConfigsService) Create(parent string, uptimecheckconfig *UptimeCheckConfig) *ProjectsUptimeCheckConfigsCreateCall {
	c := &ProjectsUptimeCheckConfigsCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.parent = parent
	c.uptimecheckconfig = uptimecheckconfig
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsUptimeCheckConfigsCreateCall) Fields(s ...googleapi.Field) *ProjectsUptimeCheckConfigsCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsUptimeCheckConfigsCreateCall) Context(ctx context.Context) *ProjectsUptimeCheckConfigsCreateCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsUptimeCheckConfigsCreateCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsUptimeCheckConfigsCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.uptimecheckconfig)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+parent}/uptimeCheckConfigs")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"parent": c.parent,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.uptimeCheckConfigs.create" call.
// Exactly one of *UptimeCheckConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *UptimeCheckConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsUptimeCheckConfigsCreateCall) Do(opts ...googleapi.CallOption) (*UptimeCheckConfig, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &UptimeCheckConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a new uptime check configuration.",
	//   "flatPath": "v3/projects/{projectsId}/uptimeCheckConfigs",
	//   "httpMethod": "POST",
	//   "id": "monitoring.projects.uptimeCheckConfigs.create",
	//   "parameterOrder": [
	//     "parent"
	//   ],
	//   "parameters": {
	//     "parent": {
	//       "description": "The project in which to create the uptime check. The format  is projects/[PROJECT_ID].",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+parent}/uptimeCheckConfigs",
	//   "request": {
	//     "$ref": "UptimeCheckConfig"
	//   },
	//   "response": {
	//     "$ref": "UptimeCheckConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.projects.uptimeCheckConfigs.delete":

type ProjectsUptimeCheckConfigsDeleteCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
	header_    http.Header
}

// Delete: Deletes an uptime check configuration. Note that this method
// will fail if the uptime check configuration is referenced by an alert
// policy or other dependent configs that would be rendered invalid by
// the deletion.
func (r *ProjectsUptimeCheckConfigsService) Delete(name string) *ProjectsUptimeCheckConfigsDeleteCall {
	c := &ProjectsUptimeCheckConfigsDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsUptimeCheckConfigsDeleteCall) Fields(s ...googleapi.Field) *ProjectsUptimeCheckConfigsDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsUptimeCheckConfigsDeleteCall) Context(ctx context.Context) *ProjectsUptimeCheckConfigsDeleteCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsUptimeCheckConfigsDeleteCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsUptimeCheckConfigsDeleteCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.uptimeCheckConfigs.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsUptimeCheckConfigsDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Deletes an uptime check configuration. Note that this method will fail if the uptime check configuration is referenced by an alert policy or other dependent configs that would be rendered invalid by the deletion.",
	//   "flatPath": "v3/projects/{projectsId}/uptimeCheckConfigs/{uptimeCheckConfigsId}",
	//   "httpMethod": "DELETE",
	//   "id": "monitoring.projects.uptimeCheckConfigs.delete",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The uptime check configuration to delete. The format  is projects/[PROJECT_ID]/uptimeCheckConfigs/[UPTIME_CHECK_ID].",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/uptimeCheckConfigs/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.projects.uptimeCheckConfigs.get":

type ProjectsUptimeCheckConfigsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// Get: Gets a single uptime check configuration.
func (r *ProjectsUptimeCheckConfigsService) Get(name string) *ProjectsUptimeCheckConfigsGetCall {
	c := &ProjectsUptimeCheckConfigsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsUptimeCheckConfigsGetCall) Fields(s ...googleapi.Field) *ProjectsUptimeCheckConfigsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsUptimeCheckConfigsGetCall) IfNoneMatch(entityTag string) *ProjectsUptimeCheckConfigsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsUptimeCheckConfigsGetCall) Context(ctx context.Context) *ProjectsUptimeCheckConfigsGetCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsUptimeCheckConfigsGetCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsUptimeCheckConfigsGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.uptimeCheckConfigs.get" call.
// Exactly one of *UptimeCheckConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *UptimeCheckConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsUptimeCheckConfigsGetCall) Do(opts ...googleapi.CallOption) (*UptimeCheckConfig, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &UptimeCheckConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a single uptime check configuration.",
	//   "flatPath": "v3/projects/{projectsId}/uptimeCheckConfigs/{uptimeCheckConfigsId}",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.uptimeCheckConfigs.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The uptime check configuration to retrieve. The format  is projects/[PROJECT_ID]/uptimeCheckConfigs/[UPTIME_CHECK_ID].",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/uptimeCheckConfigs/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "response": {
	//     "$ref": "UptimeCheckConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// method id "monitoring.projects.uptimeCheckConfigs.list":

type ProjectsUptimeCheckConfigsListCall struct {
	s            *Service
	parent       string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Lists the existing valid uptime check configurations for the
// project, leaving out any invalid configurations.
func (r *ProjectsUptimeCheckConfigsService) List(parent string) *ProjectsUptimeCheckConfigsListCall {
	c := &ProjectsUptimeCheckConfigsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.parent = parent
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of results to return in a single response. The server may further
// constrain the maximum number of results returned in a single page. If
// the page_size is <=0, the server will decide the number of results to
// be returned.
func (c *ProjectsUptimeCheckConfigsListCall) PageSize(pageSize int64) *ProjectsUptimeCheckConfigsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return more results from the previous method call.
func (c *ProjectsUptimeCheckConfigsListCall) PageToken(pageToken string) *ProjectsUptimeCheckConfigsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsUptimeCheckConfigsListCall) Fields(s ...googleapi.Field) *ProjectsUptimeCheckConfigsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsUptimeCheckConfigsListCall) IfNoneMatch(entityTag string) *ProjectsUptimeCheckConfigsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsUptimeCheckConfigsListCall) Context(ctx context.Context) *ProjectsUptimeCheckConfigsListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsUptimeCheckConfigsListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsUptimeCheckConfigsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+parent}/uptimeCheckConfigs")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"parent": c.parent,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.uptimeCheckConfigs.list" call.
// Exactly one of *ListUptimeCheckConfigsResponse or error will be
// non-nil. Any non-2xx status code is an error. Response headers are in
// either *ListUptimeCheckConfigsResponse.ServerResponse.Header or (if a
// response was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsUptimeCheckConfigsListCall) Do(opts ...googleapi.CallOption) (*ListUptimeCheckConfigsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListUptimeCheckConfigsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists the existing valid uptime check configurations for the project, leaving out any invalid configurations.",
	//   "flatPath": "v3/projects/{projectsId}/uptimeCheckConfigs",
	//   "httpMethod": "GET",
	//   "id": "monitoring.projects.uptimeCheckConfigs.list",
	//   "parameterOrder": [
	//     "parent"
	//   ],
	//   "parameters": {
	//     "pageSize": {
	//       "description": "The maximum number of results to return in a single response. The server may further constrain the maximum number of results returned in a single page. If the page_size is \u003c=0, the server will decide the number of results to be returned.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return more results from the previous method call.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "parent": {
	//       "description": "The project whose uptime check configurations are listed. The format  is projects/[PROJECT_ID].",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+parent}/uptimeCheckConfigs",
	//   "response": {
	//     "$ref": "ListUptimeCheckConfigsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsUptimeCheckConfigsListCall) Pages(ctx context.Context, f func(*ListUptimeCheckConfigsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "monitoring.projects.uptimeCheckConfigs.patch":

type ProjectsUptimeCheckConfigsPatchCall struct {
	s                 *Service
	name              string
	uptimecheckconfig *UptimeCheckConfig
	urlParams_        gensupport.URLParams
	ctx_              context.Context
	header_           http.Header
}

// Patch: Updates an uptime check configuration. You can either replace
// the entire configuration with a new one or replace only certain
// fields in the current configuration by specifying the fields to be
// updated via "updateMask". Returns the updated configuration.
func (r *ProjectsUptimeCheckConfigsService) Patch(name string, uptimecheckconfig *UptimeCheckConfig) *ProjectsUptimeCheckConfigsPatchCall {
	c := &ProjectsUptimeCheckConfigsPatchCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.uptimecheckconfig = uptimecheckconfig
	return c
}

// UpdateMask sets the optional parameter "updateMask": If present, only
// the listed fields in the current uptime check configuration are
// updated with values from the new configuration. If this field is
// empty, then the current configuration is completely replaced with the
// new configuration.
func (c *ProjectsUptimeCheckConfigsPatchCall) UpdateMask(updateMask string) *ProjectsUptimeCheckConfigsPatchCall {
	c.urlParams_.Set("updateMask", updateMask)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsUptimeCheckConfigsPatchCall) Fields(s ...googleapi.Field) *ProjectsUptimeCheckConfigsPatchCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsUptimeCheckConfigsPatchCall) Context(ctx context.Context) *ProjectsUptimeCheckConfigsPatchCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *ProjectsUptimeCheckConfigsPatchCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *ProjectsUptimeCheckConfigsPatchCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.uptimecheckconfig)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.projects.uptimeCheckConfigs.patch" call.
// Exactly one of *UptimeCheckConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *UptimeCheckConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsUptimeCheckConfigsPatchCall) Do(opts ...googleapi.CallOption) (*UptimeCheckConfig, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &UptimeCheckConfig{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an uptime check configuration. You can either replace the entire configuration with a new one or replace only certain fields in the current configuration by specifying the fields to be updated via \"updateMask\". Returns the updated configuration.",
	//   "flatPath": "v3/projects/{projectsId}/uptimeCheckConfigs/{uptimeCheckConfigsId}",
	//   "httpMethod": "PATCH",
	//   "id": "monitoring.projects.uptimeCheckConfigs.patch",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "A unique resource name for this UptimeCheckConfig. The format is:projects/[PROJECT_ID]/uptimeCheckConfigs/[UPTIME_CHECK_ID].This field should be omitted when creating the uptime check configuration; on create, the resource name is assigned by the server and included in the response.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]+/uptimeCheckConfigs/[^/]+$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "updateMask": {
	//       "description": "Optional. If present, only the listed fields in the current uptime check configuration are updated with values from the new configuration. If this field is empty, then the current configuration is completely replaced with the new configuration.",
	//       "format": "google-fieldmask",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/{+name}",
	//   "request": {
	//     "$ref": "UptimeCheckConfig"
	//   },
	//   "response": {
	//     "$ref": "UptimeCheckConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring"
	//   ]
	// }

}

// method id "monitoring.uptimeCheckIps.list":

type UptimeCheckIpsListCall struct {
	s            *Service
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
	header_      http.Header
}

// List: Returns the list of IPs that checkers run from
func (r *UptimeCheckIpsService) List() *UptimeCheckIpsListCall {
	c := &UptimeCheckIpsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of results to return in a single response. The server may further
// constrain the maximum number of results returned in a single page. If
// the page_size is <=0, the server will decide the number of results to
// be returned. NOTE: this field is not yet implemented
func (c *UptimeCheckIpsListCall) PageSize(pageSize int64) *UptimeCheckIpsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": If this field is
// not empty then it must contain the nextPageToken value returned by a
// previous call to this method. Using this field causes the method to
// return more results from the previous method call. NOTE: this field
// is not yet implemented
func (c *UptimeCheckIpsListCall) PageToken(pageToken string) *UptimeCheckIpsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *UptimeCheckIpsListCall) Fields(s ...googleapi.Field) *UptimeCheckIpsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *UptimeCheckIpsListCall) IfNoneMatch(entityTag string) *UptimeCheckIpsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *UptimeCheckIpsListCall) Context(ctx context.Context) *UptimeCheckIpsListCall {
	c.ctx_ = ctx
	return c
}

// Header returns an http.Header that can be modified by the caller to
// add HTTP headers to the request.
func (c *UptimeCheckIpsListCall) Header() http.Header {
	if c.header_ == nil {
		c.header_ = make(http.Header)
	}
	return c.header_
}

func (c *UptimeCheckIpsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	for k, v := range c.header_ {
		reqHeaders[k] = v
	}
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v3/uptimeCheckIps")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "monitoring.uptimeCheckIps.list" call.
// Exactly one of *ListUptimeCheckIpsResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ListUptimeCheckIpsResponse.ServerResponse.Header or (if a response
// was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *UptimeCheckIpsListCall) Do(opts ...googleapi.CallOption) (*ListUptimeCheckIpsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListUptimeCheckIpsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := gensupport.DecodeResponse(target, res); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns the list of IPs that checkers run from",
	//   "flatPath": "v3/uptimeCheckIps",
	//   "httpMethod": "GET",
	//   "id": "monitoring.uptimeCheckIps.list",
	//   "parameterOrder": [],
	//   "parameters": {
	//     "pageSize": {
	//       "description": "The maximum number of results to return in a single response. The server may further constrain the maximum number of results returned in a single page. If the page_size is \u003c=0, the server will decide the number of results to be returned. NOTE: this field is not yet implemented",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "If this field is not empty then it must contain the nextPageToken value returned by a previous call to this method. Using this field causes the method to return more results from the previous method call. NOTE: this field is not yet implemented",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v3/uptimeCheckIps",
	//   "response": {
	//     "$ref": "ListUptimeCheckIpsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/monitoring",
	//     "https://www.googleapis.com/auth/monitoring.read"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *UptimeCheckIpsListCall) Pages(ctx context.Context, f func(*ListUptimeCheckIpsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}
