package runners

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type restarter struct {
	client        client.Client
	metricsClient MetricsClient
	interval      time.Duration
	log           logr.Logger
	recorder      record.EventRecorder
}

func NewRestarter(c client.Client, log logr.Logger, interval time.Duration, recorder record.EventRecorder) manager.Runnable {

	return &restarter{
		client:   c,
		log:      log,
		interval: interval,
		recorder: recorder,
	}
}

// Start implements manager.Runnable
func (c *restarter) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := c.reconcile(ctx)
			if err != nil {
				c.log.Error(err, "failed to reconcile")
				return err
			}
		}
	}
}

func (c *restarter) reconcile(ctx context.Context) error {
	stopPvcs, stopPvcErr := c.getPVCListByConditionsType(ctx, v1.PersistentVolumeClaimResizing)
	if stopPvcErr != nil {
		return stopPvcErr
	}

	stopDeploy, stopSts, timeoutDeploy, timeoutSts, stopAppErr := c.getAppList(ctx, stopPvcs)
	if stopAppErr != nil {
		return stopAppErr
	}
	//stop deploy/sts
	for _, deploy := range stopDeploy {
		c.log.Info("Stopping deploy: ", deploy.Name)
		err := c.stopDeploy(ctx, deploy)
		if err != nil {
			return err
		}
	}
	for _, sts := range stopSts {
		c.log.Info("Stopping StatefulSet: ", sts.Name)
		err := c.stopSts(ctx, sts)
		if err != nil {
			return err
		}
	}

	//get pvc need restart
	startPvc, err := c.getPVCListByConditionsType(ctx, v1.PersistentVolumeClaimFileSystemResizePending)

	//get list
	startDeploy := make([]*appsV1.Deployment, 0)
	startSts := make([]*appsV1.StatefulSet, 0)
	for _, pvc := range startPvc {
		dep, err := c.getDeploy(ctx, &pvc)
		if err != nil {
			return err
		}
		if dep != nil {
			startDeploy = append(startDeploy, dep)
			continue
		}
		sts, err := c.getSts(ctx, &pvc)
		if err != nil {
			return err
		}
		if sts != nil {
			startSts = append(startSts, sts)
		}
	}

	//restart
	for _, deploy := range startDeploy {
		err := c.StartDeploy(ctx, deploy, false)
		if err != nil {
			return err
		}
	}
	for _, deploy := range timeoutDeploy {
		err := c.StartDeploy(ctx, deploy, true)
		if err != nil {
			return err
		}
	}
	for _, sts := range startSts {
		err := c.StartSts(ctx, sts, false)
		if err != nil {
			return err
		}
	}
	for _, sts := range timeoutSts {
		err := c.StartSts(ctx, sts, true)
		if err != nil {
			return err
		}
	}
	return err
}

func (c *restarter) getPVCListByConditionsType(ctx context.Context, pvcType v1.PersistentVolumeClaimConditionType) ([]v1.PersistentVolumeClaim, error) {
	pvcs := make([]v1.PersistentVolumeClaim, 0)
	pvcList := v1.PersistentVolumeClaimList{}
	var opts []client.ListOption
	err := c.client.List(ctx, &pvcList, opts...)
	if err != nil {
		return pvcs, err
	}
	scNeedRestart := make(map[string]string, 0)
	scNeedRestart, err = c.getSc(ctx)
	if err != nil {
		return nil, err
	}
	for _, pvc := range pvcList.Items {
		if len(pvc.Status.Conditions) > 0 && pvc.Status.Conditions[0].Type == pvcType {
			if _, ok := scNeedRestart[*pvc.Spec.StorageClassName]; ok {
				pvcs = append(pvcs, pvc)
			} else if val, ok := pvc.Annotations[AutoRestartEnabledKey]; ok {
				NeedRestart, err := strconv.ParseBool(val)
				if err != nil {
					continue
				}
				if NeedRestart {
					pvcs = append(pvcs, pvc)
				}
			}
		}
	}
	return pvcs, nil
}

func (c *restarter) getSc(ctx context.Context) (map[string]string, error) {
	scList := &storagev1.StorageClassList{}
	var opts []client.ListOption
	err := c.client.List(ctx, scList, opts...)
	if err != nil {
		return nil, err
	}
	scMap := make(map[string]string, 0)
	for _, sc := range scList.Items {
		if val, ok := sc.Annotations[SupportOnlineResize]; ok {
			SupportOnline, err := strconv.ParseBool(val)
			if err != nil {
				return nil, err
			}
			if !SupportOnline {
				if val, ok := sc.Annotations[AutoRestartEnabledKey]; ok {
					NeedRestart, err := strconv.ParseBool(val)
					if err != nil {
						return nil, err
					}
					if NeedRestart {
						scMap[sc.Name] = ""
					}
				}
			}
		}
	}
	return scMap, nil
}

