package business

import (
	"time"

	osappsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes/kubetest"
)

// Consolidate fake/mock data used in tests per package

func FakeDeployments() []v1beta1.Deployment {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1beta1.Deployment{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin"},
					},
				},
			},
			Status: v1beta1.DeploymentStatus{
				Replicas:            1,
				AvailableReplicas:   1,
				UnavailableReplicas: 0,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v2",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin", versionLabel: "v2"},
					},
				},
			},
			Status: v1beta1.DeploymentStatus{
				Replicas:            2,
				AvailableReplicas:   1,
				UnavailableReplicas: 1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v3",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			Status: v1beta1.DeploymentStatus{
				Replicas:            2,
				AvailableReplicas:   0,
				UnavailableReplicas: 2,
			},
		},
	}
}

func FakeDuplicatedDeployments() []v1beta1.Deployment {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1beta1.Deployment{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "duplicated-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "duplicated", versionLabel: "v1"},
					},
				},
			},
			Status: v1beta1.DeploymentStatus{
				Replicas:            1,
				AvailableReplicas:   1,
				UnavailableReplicas: 0,
			},
		},
	}
}

func FakeReplicaSets() []v1beta2.ReplicaSet {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1beta2.ReplicaSet{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicaSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin"},
					},
				},
			},
			Status: v1beta2.ReplicaSetStatus{
				Replicas:          1,
				AvailableReplicas: 1,
				ReadyReplicas:     1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicaSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v2",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin", versionLabel: "v2"},
					},
				},
			},
			Status: v1beta2.ReplicaSetStatus{
				Replicas:          2,
				AvailableReplicas: 1,
				ReadyReplicas:     1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicaSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v3",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			Status: v1beta2.ReplicaSetStatus{
				Replicas:          2,
				AvailableReplicas: 0,
				ReadyReplicas:     2,
			},
		},
	}
}

func FakeDuplicatedReplicaSets() []v1beta2.ReplicaSet {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	controller := true
	return []v1beta2.ReplicaSet{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicaSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "duplicated-v1-12345",
				CreationTimestamp: meta_v1.NewTime(t1),
				OwnerReferences: []meta_v1.OwnerReference{meta_v1.OwnerReference{
					Controller: &controller,
					Kind:       "Deployment",
					Name:       "duplicated-v1",
				}},
			},
			Spec: v1beta2.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "duplicated", versionLabel: "v1"},
					},
				},
			},
			Status: v1beta2.ReplicaSetStatus{
				Replicas:          1,
				AvailableReplicas: 1,
				ReadyReplicas:     1,
			},
		},
	}
}

func FakeReplicationControllers() []v1.ReplicationController {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1.ReplicationController{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicationController",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1.ReplicationControllerSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin"},
					},
				},
			},
			Status: v1.ReplicationControllerStatus{
				Replicas:          1,
				AvailableReplicas: 1,
				ReadyReplicas:     1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicationController",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v2",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1.ReplicationControllerSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin", versionLabel: "v2"},
					},
				},
			},
			Status: v1.ReplicationControllerStatus{
				Replicas:          2,
				AvailableReplicas: 1,
				ReadyReplicas:     1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicationController",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v3",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1.ReplicationControllerSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			Status: v1.ReplicationControllerStatus{
				Replicas:          2,
				AvailableReplicas: 0,
				ReadyReplicas:     2,
			},
		},
	}
}

func FakeDeploymentConfigs() []osappsv1.DeploymentConfig {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []osappsv1.DeploymentConfig{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "DeploymentConfig",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: osappsv1.DeploymentConfigSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin"},
					},
				},
			},
			Status: osappsv1.DeploymentConfigStatus{
				Replicas:            1,
				AvailableReplicas:   1,
				UnavailableReplicas: 0,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "DeploymentConfig",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v2",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: osappsv1.DeploymentConfigSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin", versionLabel: "v2"},
					},
				},
			},
			Status: osappsv1.DeploymentConfigStatus{
				Replicas:            2,
				AvailableReplicas:   1,
				UnavailableReplicas: 1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "DeploymentConfig",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v3",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: osappsv1.DeploymentConfigSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			Status: osappsv1.DeploymentConfigStatus{
				Replicas:            2,
				AvailableReplicas:   0,
				UnavailableReplicas: 2,
			},
		},
	}
}

func FakeStatefulSets() []v1beta2.StatefulSet {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1beta2.StatefulSet{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "StatefulSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin"},
					},
				},
			},
			Status: v1beta2.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "StatefulSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v2",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "httpbin", versionLabel: "v2"},
					},
				},
			},
			Status: v1beta2.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 1,
			},
		},
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "StatefulSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "httpbin-v3",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			Status: v1beta2.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 2,
			},
		},
	}
}

func FakeDuplicatedStatefulSets() []v1beta2.StatefulSet {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1beta2.StatefulSet{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "StatefulSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "duplicated-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta2.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "duplicated", versionLabel: "v1"},
					},
				},
			},
			Status: v1beta2.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		},
	}
}

func FakeDepSyncedWithRS() []v1beta1.Deployment {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	return []v1beta1.Deployment{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "details-v1",
				CreationTimestamp: meta_v1.NewTime(t1),
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "details", versionLabel: "v1"},
					},
				},
			},
			Status: v1beta1.DeploymentStatus{
				Replicas:            1,
				AvailableReplicas:   1,
				UnavailableReplicas: 0,
			},
		},
	}
}

