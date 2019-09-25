package options

func (s *ServerRunOptions) Validate() []error {
	errs := []error{}

	errs = append(errs, s.KubernetesOptions.Validate()...)
	errs = append(errs, s.GenericServerRunOptions.Validate()...)
	errs = append(errs, s.LdapOptions.Validate()...)

	return errs
}
