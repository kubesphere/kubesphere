package k8s

import (
	monit "github.com/coreos/prometheus-operator/pkg/client/versioned"
	"log"
	"sync"
)

var (
	monitClient     *monit.Clientset
	monitClientOnce sync.Once
)

func MonitClient() *monit.Clientset {

	monitClientOnce.Do(func() {

		config, err := Config()

		if err != nil {
			log.Fatalln(err)
		}

		monitClient = monit.NewForConfigOrDie(config)
	})

	return monitClient
}
