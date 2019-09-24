package alerting

type AlertingOptions struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

func NewAlertingOptions() *AlertingOptions {
	return &AlertingOptions{
		Endpoint: "",
	}
}

func (s *AlertingOptions) ApplyTo(options *AlertingOptions) {
	if options == nil {
		options = s
		return
	}

	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}
