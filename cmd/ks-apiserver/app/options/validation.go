/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
