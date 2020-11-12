package checkers

import "kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/models"

type Checker interface {
	Check() ([]*models.IstioCheck, bool)
}

type GroupChecker interface {
	Check() models.IstioValidations
}