func (c *restarter) stopDeploy(ctx context.Context, deploy *appsV1.Deployment) error {
	var zero int32
	zero = 0
	if val, ok := deploy.Annotations[RestartSkip]; ok {
		skip, _ := strconv.ParseBool(val)
		if skip {
			c.log.Info("Skip restart deploy ", deploy.Name)
			return nil
		}
	}
	if stage, ok := deploy.Annotations[RestartStage]; ok {
		if stage == "resizing" {
			return nil
		}
	}
	replicas := *deploy.Spec.Replicas
	updateDeploy := deploy.DeepCopy()

	// add annotations
	updateDeploy.Annotations[RestartStopTime] = strconv.FormatInt(time.Now().Unix(), 10)
	updateDeploy.Annotations[ExpectReplicaNums] = strconv.Itoa(int(replicas))
	updateDeploy.Annotations[RestartStage] = "resizing"
	updateDeploy.Spec.Replicas = &zero
	var opts []client.UpdateOption
	c.log.Info("stop deployment:" + deploy.Name)
	updateErr := c.client.Update(ctx, updateDeploy, opts...)
	return updateErr
}

func (c *restarter) stopSts(ctx context.Context, sts *appsV1.StatefulSet) error {
	var zero int32
	zero = 0
	if val, ok := sts.Annotations[RestartSkip]; ok {
		skip, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
	}
	if stage, ok := sts.Annotations[RestartStage]; ok {
		if stage == "resizing" {
			return nil
		}
	}
	replicas := *sts.Spec.Replicas
	updateSts := sts.DeepCopy()

	// add annotations
	updateSts.Annotations[RestartStopTime] = strconv.FormatInt(time.Now().Unix(), 10)
	updateSts.Annotations[ExpectReplicaNums] = strconv.Itoa(int(replicas))
	updateSts.Annotations[RestartStage] = "resizing"
	updateSts.Spec.Replicas = &zero
	var opts []client.UpdateOption
	c.log.Info("stop deployment:" + sts.Name)
	updateErr := c.client.Update(ctx, updateSts, opts...)
	return updateErr
}

func (c *restarter) StartDeploy(ctx context.Context, deploy *appsV1.Deployment, timeout bool) error {
	if _, ok := deploy.Annotations[RestartStage]; !ok {
		return nil
	}
	if deploy.Annotations[RestartStage] != "resizing" {
		return fmt.Errorf("Unknown stage, skip ")
	}
	updateDeploy := deploy.DeepCopy()
	if _, ok := deploy.Annotations[ExpectReplicaNums]; !ok {
		return fmt.Errorf("Cannot find replica numbers before stop ")
	}
	expectReplicaNums, err := strconv.Atoi(deploy.Annotations[ExpectReplicaNums])
	if err != nil {
		return err
	}
	replicas := int32(expectReplicaNums)
	if timeout {
		updateDeploy.Annotations[RestartSkip] = "true"
	}
	delete(updateDeploy.Annotations, RestartStopTime)
	delete(updateDeploy.Annotations, ExpectReplicaNums)
	delete(updateDeploy.Annotations, RestartStage)
	updateDeploy.Spec.Replicas = &replicas
	var opts []client.UpdateOption
	c.log.Info("start deployment: " + deploy.Name)
	err = c.client.Update(ctx, updateDeploy, opts...)
	return err
}

func (c *restarter) StartSts(ctx context.Context, sts *appsV1.StatefulSet, timeout bool) error {
	if _, ok := sts.Annotations[RestartStage]; !ok {
		return nil
	}
	if sts.Annotations[RestartStage] != "resizing" {
		return fmt.Errorf("Unknown stage, skip ")
	}
	updateSts := sts.DeepCopy()
	if _, ok := sts.Annotations[ExpectReplicaNums]; !ok {
		return fmt.Errorf("Cannot find replica numbers before stop ")
	}
	expectReplicaNums, err := strconv.Atoi(sts.Annotations[ExpectReplicaNums])
	if err != nil {
		return err
	}
	replicas := int32(expectReplicaNums)
	if timeout {
		updateSts.Annotations[RestartSkip] = "true"
	}
	delete(updateSts.Annotations, RestartStopTime)
	delete(updateSts.Annotations, ExpectReplicaNums)
	delete(updateSts.Annotations, RestartStage)
	updateSts.Spec.Replicas = &replicas
	var opts []client.UpdateOption
	c.log.Info("start deployment: " + sts.Name)
	err = c.client.Update(ctx, updateSts, opts...)
	return err
}

