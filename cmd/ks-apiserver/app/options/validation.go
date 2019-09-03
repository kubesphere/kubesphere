package options

func (s *ServerRunOptions) Validate() []error {
	var errors []error

	errors = append(errors, s.DevopsOptions.Validate()...)
	errors = append(errors, s.KubernetesOptions.Validate()...)
	errors = append(errors, s.MySQLOptions.Validate()...)
	errors = append(errors, s.ServiceMeshOptions.Validate()...)
	errors = append(errors, s.MonitoringOptions.Validate()...)
	errors = append(errors, s.SonarQubeOptions.Validate()...)

	return errors
}
