package notification

type Options struct {
	Endpoint string
}

func NewNotificationOptions() *Options {
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