func (c *restarter) getDeploy(ctx context.Context, pvc *v1.PersistentVolumeClaim) (*appsV1.Deployment, error) {
	// get deploy list
	deployList := &appsV1.DeploymentList{}
	var opts []client.ListOption
	err := c.client.List(ctx, deployList, opts...)
	if err != nil {
		return nil, err
	}

	for _, deploy := range deployList.Items {
		if len(deploy.Spec.Template.Spec.Volumes) > 0 {
			for _, vol := range deploy.Spec.Template.Spec.Volumes {
				if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvc.Name {
					return &deploy, nil
				}
			}
		}
	}
	return nil, nil
}

func (c *restarter) getSts(ctx context.Context, targetPvc *v1.PersistentVolumeClaim) (*appsV1.StatefulSet, error) {
	//get all sts
	stsList := &appsV1.StatefulSetList{}
	var opts []client.ListOption
	err := c.client.List(ctx, stsList, opts...)
	if err != nil {
		return nil, err
	}

	for _, sts := range stsList.Items {
		if len(sts.Spec.Template.Spec.Volumes) > 0 {
			for _, vol := range sts.Spec.Template.Spec.Volumes {
				if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == targetPvc.Name {
					return &sts, nil
				}
			}
		}
		for _, pvc := range sts.Spec.VolumeClaimTemplates {
			if pvc.Name == targetPvc.Name {
				return &sts, nil
			}
		}
	}

	return nil, fmt.Errorf("Cannot get deployment or statefulSet which pod mounted the pvc %s ", targetPvc.Name)
}

func (c *restarter) IfDeployTimeout(ctx context.Context, scName string, deploy *appsV1.Deployment) bool {
	sc := &storagev1.StorageClass{}
	err := c.client.Get(ctx, types.NamespacedName{Namespace: "", Name: scName}, sc)
	maxTime := 300
	if val, ok := sc.Annotations[ResizingMaxTime]; ok {
		userSetTime, err := strconv.Atoi(val)
		if err == nil {
			maxTime = userSetTime
		}
	}
	if _, ok := deploy.Annotations[RestartStopTime]; !ok {
		return false
	}
	startResizeTime, err := strconv.Atoi(deploy.Annotations[RestartStopTime])
	if err != nil {
		return true
	}
	timeout := int(time.Now().Unix())-startResizeTime > maxTime
	return timeout
}

func (c *restarter) IfStsTimeout(ctx context.Context, scName string, sts *appsV1.StatefulSet) bool {
	sc := &storagev1.StorageClass{}
	err := c.client.Get(ctx, types.NamespacedName{Namespace: "", Name: scName}, sc)
	maxTime := 300
	if val, ok := sc.Annotations[ResizingMaxTime]; ok {
		userSetTime, err := strconv.Atoi(val)
		if err == nil {
			maxTime = userSetTime
		}
	}
	if _, ok := sts.Annotations[RestartStopTime]; !ok {
		return false
	}
	startResizeTime, err := strconv.Atoi(sts.Annotations[RestartStopTime])
	if err != nil {
		return true
	}
	timeout := int(time.Now().Unix())-startResizeTime > maxTime
	return timeout
}

func (c *restarter) getAppList(ctx context.Context, pvcs []v1.PersistentVolumeClaim) (deployToStop []*appsV1.Deployment, stsToStop []*appsV1.StatefulSet, deployTimeout []*appsV1.Deployment, stsTimeout []*appsV1.StatefulSet, err error) {
	for _, pvc := range pvcs {
		dep, err := c.getDeploy(ctx, &pvc)
		if err != nil {
			return deployToStop, stsToStop, deployTimeout, stsTimeout, err
		}
		if dep != nil {
			if timeout := c.IfDeployTimeout(ctx, *pvc.Spec.StorageClassName, dep); timeout {
				deployTimeout = append(deployTimeout, dep)
			} else {
				deployToStop = append(deployToStop, dep)
			}
			continue
		}
		sts, stsErr := c.getSts(ctx, &pvc)
		if stsErr != nil {
			return deployToStop, stsToStop, deployTimeout, stsTimeout, err
		}
		if sts != nil {
			if timeout := c.IfStsTimeout(ctx, *pvc.Spec.StorageClassName, sts); timeout {
				stsTimeout = append(stsTimeout, sts)
			} else {
				stsToStop = append(stsToStop, sts)
			}
		}
	}
	return
}
