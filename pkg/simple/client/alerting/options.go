package alerting

type AlertingOptions struct {
	Endpoint string
}

func NewAlertingOptions() *AlertingOptions {
	return &AlertingOptions{
		Endpoint: "",
	}
}

func (s *AlertingOptions) ApplyTo(options *AlertingOptions) {
	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}
