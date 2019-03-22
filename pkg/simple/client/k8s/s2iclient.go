package k8s

import (
	"log"
	"sync"

	s2i "github.com/kubesphere/s2ioperator/pkg/client/clientset/versioned"
)

var (
	s2iClient     *s2i.Clientset
	s2iClientOnce sync.Once
)

func S2iClient() *s2i.Clientset {

	s2iClientOnce.Do(func() {

		config, err := Config()

		if err != nil {
			log.Fatalln(err)
		}

		s2iClient = s2i.NewForConfigOrDie(config)

		KubeConfig = config
	})

	return s2iClient
}
