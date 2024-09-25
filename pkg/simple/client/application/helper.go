package application

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"kubesphere.io/utils/helm"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	pkgconstants "kubesphere.io/kubesphere/pkg/constants"

	k8serr "k8s.io/apimachinery/pkg/api/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/example/stringutil"
	"github.com/speps/go-hashids"

	"kubesphere.io/kubesphere/pkg/utils/idutils"

	"k8s.io/klog/v2"

	"helm.sh/helm/v3/pkg/chart"

	"helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"

	"kubesphere.io/utils/s3"

	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	appv2 "kubesphere.io/api/application/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/scheme"
)

type AppRequest struct {
	RepoName     string                       `json:"repoName,omitempty"`
	AppName      string                       `json:"appName,omitempty"`
	OriginalName string                       `json:"originalName,omitempty"`
	AliasName    string                       `json:"aliasName,omitempty"`
	VersionName  string                       `json:"versionName,omitempty"`
	AppHome      string                       `json:"appHome,omitempty"`
	Url          string                       `json:"url,omitempty"`
	Icon         string                       `json:"icon,omitempty"`
	Digest       string                       `json:"digest,omitempty"`
	Workspace    string                       `json:"workspace,omitempty"`
	Description  string                       `json:"description,omitempty"`
	CategoryName string                       `json:"categoryName,omitempty"`
	AppType      string                       `json:"appType,omitempty"`
	Package      []byte                       `json:"package,omitempty"`
	PullUrl      string                       `json:"pullUrl,omitempty"`
	Credential   appv2.RepoCredential         `json:"credential,omitempty"`
	Maintainers  []appv2.Maintainer           `json:"maintainers,omitempty"`
	Abstraction  string                       `json:"abstraction,omitempty"`
	Attachments  []string                     `json:"attachments,omitempty"`
	FromRepo     bool                         `json:"fromRepo,omitempty"`
	Resources    []appv2.GroupVersionResource `json:"resources,omitempty"`
}

func GetMaintainers(maintainers []*chart.Maintainer) []appv2.Maintainer {
	result := make([]appv2.Maintainer, len(maintainers))
	for i, maintainer := range maintainers {
		result[i] = appv2.Maintainer{
			Name:  maintainer.Name,
			Email: maintainer.Email,
			URL:   maintainer.URL,
		}
	}
	return result
}

func CreateOrUpdateApp(client runtimeclient.Client, vRequests []AppRequest, cmStore, ossStore s3.Interface, owns ...metav1.OwnerReference) error {
	ctx := context.Background()
	if len(vRequests) == 0 {
		return errors.New("version request is empty")
	}
	request := vRequests[0]

	app := appv2.Application{}
	app.Name = request.AppName

	operationResult, err := controllerutil.CreateOrUpdate(ctx, client, &app, func() error {
		app.Spec = appv2.ApplicationSpec{
			Icon:        request.Icon,
			AppHome:     request.AppHome,
			AppType:     request.AppType,
			Abstraction: request.Abstraction,
			Attachments: request.Attachments,
		}
		if len(owns) > 0 {
			app.OwnerReferences = owns
		}

		labels := app.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[appv2.RepoIDLabelKey] = request.RepoName
		labels[appv2.AppTypeLabelKey] = request.AppType

		if request.CategoryName != "" {
			labels[appv2.AppCategoryNameKey] = request.CategoryName
		} else {
			labels[appv2.AppCategoryNameKey] = appv2.UncategorizedCategoryID
		}

		labels[constants.WorkspaceLabelKey] = request.Workspace
		app.SetLabels(labels)

		ant := app.GetAnnotations()
		if ant == nil {
			ant = make(map[string]string)
		}
		ant[constants.DisplayNameAnnotationKey] = request.AliasName
		ant[constants.DescriptionAnnotationKey] = request.Description
		ant[appv2.AppOriginalNameLabelKey] = request.OriginalName

		if len(request.Maintainers) > 0 {
			ant[appv2.AppMaintainersKey] = request.Maintainers[0].Name
		}
		app.SetAnnotations(ant)

		return nil
	})
	if err != nil {
		klog.Errorf("failed create or update app %s, err:%v", app.Name, err)
		return err
	}

	if operationResult == controllerutil.OperationResultCreated {
		if request.FromRepo {
			app.Status.State = appv2.ReviewStatusActive
		} else {
			app.Status.State = appv2.ReviewStatusDraft
		}
	}

	app.Status.UpdateTime = &metav1.Time{Time: time.Now()}
	patch, _ := json.Marshal(app)
	if err = client.Status().Patch(ctx, &app, runtimeclient.RawPatch(runtimeclient.Merge.Type(), patch)); err != nil {
		klog.Errorf("failed to update app status, err:%v", err)
		return err
	}

	for _, vRequest := range vRequests {
		if err = CreateOrUpdateAppVersion(ctx, client, app, vRequest, cmStore, ossStore); err != nil {
			klog.Errorf("failed to create or update app version, err:%v", err)
			return err
		}
	}

	err = UpdateLatestAppVersion(ctx, client, app)
	if err != nil {
		klog.Errorf("failed to update latest app version, err:%v", err)
		return err
	}
	return nil
}

