package notification

type NotificationOptions struct {
	Endpoint string
}

func NewNotificationOptions() *NotificationOptions {
	return &NotificationOptions{
		Endpoint: "",
	}
}

func (s *NotificationOptions) ApplyTo(options *NotificationOptions) {
	if options == nil {
		options = s
		return
	}

	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}
