package options

// Validate validates server run options, to find
// options' misconfiguration
func (s *ServerRunOptions) Validate() []error {
	var errors []error

	if s.DevopsOptions != nil {
		errors = append(errors, s.DevopsOptions.Validate()...)
	}
	if s.KubernetesOptions != nil {
		errors = append(errors, s.KubernetesOptions.Validate()...)
	}
	if s.MySQLOptions != nil {
		errors = append(errors, s.MySQLOptions.Validate()...)
	}
	if s.ServiceMeshOptions != nil {
		errors = append(errors, s.ServiceMeshOptions.Validate()...)
	}
	if s.MonitoringOptions != nil {
		errors = append(errors, s.MonitoringOptions.Validate()...)
	}
	if s.SonarQubeOptions != nil {
		errors = append(errors, s.SonarQubeOptions.Validate()...)
	}
	if s.LdapOptions != nil {
		errors = append(errors, s.LdapOptions.Validate()...)
	}
	if s.S3Options != nil {
		errors = append(errors, s.S3Options.Validate()...)
	}
	if s.RedisOptions != nil {
		errors = append(errors, s.RedisOptions.Validate()...)
	}
	if s.OpenPitrixOptions != nil {
		errors = append(errors, s.OpenPitrixOptions.Validate()...)
	}
	if s.LoggingOptions != nil {
		errors = append(errors, s.LoggingOptions.Validate()...)
	}
	return errors
}