func CreateOrUpdateAppVersion(ctx context.Context, client runtimeclient.Client, app appv2.Application, vRequest AppRequest, cmStore, ossStore s3.Interface) error {

	//1. create or update app version
	appVersion := appv2.ApplicationVersion{}
	appVersion.Name = fmt.Sprintf("%s-%s", app.Name, vRequest.VersionName)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(&app, &appVersion, scheme.Scheme); err != nil {
			klog.Errorf("%s SetControllerReference failed, err:%v", appVersion.Name, err)
			return err
		}
		appVersion.Spec = appv2.ApplicationVersionSpec{
			VersionName: vRequest.VersionName,
			AppHome:     vRequest.AppHome,
			Icon:        vRequest.Icon,
			Created:     &metav1.Time{Time: time.Now()},
			Digest:      vRequest.Digest,
			AppType:     vRequest.AppType,
			Maintainer:  vRequest.Maintainers,
			PullUrl:     vRequest.PullUrl,
		}
		appVersion.Finalizers = []string{appv2.StoreCleanFinalizer}

		labels := appVersion.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[appv2.RepoIDLabelKey] = vRequest.RepoName
		labels[appv2.AppIDLabelKey] = vRequest.AppName
		labels[appv2.AppTypeLabelKey] = vRequest.AppType
		labels[constants.WorkspaceLabelKey] = vRequest.Workspace
		appVersion.SetLabels(labels)

		ant := appVersion.GetAnnotations()
		if ant == nil {
			ant = make(map[string]string)
		}
		ant[constants.DisplayNameAnnotationKey] = vRequest.AliasName
		ant[constants.DescriptionAnnotationKey] = vRequest.Description
		if len(vRequest.Maintainers) > 0 {
			ant[appv2.AppMaintainersKey] = vRequest.Maintainers[0].Name
		}
		appVersion.SetAnnotations(ant)
		return nil
	}
	_, err := controllerutil.CreateOrUpdate(ctx, client, &appVersion, mutateFn)

	if err != nil {
		klog.Errorf("failed create or update app version %s, err:%v", appVersion.Name, err)
		return err
	}

	if !vRequest.FromRepo {
		err = FailOverUpload(cmStore, ossStore, appVersion.Name, bytes.NewReader(vRequest.Package), len(vRequest.Package))
		if err != nil {
			klog.Errorf("upload package failed, error: %s", err)
			return err
		}
	}

	//3. update app version status
	if vRequest.FromRepo {
		appVersion.Status.State = appv2.ReviewStatusActive
	} else {
		appVersion.Status.State = appv2.ReviewStatusDraft
	}
	appVersion.Status.Updated = &metav1.Time{Time: time.Now()}
	patch, _ := json.Marshal(appVersion)
	if err = client.Status().Patch(ctx, &appVersion, runtimeclient.RawPatch(runtimeclient.Merge.Type(), patch)); err != nil {
		klog.Errorf("failed to update app version status, err:%v", err)
		return err
	}

	return err
}

