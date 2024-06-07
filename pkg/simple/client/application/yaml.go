package application

import (
	"context"
	"encoding/json"
	"time"

	helmrelease "helm.sh/helm/v3/pkg/release"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"kubesphere.io/utils/helm"
)

var _ helm.Executor = &YamlInstaller{}

type YamlInstaller struct {
	Mapper      meta.RESTMapper
	DynamicCli  *dynamic.DynamicClient
	GvrListInfo []InsInfo
	Namespace   string
}
type InsInfo struct {
	schema.GroupVersionResource
	Name      string
	Namespace string
}

func (t YamlInstaller) Install(ctx context.Context, release, chart string, values []byte, options ...helm.HelmOption) (string, error) {
	return "", nil
}

func (t YamlInstaller) Upgrade(ctx context.Context, release, chart string, values []byte, options ...helm.HelmOption) (string, error) {
	yamlList, err := ReadYaml(values)
	if err != nil {
		return "", err
	}
	klog.Infof("attempting to apply %d yaml files", len(yamlList))

	err = t.ForApply(yamlList)

	return "", err
}

func (t YamlInstaller) Uninstall(ctx context.Context, release string, options ...helm.HelmOption) (string, error) {
	for _, i := range t.GvrListInfo {
		err := t.DynamicCli.Resource(i.GroupVersionResource).Namespace(i.Namespace).
			Delete(ctx, i.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func (t YamlInstaller) ForceDelete(ctx context.Context, release string, options ...helm.HelmOption) error {
	_, err := t.Uninstall(ctx, release, options...)
	return err
}

func (t YamlInstaller) Get(ctx context.Context, releaseName string, options ...helm.HelmOption) (*helmrelease.Release, error) {
	rv := &helmrelease.Release{}
	rv.Info = &helmrelease.Info{Status: helmrelease.StatusDeployed}
	return rv, nil
}

func (t YamlInstaller) WaitingForResourcesReady(ctx context.Context, release string, timeout time.Duration, options ...helm.HelmOption) (bool, error) {
	return true, nil
}

func (t YamlInstaller) ForApply(tasks []json.RawMessage) (err error) {

	for idx, js := range tasks {

		gvr, utd, err := GetInfoFromBytes(js, t.Mapper)
		if err != nil {
			return err
		}
		opt := metav1.PatchOptions{FieldManager: "v1.FieldManager"}
		_, err = t.DynamicCli.Resource(gvr).
			Namespace(utd.GetNamespace()).
			Patch(context.TODO(), utd.GetName(), types.ApplyPatchType, js, opt)

		if err != nil {
			return err
		}
		klog.Infof("[%d/%d] %s/%s applied", idx+1, len(tasks), gvr.Resource, utd.GetName())
	}
	return nil
}