func FakeRSSyncedWithPods() []v1beta2.ReplicaSet {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	controller := true
	return []v1beta2.ReplicaSet{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "ReplicaSet",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "details-v1-3618568057",
				CreationTimestamp: meta_v1.NewTime(t1),
				OwnerReferences: []meta_v1.OwnerReference{meta_v1.OwnerReference{
					Controller: &controller,
					Kind:       "Deployment",
					Name:       "details-v1",
				}},
			},
			Spec: v1beta2.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{appLabel: "details", versionLabel: "v1"},
					},
				},
			},
			Status: v1beta2.ReplicaSetStatus{
				Replicas:          1,
				AvailableReplicas: 1,
				ReadyReplicas:     0,
			},
		},
	}
}

func FakePodsSyncedWithDeployments() []v1.Pod {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	controller := true
	return []v1.Pod{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "details-v1-3618568057-dnkjp",
				CreationTimestamp: meta_v1.NewTime(t1),
				Labels:            map[string]string{appLabel: "httpbin", versionLabel: "v1"},
				OwnerReferences: []meta_v1.OwnerReference{meta_v1.OwnerReference{
					Controller: &controller,
					Kind:       "ReplicaSet",
					Name:       "details-v1-3618568057",
				}},
				Annotations: kubetest.FakeIstioAnnotations(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{Name: "details", Image: "whatever"},
					v1.Container{Name: "istio-proxy", Image: "docker.io/istio/proxy:0.7.1"},
				},
				InitContainers: []v1.Container{
					v1.Container{Name: "istio-init", Image: "docker.io/istio/proxy_init:0.7.1"},
					v1.Container{Name: "enable-core-dump", Image: "alpine"},
				},
			},
		},
	}
}

func FakePodsSyncedWithDuplicated() []v1.Pod {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	controller := true
	return []v1.Pod{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "duplicated-v1-3618568057-1",
				CreationTimestamp: meta_v1.NewTime(t1),
				Labels:            map[string]string{appLabel: "duplicated", versionLabel: "v1"},
				OwnerReferences: []meta_v1.OwnerReference{meta_v1.OwnerReference{
					Controller: &controller,
					Kind:       "StatefulSet",
					Name:       "duplicated-v1",
				}},
				Annotations: kubetest.FakeIstioAnnotations(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{Name: "details", Image: "whatever"},
					v1.Container{Name: "istio-proxy", Image: "docker.io/istio/proxy:0.7.1"},
				},
				InitContainers: []v1.Container{
					v1.Container{Name: "istio-init", Image: "docker.io/istio/proxy_init:0.7.1"},
					v1.Container{Name: "enable-core-dump", Image: "alpine"},
				},
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "duplicated-v1-3618568057-3",
				CreationTimestamp: meta_v1.NewTime(t1),
				Labels:            map[string]string{appLabel: "duplicated", versionLabel: "v1"},
				OwnerReferences: []meta_v1.OwnerReference{meta_v1.OwnerReference{
					Controller: &controller,
					Kind:       "StatefulSet",
					Name:       "duplicated-v1",
				}},
				Annotations: kubetest.FakeIstioAnnotations(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{Name: "details", Image: "whatever"},
					v1.Container{Name: "istio-proxy", Image: "docker.io/istio/proxy:0.7.1"},
				},
				InitContainers: []v1.Container{
					v1.Container{Name: "istio-init", Image: "docker.io/istio/proxy_init:0.7.1"},
					v1.Container{Name: "enable-core-dump", Image: "alpine"},
				},
			},
		},
	}
}

func FakePodsNoController() []v1.Pod {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")

	return []v1.Pod{
		{
			TypeMeta: meta_v1.TypeMeta{
				Kind: "Pod",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "orphan-pod",
				CreationTimestamp: meta_v1.NewTime(t1),
				Labels:            map[string]string{appLabel: "httpbin", versionLabel: "v1"},
				Annotations:       kubetest.FakeIstioAnnotations(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{Name: "details", Image: "whatever"},
					v1.Container{Name: "istio-proxy", Image: "docker.io/istio/proxy:0.7.1"},
				},
				InitContainers: []v1.Container{
					v1.Container{Name: "istio-init", Image: "docker.io/istio/proxy_init:0.7.1"},
					v1.Container{Name: "enable-core-dump", Image: "alpine"},
				},
			},
		},
	}
}

func FakePodsFromDaemonSet() []v1.Pod {
	conf := config.NewConfig()
	config.Set(conf)
	appLabel := conf.IstioLabels.AppLabelName
	versionLabel := conf.IstioLabels.VersionLabelName
	t1, _ := time.Parse(time.RFC822Z, "08 Mar 18 17:44 +0300")
	controller := true
	return []v1.Pod{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              "daemon-pod",
				CreationTimestamp: meta_v1.NewTime(t1),
				Labels:            map[string]string{appLabel: "httpbin", versionLabel: "v1"},
				OwnerReferences: []meta_v1.OwnerReference{meta_v1.OwnerReference{
					Controller: &controller,
					Kind:       "DaemonSet",
					Name:       "daemon-controller",
				}},
				Annotations: kubetest.FakeIstioAnnotations(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{Name: "details", Image: "whatever"},
					v1.Container{Name: "istio-proxy", Image: "docker.io/istio/proxy:0.7.1"},
				},
				InitContainers: []v1.Container{
					v1.Container{Name: "istio-init", Image: "docker.io/istio/proxy_init:0.7.1"},
					v1.Container{Name: "enable-core-dump", Image: "alpine"},
				},
			},
		},
	}
}

func FakeServices() []v1.Service {
	return []v1.Service{
		{
			ObjectMeta: meta_v1.ObjectMeta{Name: "httpbin"},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{"app": "httpbin"},
			},
		},
	}
}