func UpdateLatestAppVersion(ctx context.Context, client runtimeclient.Client, app appv2.Application) (err error) {
	//4. update app latest version
	err = client.Get(ctx, runtimeclient.ObjectKey{Name: app.Name, Namespace: app.Namespace}, &app)
	if err != nil {
		klog.Errorf("failed to get app, err:%v", err)
		return err
	}

	appVersionList := appv2.ApplicationVersionList{}
	lbs := labels.SelectorFromSet(labels.Set{appv2.AppIDLabelKey: app.Name})
	opt := runtimeclient.ListOptions{LabelSelector: lbs}
	err = client.List(ctx, &appVersionList, &opt)
	if err != nil {
		klog.Errorf("failed to list app version, err:%v", err)
		return err
	}
	if len(appVersionList.Items) == 0 {
		return nil
	}

	latestAppVersion := appVersionList.Items[0].Spec.VersionName
	for _, v := range appVersionList.Items {
		parsedVersion, err := semver.Make(strings.TrimPrefix(v.Spec.VersionName, "v"))
		if err != nil {
			klog.Warningf("failed to parse version: %s, use first version %s", v.Spec.VersionName, latestAppVersion)
			continue
		}
		if parsedVersion.GT(semver.MustParse(strings.TrimPrefix(latestAppVersion, "v"))) {
			latestAppVersion = v.Spec.VersionName
		}
	}

	ant := app.GetAnnotations()
	ant[appv2.LatestAppVersionKey] = latestAppVersion
	app.SetAnnotations(ant)
	err = client.Update(ctx, &app)
	return err
}

func HelmPull(u string, cred appv2.RepoCredential) (*bytes.Buffer, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	var resp *bytes.Buffer

	skipTLS := true
	if cred.InsecureSkipTLSVerify != nil && !*cred.InsecureSkipTLSVerify {
		skipTLS = false
	}

	indexURL := parsedURL.String()
	g, _ := getter.NewHTTPGetter()
	options := []getter.Option{
		getter.WithTimeout(5 * time.Minute),
		getter.WithURL(u),
		getter.WithInsecureSkipVerifyTLS(skipTLS),
		getter.WithTLSClientConfig(cred.CertFile, cred.KeyFile, cred.CAFile),
		getter.WithBasicAuth(cred.Username, cred.Password)}

	if skipTLS {
		options = append(options, getter.WithTransport(
			&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		))
	}

	resp, err = g.Get(indexURL, options...)
	return resp, err
}

func LoadRepoIndex(u string, cred appv2.RepoCredential) (idx helmrepo.IndexFile, err error) {
	if registry.IsOCI(u) {
		return LoadRepoIndexFromOci(u, cred)
	}

	if !strings.HasSuffix(u, "/") {
		u = fmt.Sprintf("%s/index.yaml", u)
	} else {
		u = fmt.Sprintf("%sindex.yaml", u)
	}

	resp, err := HelmPull(u, cred)
	if err != nil {
		return idx, err
	}
	if err = yaml.Unmarshal(resp.Bytes(), &idx); err != nil {
		return idx, err
	}
	idx.SortEntries()

	return idx, nil
}

func ReadYaml(data []byte) (jsonList []json.RawMessage, err error) {
	reader := bytes.NewReader(data)
	bufReader := bufio.NewReader(reader)
	r := yaml.NewYAMLReader(bufReader)
	for {
		d, err := r.Read()
		if err != nil && err == io.EOF {
			break
		}
		jsonData, err := yaml.ToJSON(d)
		if err != nil {
			return nil, err
		}
		_, _, err = Decode(jsonData)
		if err != nil {
			return nil, err
		}
		jsonList = append(jsonList, jsonData)
	}
	return jsonList, nil
}

