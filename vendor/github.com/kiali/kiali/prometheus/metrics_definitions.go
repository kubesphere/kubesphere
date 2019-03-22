package prometheus

type istioMetric struct {
	kialiName      string
	istioName      string
	isHisto        bool
	useErrorLabels bool
}

var istioMetrics = []istioMetric{
	istioMetric{
		kialiName: "request_count",
		istioName: "istio_requests_total",
		isHisto:   false,
	},
	istioMetric{
		kialiName:      "request_error_count",
		istioName:      "istio_requests_total",
		isHisto:        false,
		useErrorLabels: true,
	},
	istioMetric{
		kialiName: "request_duration",
		istioName: "istio_request_duration_seconds",
		isHisto:   true,
	},
	istioMetric{
		kialiName: "request_size",
		istioName: "istio_request_bytes",
		isHisto:   true,
	},
	istioMetric{
		kialiName: "response_size",
		istioName: "istio_response_bytes",
		isHisto:   true,
	},
	istioMetric{
		kialiName: "tcp_received",
		istioName: "istio_tcp_received_bytes_total",
		isHisto:   false,
	},
	istioMetric{
		kialiName: "tcp_sent",
		istioName: "istio_tcp_sent_bytes_total",
		isHisto:   false,
	},
}

func (in *istioMetric) labelsToUse(labels, labelsError string) string {
	if in.useErrorLabels {
		return labelsError
	}
	return labels
}
