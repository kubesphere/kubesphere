package options

// Validate validates server run options, to find
// options' misconfiguration
func (s *ServerRunOptions) Validate() []error {
	var errors []error

	errors = append(errors, s.GenericServerRunOptions.Validate()...)
	errors = append(errors, s.DevopsOptions.Validate()...)
	errors = append(errors, s.KubernetesOptions.Validate()...)
	errors = append(errors, s.ServiceMeshOptions.Validate()...)
	errors = append(errors, s.MonitoringOptions.Validate()...)
	errors = append(errors, s.SonarQubeOptions.Validate()...)
	errors = append(errors, s.S3Options.Validate()...)
	errors = append(errors, s.OpenPitrixOptions.Validate()...)
	errors = append(errors, s.NetworkOptions.Validate()...)
	errors = append(errors, s.LoggingOptions.Validate()...)
	errors = append(errors, s.AuthorizationOptions.Validate()...)
	errors = append(errors, s.EventsOptions.Validate()...)
	errors = append(errors, s.AuditingOptions.Validate()...)

	return errors
}