func Decode(data []byte) (obj runtime.Object, gvk *schema.GroupVersionKind, err error) {
	decoder := unstructured.UnstructuredJSONScheme
	obj, gvk, err = decoder.Decode(data, nil, nil)
	return obj, gvk, err
}

func UpdateHelmStatus(kubeConfig []byte, release *helmrelease.Release) (deployed bool, err error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		klog.Errorf("failed to get rest config, err:%v", err)
		return deployed, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("failed to get kubernetes client, err:%v", err)
		return deployed, err
	}

	actionConfig := new(action.Configuration)
	store := storage.Init(driver.NewSecrets(clientSet.CoreV1().Secrets(release.Namespace)))
	actionConfig.Releases = store

	deployed, err = checkReady(release, clientSet, kubeConfig)
	if err != nil {
		klog.Errorf("failed to check helm ready, err:%v", err)
		return deployed, err
	}
	if !deployed {
		klog.Infof("helm release %s not ready", release.Name)
		return deployed, nil
	}

	klog.Infof("helm release %s now ready", release.Name)
	release.SetStatus("deployed", "Successfully deployed")

	if err = actionConfig.Releases.Update(release); err != nil {
		klog.Errorf("failed to update release: %v", err)
		return deployed, err
	}
	klog.Infof("update release %s status successfully", release.Name)
	return true, err
}

func checkReady(release *helmrelease.Release, clientSet *kubernetes.Clientset, kubeConfig []byte) (allReady bool, err error) {

	checker := kube.NewReadyChecker(clientSet, nil, kube.PausedAsReady(true), kube.CheckJobs(true))

	helmConf, err := helm.InitHelmConf(kubeConfig, release.Namespace)
	if err != nil {
		klog.Errorf("failed to init helm conf, err:%v", err)
		return allReady, err
	}

	allResources := make([]*resource.Info, 0)
	for _, i := range release.Hooks {
		hookResources, err := helmConf.KubeClient.Build(bytes.NewBufferString(i.Manifest), false)
		if err != nil {
			klog.Errorf("failed to get helm hookResources, err:%v", err)
			return allReady, err
		}
		allResources = append(allResources, hookResources...)
	}
	klog.Infof("%s get helm hookResources %d", release.Name, len(allResources))

	chartResources, err := helmConf.KubeClient.Build(bytes.NewBufferString(release.Manifest), false)
	if err != nil {
		klog.Errorf("failed to get helm resources, err:%v", err)
		return allReady, err
	}
	allResources = append(allResources, chartResources...)
	klog.Infof("%s get helm chartResources %d", release.Name, len(chartResources))

	for idx, j := range allResources {
		kind := j.Object.GetObjectKind().GroupVersionKind().Kind
		klog.Infof("[%d/%d] check helm release %s  %s: %s/%s", idx+1, len(allResources),
			release.Name, kind, j.Namespace, j.Name)
		ready, err := checker.IsReady(context.Background(), j)
		if k8serr.IsNotFound(err) {
			//pre-job-->chart-resource-->post-job
			//If a certain step times out, the subsequent steps will not be created,
			//and the status is considered failed, no repair will be made.
			klog.Warningf("[%d/%d] helm release %s resource %s: %s/%s not found", idx+1, len(allResources), release.Name, kind, j.Namespace, j.Name)
			return false, nil
		}

		if err != nil {
			klog.Errorf("failed to check resource ready, err:%v", err)
			return allReady, err
		}
		if !ready {
			klog.Infof("[%d/%d] helm release %s resource %s: %s/%s not ready", idx+1, len(allResources), release.Name, kind, j.Namespace, j.Name)
			return allReady, nil
		}
	}
	return true, nil
}

