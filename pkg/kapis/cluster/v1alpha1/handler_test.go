package v1alpha1

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	fake2 "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"testing"
)

const (
	proxyAddress = "http://139.198.121.121:8080"
	agentImage   = "kubesphere/tower:v1.0"
	proxyService = "tower.kubesphere-system.svc"
)

var cluster = &v1alpha1.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: "gondor",
	},
	Spec: v1alpha1.ClusterSpec{
		Connection: v1alpha1.Connection{
			Type:  v1alpha1.ConnectionTypeProxy,
			Token: "randomtoken",
		},
	},
}

var service = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "tower",
		Namespace: "kubesphere-system",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Port:     8080,
				Protocol: corev1.ProtocolTCP,
			},
		},
	},
	Status: corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					IP:       "139.198.121.121",
					Hostname: "foo.bar",
				},
			},
		},
	},
}

var expected = `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  name: cluster-agent
spec:
  selector: {}
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers:
      - command:
        - /agent
        - --name=gondor
        - --token=randomtoken
        - --proxy-server=http://139.198.121.121:8080
        - --kubesphere-service=ks-apiserver.kubesphere-system.svc:80
        - --kubernetes-service=kubernetes.default.svc:443
        image: kubesphere/tower:v1.0
        name: agent
        resources:
          limits:
            cpu: "1"
            memory: 200M
          requests:
            cpu: 100m
            memory: 100M
      serviceAccountName: kubesphere
status: {}
`

func TestGeranteAgentDeployment(t *testing.T) {
	k8sclient := fake2.NewSimpleClientset(service)
	ksclient := fake.NewSimpleClientset(cluster)

	informersFactory := informers.NewInformerFactories(k8sclient, ksclient, nil, nil)

	informersFactory.KubernetesSharedInformerFactory().Core().V1().Services().Informer().GetIndexer().Add(service)
	informersFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters().Informer().GetIndexer().Add(cluster)

	directConnectionCluster := cluster.DeepCopy()
	directConnectionCluster.Spec.Connection.Type = v1alpha1.ConnectionTypeDirect

	var testCases = []struct {
		description    string
		expectingError bool
		expectedError  error
		cluster        *v1alpha1.Cluster
		expected       string
	}{
		{
			description:    "test normal case",
			expectingError: false,
			expected:       expected,
			cluster:        cluster,
		},
		{
			description:    "test direct connection cluster",
			expectingError: true,
			expectedError:  ErrClusterConnectionIsNotProxy,
			cluster:        directConnectionCluster,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.description, func(t *testing.T) {
			h := NewHandler(informersFactory.KubernetesSharedInformerFactory().Core().V1().Services().Lister(),
				informersFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters().Lister(),
				proxyService,
				"",
				agentImage)

			var buf bytes.Buffer

			err := h.populateProxyAddress()
			if err != nil {
				t.Error(err)
			}

			err = h.generateDefaultDeployment(testCase.cluster, &buf)
			if testCase.expectingError {
				if err == nil {
					t.Fatalf("expecting error %v, got nil", testCase.expectedError)
				} else if err != testCase.expectedError {
					t.Fatalf("expecting error %v, got %v", testCase.expectedError, err)
				}
			}
		})
	}

}

func TestInnerGenerateAgentDeployment(t *testing.T) {
	h := &handler{
		proxyAddress: proxyAddress,
		agentImage:   agentImage,
		yamlPrinter:  &printers.YAMLPrinter{},
	}

	var buf bytes.Buffer

	err := h.generateDefaultDeployment(cluster, &buf)

	if err != nil {
		t.Error(err)
	}

	t.Log(buf.String())

	if diff := cmp.Diff(buf.String(), expected); len(diff) != 0 {
		t.Error(diff)
	}

}
