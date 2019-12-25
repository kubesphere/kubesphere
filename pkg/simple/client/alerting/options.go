package alerting

type Options struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

func NewAlertingOptions() *Options {
	return &Options{
		Endpoint: "",
	}
}

func (s *Options) ApplyTo(options *Options) {
	if options == nil {
		options = s
		return
	}

	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}
