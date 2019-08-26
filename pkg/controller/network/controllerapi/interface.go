package controllerapi

// Controller expose Run method
type Controller interface {
	Run(threadiness int, stopCh <-chan struct{}) error
}
