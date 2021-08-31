/*

 Copyright 2021 The KubeSphere Authors.

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
package kubeconfig

import (
	"testing"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/constants"
)

const fakeKubeConfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeE1EWXlPVEEzTXpBME5Wb1hEVE14TURZeU56QTNNekEwTlZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTHFKCk52NnRiWWdyampJbFliSkZDWFNkaVNjYWxuckE2cGNEakQya2tBaW1RNXlkNEdrV0QwcTQ0WEpGSFdxcjRsWkwKTkJJSjlQSUFNNzVWRVVWcjh6NFNOVVBmckEvVERtSTZhaTlQVGNYOWtlOGFCZzV1U0dsbS9LUkZlcVVwVXA3awozUkp5MjlVcGNYb2pITm1EY0FPaWFLRi9NbnliZG1pV0lmcUJHaVZMSEdhcmdleTZCVzgrTGVNR3NWV0lpWVhVCkUwK3F0MG96R0lJaUNhVC9CaEdwNHlLczVWT0NheWRjNStaUnppYUJQMTk1Q3JqRllJNVR0UHMzb3JBcGhVSzcKZmd3NjFSZWhsMHQyd0x6bFFLSjM4RXJSNlUzMGwwR3h0MzhRTTVwbkt3cTQvOFBvbjkxYTlaNE1Dc3J6aDVYegpnbXZ4RmFyS0kxMWNQclRwaCtjQ0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZPa3NZMHExdEFUL3RUZ1JldG1kVHNDamN2Nm9NQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFBNXhuQngrdDZickMxNWttSHdLemdId09RZDlvNHpwR2orcDdmb2x5RnlEempuMDFOMwp4YXpwUUF4TU1UVHNtVjJMWWFyVm9KOU4xOGlmVndQV01HNEoyTGc3TTFBVUFKVU1BdmVYU0cveVY5eGx2QUtlCkdsek0wRSs5Y1IxR1cxL1hQcHc0ZWpmYWM0T2hlN09XUEhDcVVFVHJ0eWlTcmJGcWU3dmNLbS82dGlhQWphclUKMllzOGMzbjAyZUdKc1B4RUVwazVjRC9WQUxNOWlCUzJZQnBCanc0dDdHWTFERWtya2xsNkx1R0VtS05GRVBKOQpLOHFIYTQ2TFVTT3pNS1NLM2xndFIxQ2ZpSTBJZFBhdUQ5eGdaZ0VqZGdkcloxTHhYT01RTXlmOGl1Z0ZWblQvCmcyU0pjSEQ4QUZLQmwrUEZJdExuTVhBcEh1aUd2SkVLNzg1NQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://192.168.0.3:6443
  name: cluster.local
contexts:
- context:
    cluster: cluster.local
    user: kubernetes-admin
  name: kubernetes-admin@cluster.local
current-context: kubernetes-admin@cluster.local
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURFekNDQWZ1Z0F3SUJBZ0lJSkE2T3o0VitnTTR3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRBMk1qa3dOek13TkRWYUZ3MHlNakEyTWprd056TXdORGhhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQTBDaVd3OFl3cUpuT1FNU2kKQk9tZy9CeldseXZtU2dRU25iSlFyelZ3dUpYNWd2c0ZidFkwRmZwRFhqOTBvblg0UnRlelJWV3BtdHRrVEJXQwpzeHZqcitENUU0Rk9oV3AySHNsM1V3NWdTdVk0KzhvYWV3ZUwrRXRLL0FIeFlnL3Q5SE9reGhwRi9iYmVSTzhvCkN1NmRiT1dMZFpvTmN5RDlUaEgwZC8rZy9CakwwbklHQ0tpNk4rRloyQk5ZRkMxMWhmaitPUm1WRTdnTmQwYkQKVlp6YXYvOXVoZmljWUlBQ0FYa2d5NU5EWHY4enFXQ2NZY0VwbWppZ1RtZGFaV3N2c0F5QUh6c1gzS1JaNHU2VgpVbktqY09jWFdaZ1RqSE5xb3pjUEw4cEszQVBsbndsTThYcXd1S1JILzdnREVTNWRGaEZRdFdRMjB5TEVtNkNHCkJ5ekRTd0lEQVFBQm8wZ3dSakFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0h3WURWUjBqQkJnd0ZvQVU2U3hqU3JXMEJQKzFPQkY2MloxT3dLTnkvcWd3RFFZSktvWklodmNOQVFFTApCUUFEZ2dFQkFMY1J5Q3RGY2cwc0JIUFNibm9kQ21ralhUaFFKRFRDRWNObDNOMVJySWlZYncxRTUyV0hoZURkCll2OFBiK0hoeHRtYzNETzJSV2V5MWhJT2NGL1JHRE11bXpYMkJRMThuSU5zRk9ZTzF5ejNEamdsQ1RHdVdqejYKcmc4ZEZBWjVxMzhxT1pQYjF6RE1sWVZIdGQ2QVR3eFRxbjZhL3N3RXdsYVo1ME5JMzBCNTJMTXNYWWVJSlJ3NQpEUlZ3KzhVR3l3dDgwU3YxU3dvamRMd3dWcHhCc0lYemJBNXJjR3B6by9jayt2ZDI0Yys3bzYvVGJJV0hmVWxNCloyMzBobGNGS2t1OU8wb2habEVYVGpOQTVQcUdSdG5ieXlsaEdOWWxHaUVMQTQvK1Z6ZWZ0YXJoMmwvL1E4d3EKRElNTlJmazBwQTBTb21IUWl4d1FlTktCRDBYd3ZZRT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBMENpV3c4WXdxSm5PUU1TaUJPbWcvQnpXbHl2bVNnUVNuYkpRcnpWd3VKWDVndnNGCmJ0WTBGZnBEWGo5MG9uWDRSdGV6UlZXcG10dGtUQldDc3h2anIrRDVFNEZPaFdwMkhzbDNVdzVnU3VZNCs4b2EKZXdlTCtFdEsvQUh4WWcvdDlIT2t4aHBGL2JiZVJPOG9DdTZkYk9XTGRab05jeUQ5VGhIMGQvK2cvQmpMMG5JRwpDS2k2TitGWjJCTllGQzExaGZqK09SbVZFN2dOZDBiRFZaemF2Lzl1aGZpY1lJQUNBWGtneTVORFh2OHpxV0NjClljRXBtamlnVG1kYVpXc3ZzQXlBSHpzWDNLUlo0dTZWVW5LamNPY1hXWmdUakhOcW96Y1BMOHBLM0FQbG53bE0KOFhxd3VLUkgvN2dERVM1ZEZoRlF0V1EyMHlMRW02Q0dCeXpEU3dJREFRQUJBb0lCQUhyU0NDc1pyS28rbmprUApESDRUajc1U0ViZisyaEdBRjYvZWY4YnhwRUgxazlSWjRwbkVYOVU3NWpZZEFPZSs3YkIzSXpyYzBZY2l2aW82Cll2VGxsdEcyejZCWG9vb01DQWdnWFh5dk5kZmJ3WEdualRwY2VKVVhiL1lEKzNZZDZneGJrN1NqMmZwYXhRa3QKaDVYenR3V0M1MmVMYnpZb0YrM1JvRXFSbFY1SS9zTEZRVndMNmhEZkF2cnJXUTBHajVJZmZtNUxqeTR3VFdmQQpCWkRidEVWWVVZVkdqVWZ3ZndFTUpoUjlaZmk3OUhBNVgxZ0xLMWkxTDJNdTdzZzl5YUoxMDRubVpuQW5Fc2ViCkRDRnlJbHBCMUM1Tlh4Qm0rSGtCL2JVMkd0OTdwVXVHR2h4RkRtTmJlTyt4NS92Wm1Ha0tvTEE1V242NWM3RG8KeWJUV3hGRUNnWUVBNmNyV3BmT0Nhb29TWmc1eFBCSHpwbkd5d29VN0t0UTJGcURxc0I1djdIOFpTbGxhS0VaNQpGZkUwRCszcUFXeWtmTTQ0bm54Uk5xN1hOR3Rwa1c0MFhRREg3V2JyTjdVYUV4TjJCSnVXUUNWbEtiaG8yMUszCkp1K0lUaDQ5bUlPSWkrcUFFWHVxZ2hLaHpkQXZsZGF0L2YwVThHK0srQVJoT3dOalJJcTdEN01DZ1lFQTQrNW8KOWdEekI1eWtPQ1NBVEw2ckYyTmFjUGlLaUdZSEoyV1dPY0RHeGFvSUs2VUk4WjUrY0VmSHFDbG9TSFFjem9yVApyYnAzM1R0eFhNbzhhQlZNc243Q0NNU25FMmxsT1NsWllxNGV6NEh6bStVSU92WVJ0cCtuTnk0SUJOU2ZRUE5wCnR5b2I3VjZqSXZTanpiOVNEQzVOakpweFFYU0txbStBWC9ZMjhna0NnWUVBb29HZHBnaVhWRnJZNHh1UzFnQmMKYmd1R0IvUDM1cE5QYlhjNDZtYWR3Yk91N3FFaEsvR2daUUllQUJ5TmxhUGd5ZWZHTDFPV1YvNDhGSEc5Rlp1Vwp4amF1d1hQU2VBeG9MVzVQa0hCZGhnVDRSb0dxVVJrenVkcXgwaXJ2QWI0Y0FiVmtnOEtFQ0puTzRuS2RRUGZTClJVUFBkRGowVGVVdGVJbW9USkpwNkVVQ2dZRUF6SHlibGZpTUVJd3JtR0xHNkJNM0U2aUMvMDg3bWR0UEY3MC8KNVZoWi9BUHJpSnhyUmJuWDNZdklSOG0rVVNJNnBlSk92bEhJTDZhZ3NZcU9YeUtjeUphSUphMm41dlpyWmJqLwpCRlVLTjBoeThhMnNrSmtxa3hqd3Y4U0FWVFVjR3YxR0hwbWNySHgzQjJsTGU4N2xJU0I1V21kRXJHQ045eEpKCnJjNEt4V0VDZ1lFQTV6NUR2QlFxdzc2bXhEQ3lUd05HcUh3MVFqZmYwdXdTNmU0cGNVV1M0MVdjT3dlUm9NNGwKakZWSXlUNVRUbXd1QTBlS0VtMXZiM000VU5TeHN0eG14WDlhbVg2RzRlSSt4Uy94QmdHbTMyTWNMVzl5NjQyZwpSaDhScjVETTdpUlI4VzZFWVpnMStLR2sxcHAvRmxFWFBlWVhnY3hzK01NV3NTNmhaK0YzQzU4PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`

func Test_operator_CreateKubeConfig(t *testing.T) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(fakeKubeConfig))
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset()
	k8sInformers := k8sinformers.NewSharedInformerFactory(k8sClient, 0)
	operator := NewOperator(k8sClient, k8sInformers.Core().V1().ConfigMaps().Lister(), config)

	user1 := &iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{
			Kind:       iamv1alpha2.ResourceKindUser,
			APIVersion: iamv1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "user1",
		},
	}

	if err := operator.CreateKubeConfig(user1); err != nil {
		t.Errorf("CreateKubeConfig() unexpected error %v", err)
	}

	if len(k8sClient.Actions()) != 2 {
		t.Errorf("CreateKubeConfig() unexpected action %v", k8sClient.Actions())
		return
	}

	csrCreateAction, ok := k8sClient.Actions()[0].(k8stesting.CreateActionImpl)
	if !ok {
		t.Errorf("CreateKubeConfig() unexpected action %v", k8sClient.Actions()[0])
		return
	}

	csr, ok := csrCreateAction.Object.(*certificatesv1.CertificateSigningRequest)
	if !ok {
		t.Errorf("CreateKubeConfig() unexpected object %v", csrCreateAction.Object)
		return
	}
	if csr.Labels[constants.UsernameLabelKey] != user1.Name || csr.Annotations[privateKeyAnnotation] == "" {
		t.Errorf("CreateKubeConfig() unexpected CertificateSigningRequest %v", csr)
		return
	}
	cmCreateAction := k8sClient.Actions()[1].(k8stesting.CreateActionImpl)
	if !ok {
		t.Errorf("CreateKubeConfig() unexpected action %v", k8sClient.Actions()[1])
		return
	}
	cm, ok := cmCreateAction.Object.(*corev1.ConfigMap)
	if !ok {
		t.Errorf("CreateKubeConfig() unexpected object %v", cmCreateAction.Object)
		return
	}
	if cm.Labels[constants.UsernameLabelKey] != user1.Name || len(cm.Data) == 0 {
		t.Errorf("CreateKubeConfig() unexpected ConfigMap %v", cm)
		return
	}
}
