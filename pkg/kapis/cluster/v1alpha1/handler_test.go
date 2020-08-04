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

package v1alpha1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/version"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration"
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
  namespace: kubesphere-system
spec:
  selector:
    matchLabels:
      app: agent
      app.kubernetes.io/part-of: tower
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: agent
        app.kubernetes.io/part-of: tower
    spec:
      containers:
      - command:
        - /agent
        - --name=gondor
        - --token=randomtoken
        - --proxy-server=http://139.198.121.121:8080
        - --keepalive=10s
        - --kubesphere-service=ks-apiserver.kubesphere-system.svc:80
        - --kubernetes-service=kubernetes.default.svc:443
        - --v=0
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
	k8sclient := k8sfake.NewSimpleClientset(service)
	ksclient := fake.NewSimpleClientset(cluster)

	informersFactory := informers.NewInformerFactories(k8sclient, ksclient, nil, nil, nil, nil)

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
			expectedError:  errClusterConnectionIsNotProxy,
			cluster:        directConnectionCluster,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.description, func(t *testing.T) {
			h := newHandler(informersFactory.KubernetesSharedInformerFactory().Core().V1().Services().Lister(),
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

			if diff := cmp.Diff(testCase.expected, buf.String()); len(diff) != 0 {
				t.Errorf("%T, got +, expected -, %s", testCase.expected, diff)
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

	if diff := cmp.Diff(buf.String(), expected); len(diff) != 0 {
		t.Error(diff)
	}

}

var base64EncodedKubeConfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKYzNCb1pYSmxNQjRYRFRJd01EVXlOekEzTWpFME1Gb1hEVE13TURVeU5UQTNNakUwTUZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYTndhR1Z5WlRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTWVxCjRwNldVQVJmeUtRcEpZM1ZGcEVhT3Y5REZBQkZhQXBpbUswQTBBL05KV25yaGl0MjhnL3ZXMno5bmxkeHQwZzgKc1J0ZEp4TUxmOHF5WkEramZudU5jUDBLUTJ5a3VxZE41c29MdE1TUmt1K1dHZFNkWlJvTVpEakdKbHRSUEdVRQpnOHd6OE9zdWRMcmZ5Zlcxdy8vUFRPWnRLNmsrNUhQZGtuWU5KcU9UZGJDTEVma2RZbFB1ZXNZTTFKamRacXlNClJQRDM1RXpUSEowR05jYlBzbUlyZ05WZGR4Nmh5RmIxTFZ0QXRqa0tWY2lNR1k1UlFTQWlQdVVGaGQreUcrOHUKUWlqQlgvV09ESlFOelJCQWFPdWRWdURqSkhIZ0lBV3FzcC9qSllWQkdRYnNad3djTTRUSEJPb2k3N1krYVR3SAphcHRaMitVMzgwbFY4d0tJR3NFQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDVjY0VVpCZkcxc1d4ampJWTZ5Vm0rdzhxTFgKS0ZYanR3VElGb1Y3WENOT2RDbEQ0a2NsN1pyTnl5UEtUaTFOV3BDbUVzSXZPVjRSRXV6a2ZWZ2Rzc0tHL2dYVwpGNGV2dStZTFh1UEk1RHJFejNGaW5OMGxVcFNZMVg3b1Y5N2JyU1lmdE53aWJQbUVFTEVHbWJvMnNHS3VoL21BCjRJZ0tmZklqWVdSYjE3TlZLb2s0am1WTnhkMmNCL09GeFlsY2xndlc2THpxc1BDdnpWbDdDRWErRElyNHZLamgKRlhvOERXejMzclUySCs0RjNMOThzWGE1OWNITWZPb1kzZ3pzUE5LYnQwbWsxeTNOUmZocTZYTTJXemkrUFZkcAppbUVqUlY4UUl5c3Zhc25sTjFIY1FtcWtMaFdLYlRoOEc5MGhiTzlwdTFRNE0wMmZ3STg0ZVlSUWZNUT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: 127.0.0.1:6578
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJZkcyYVlKYnRYT2N3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWE53YUdWeVpUQWVGdzB5TURBMU1qY3dOekl4TkRCYUZ3MHpNREExTWpVd056SXhOREJhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRiR3JpeFBxOWF6QU1wNGsKM0VsaElmanhOZXhkY2VORGJENkFMMXVGSWJUMk1uVnRXeFFXOU5ueVFwVU9NMzN4clNJMkJlVW1uU0U5RlR6cAo1bTBmaGVBOGhtRHRSZWVOckNxVnBJQy9kNUV2eDJ3NTJ2NXVCYUNXd09ZUGpQMi9uL0Y3THQwMmN3TkFXWE5pCmFRTi9uRGl6TjJrRXFoSmZiL0tyNGx3eEEzTDExVXJhMDNTRUp0U3FXSzBKQ1pnL0lzUnRFNFFqZXp1WWhiVWkKcWU0TmdqZjN1ZFBMMXQyeVpCK3hSTE1sNTFqenhYaXQ0U2pHSFJ2UEt1VHlkc1AxdEtINXdYdnhqaEZTZjN3UQpMSHFQd3hQWXVKSG5MWlhPaElnTnEzNnk4Rlp0WkdQRUhDVDFlUEh5cWhQaEFZZGlBRlRXT3pRN2FlS1puZWJzClpSblJMUUlEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFCZGNMN2I4akVuYjRSSndqZ3ZrVUxSQXJLVGk1WFhRSzllUwpsMlNkNEF3Ris5QWg3WElUazJFQ3J0WVR2WEM0K2trS1BxK2FuYjlOOFptenA0R0ovd3N5VEN2dHo5eGlQMTd3ClFCUXREdFA3eS9venlLc24rUzYvSDg2Y0JqdGZGT0dMYm5CekRBY3J0S3YyeUxxY3pyNlYzSDBnZDI3MVdlSkQKcU81U3czSEZoTHhERDRXSVVSQnFLOTJPVUhnSzVQOWRHaWdkK2MvQ0xWMFdJS1kzN3JGR2MrU3VUa2JOQXNmaQpmVmhBYXRsYlpQdE1QekJoV1hkM0JWcTMxTmtBM1F4aWUzdWc4Tm9OZ0czUHFPanJkZHA2aEs0Sk5wMGtkSGFvClZHRXMxcUdzOGJjekxNTTVzeUdkSndNbXVNUEdnaGxheUZrZlJ2RmpZWDlwbllJUStpTT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBdGJHcml4UHE5YXpBTXA0azNFbGhJZmp4TmV4ZGNlTkRiRDZBTDF1RkliVDJNblZ0Cld4UVc5Tm55UXBVT00zM3hyU0kyQmVVbW5TRTlGVHpwNW0wZmhlQThobUR0UmVlTnJDcVZwSUMvZDVFdngydzUKMnY1dUJhQ1d3T1lQalAyL24vRjdMdDAyY3dOQVdYTmlhUU4vbkRpek4ya0VxaEpmYi9LcjRsd3hBM0wxMVVyYQowM1NFSnRTcVdLMEpDWmcvSXNSdEU0UWplenVZaGJVaXFlNE5namYzdWRQTDF0MnlaQit4UkxNbDUxanp4WGl0CjRTakdIUnZQS3VUeWRzUDF0S0g1d1h2eGpoRlNmM3dRTEhxUHd4UFl1SkhuTFpYT2hJZ05xMzZ5OEZadFpHUEUKSENUMWVQSHlxaFBoQVlkaUFGVFdPelE3YWVLWm5lYnNaUm5STFFJREFRQUJBb0lCQVFDSDFnQ054YUpQY1l0dgpURlA2Yk5HMWVFdTlLS3pqekNoSDhLSWN4YXRPZTkvajhXNkVQUXk4bVlSSXl1OEhDQTE2aHEwazB5Qi9NSzVlCkJtQkg2U1U4RFZ5eWloeFp1cmRzRTVvMGxoeU80M2g0K3l4MTBPbW9RMXJ4ZEE0RU5tRGd6c1J0VU95NEo2SWcKUGVkQTQyQ3dCcVBWdFNuTGpGalZkUE9VRTZDQkZsa01nTVR2R1I5SE9EaVovdmhqRXhFQWg0UFFmMkl3ZXZhVwpxTXFsOVJNVFprVEUyVUJEbHZMWVRNcG5pdGlSSWdLdEdmSk1WTW81WWR4SWpOMXRnUHcwZmN5ZXkrM2JhdDZQCkVUSjk5K21KNDJMZGYzaTlNNkNvTDVaa1BjcEE0ckVSR1RKamtaT3FESzhPMVZXSDNtQ21WYTFyVGN0bmpJbVMKaGRGWElJWUJBb0dCQU84R25uekhzSVFXaExBSkw5U3VWWjhBODl5WHZuWHBFYUUxaHBRdEViRFlEaTBPcGg1UAp0cVVISlZ0T0ViaTYvai9tY1pwTnNNVzR2QitiV3hkbDA4U0g1Qnp2eW9qOEkvbUlVNDFCQkYya3hxMk8yNlJRCjRkdFQrN3NWUUFCT1ArUE0vcWtMOWVKdFI3blYwamRGd3BvcmtzcFo5cE9BZHNnUGVnUzlxUHFOQW9HQkFNS1kKeU5OQWdYRjhBdEV3ZEk5OVB2ZTVoYzN6QjZUclVOMStYNVNUZWNYWnFVSUNwMlJGTFNtL3RiVkRuay9VaXBXeQpYSjgzT084VGFIUjIrYk5OZU5NK2REK0diakVWdDNVMk1XSzN2azFjcy9oWjZrK0F2aE9iSXBrLzR2VEVRUW1FCjh4bkl5bkpPNWJleFQybkEyUWVYblVPbjZzdWtxVnk3YlBLSnNOa2hBb0dBV1YzNUxhQWZvQk1uUXdYOFN5RnYKUThiQVptNlp1RTRPMkY1QjFlN1AyWFcrUHh4bUFaaytLWTkxYVNEVVFXUXdvVVdRbmVlRU96aXBwWXVaVURNegpMUnk5cmcvOWdwLzY5MVlBSHlUNjgrUWlvRXQwVllna0diUFp2NFhmYXYzV3AxNUNySU9iU0RBaGpCcWt3U09rCjhhMXU4WmNYT09qa0FFTEJGVHF3RGhVQ2dZQVVqQjFvY1A4NkJHWW53SDRPU0tORmRRbHozWjJKQkcvZGMyS1UKUlo0dURmV1pTcjV5RC92YzFLbFRJbmlzNVR4YzRpQjFqMWNycDFqNE16ZmFmdXVySW9VVDBCWUNpTkIrUitLZgpFZGUrUTNPZFhhRW9FK2YrR2Z0bFF5R3J4cTAzWEJwdk5veHAxWHJjRXBUWURjemN5RjJLcjBoVGlHZDVxekN0Cnkyd3BBUUtCZ1FDWTFKdlp0YkFlZll2RXQ4c3JETVZsb1FOWVVKTURxWnBLenpqR0w5S3FqdlFwbjJOL2EwNmsKUzNsMndWWExXNy8xb0RMV2gxZGZLbUJlUlJhZUxJMXIvS3FyeTRjUStiVEIzTzBwS2R3WjFFYXNQajBDTnlaUAp0YnFkOEFWa09pRkFNYlRWaVZuR3RzSnJ3c1V2N1NSTUU4ckloMVJva0ZtRGRpN0J1UVl5emc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`

func TestValidateKubeConfig(t *testing.T) {
	k8sclient := k8sfake.NewSimpleClientset(service)
	ksclient := fake.NewSimpleClientset(cluster)

	informersFactory := informers.NewInformerFactories(k8sclient, ksclient, nil, nil, nil, nil)

	informersFactory.KubernetesSharedInformerFactory().Core().V1().Services().Informer().GetIndexer().Add(service)
	informersFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters().Informer().GetIndexer().Add(cluster)

	h := newHandler(informersFactory.KubernetesSharedInformerFactory().Core().V1().Services().Lister(),
		informersFactory.KubeSphereSharedInformerFactory().Cluster().V1alpha1().Clusters().Lister(),
		proxyService,
		"",
		agentImage)

	config, err := loadKubeConfigFromBytes([]byte(base64EncodedKubeConfig))
	if err != nil {
		t.Fatal(err)
	}

	// config.Host is schemaless, we need to add schema manually
	u, err := url.Parse(fmt.Sprintf("http://%s", config.Host))
	if err != nil {
		t.Fatal(err)
	}

	// we need to specify apiserver port to match above kubeconfig
	env := &envtest.Environment{
		Config: config,
		ControlPlane: integration.ControlPlane{
			APIServer: &integration.APIServer{
				Args: envtest.DefaultKubeAPIServerFlags,
				URL:  u,
			},
		},
	}

	cfg, err := env.Start()
	if err != nil {
		t.Log(cfg)
		t.Fatal(err)
	}

	defer func() {
		_ = env.Stop()
	}()

	err = h.validateKubeConfig([]byte(base64EncodedKubeConfig))
	if err != nil {
		t.Fatal(err)
	}
}

var ver = version.Get()

func endpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ver)
}

func TestValidateKubeSphereEndpoint(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(endpoint))
	defer svr.Close()

	got, err := validateKubeSphereAPIServer(svr.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&ver, got); len(diff) != 0 {
		t.Errorf("%T +got, -expected %v", ver, diff)
	}

}