func GvkToGvr(gvk *schema.GroupVersionKind, mapper meta.RESTMapper) (schema.GroupVersionResource, error) {
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if meta.IsNoMatchError(err) || err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, nil
}
func GetInfoFromBytes(bytes json.RawMessage, mapper meta.RESTMapper) (gvr schema.GroupVersionResource, utd *unstructured.Unstructured, err error) {
	obj, gvk, err := Decode(bytes)
	if err != nil {
		return gvr, utd, err
	}
	gvr, err = GvkToGvr(gvk, mapper)
	if err != nil {
		return gvr, utd, err
	}
	utd, err = ConvertToUnstructured(obj)
	return gvr, utd, err
}
func ConvertToUnstructured(obj any) (*unstructured.Unstructured, error) {
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	return &unstructured.Unstructured{Object: objMap}, err
}

func ComplianceCheck(values, tempLate []byte, mapper meta.RESTMapper, ns string) (result []json.RawMessage, err error) {
	yamlList, err := ReadYaml(values)
	if err != nil {
		return nil, err
	}
	yamlTempList, err := ReadYaml(tempLate)
	if err != nil {
		return nil, err
	}

	if len(yamlTempList) != len(yamlList) {
		return nil, errors.New("yamlList and yamlTempList length not equal")
	}
	for idx := range yamlTempList {
		_, utd, err := GetInfoFromBytes(yamlList[idx], mapper)
		if err != nil {
			return nil, err
		}
		_, utdTemp, err := GetInfoFromBytes(yamlTempList[idx], mapper)
		if err != nil {
			return nil, err
		}
		if utdTemp.GetKind() != utd.GetKind() || utdTemp.GetAPIVersion() != utd.GetAPIVersion() {
			return nil, errors.New("yamlList and yamlTempList not equal")
		}
		if utd.GetNamespace() != ns {
			return nil, errors.New("subresource must have same namespace with app release")
		}
	}
	return yamlList, nil
}

func GetUuid36(prefix string) string {
	id := idutils.GetIntId()
	hd := hashids.NewData()
	hd.Alphabet = idutils.Alphabet36
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}
	i, err := h.Encode([]int{int(id)})
	if err != nil {
		panic(err)
	}
	//hashids.minAlphabetLength = 16
	add := stringutil.Reverse(i)[:5]

	return prefix + add
}

func GenerateShortNameMD5Hash(input string) string {
	input = strings.ToLower(input)
	errs := validation.IsDNS1123Subdomain(input)
	if len(input) > 14 || len(errs) != 0 {
		hash := md5.New()
		hash.Write([]byte(input))
		hashInBytes := hash.Sum(nil)
		hashString := hex.EncodeToString(hashInBytes)
		return hashString[:10]
	}
	return input
}

func FormatVersion(input string) string {
	re := regexp.MustCompile(`[^a-z0-9-.]`)
	errs := validation.IsDNS1123Subdomain(input)
	if len(errs) != 0 {
		klog.Warningf("Version %s does not meet the Kubernetes naming standard, replacing invalid characters with '-'", input)
		input = re.ReplaceAllStringFunc(input, func(s string) string {
			return "-"
		})
	}
	return input
}

func GetHelmKubeConfig(ctx context.Context, cluster *clusterv1alpha1.Cluster, runClient client.Client) (config []byte, err error) {

	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {
		klog.Infof("cluster %s is proxy cluster", cluster.Name)
		secret := &corev1.Secret{}
		key := types.NamespacedName{Namespace: pkgconstants.KubeSphereNamespace, Name: "kubeconfig-admin"}
		err = runClient.Get(ctx, key, secret)
		if err != nil {
			klog.Errorf("failed to get kubeconfig-admin secret: %v", err)
			return nil, err
		}
		config = secret.Data["config"]
		return config, err
	}
	return cluster.Spec.Connection.KubeConfig, nil
}
