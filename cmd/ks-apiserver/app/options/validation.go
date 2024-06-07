/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package options

// Validate validates server run options, to find
// options' misconfiguration
func (s *APIServerOptions) Validate() []error {
	var errors []error
	errors = append(errors, s.GenericServerRunOptions.Validate()...)
	errors = append(errors, s.KubernetesOptions.Validate()...)
	errors = append(errors, s.AuthenticationOptions.Validate()...)
	errors = append(errors, s.AuthorizationOptions.Validate()...)
	errors = append(errors, s.AuditingOptions.Validate()...)
	return errors
}
