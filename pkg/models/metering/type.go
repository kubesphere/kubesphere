// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package metering

type PriceInfo struct {
	// currency unit, currently support CNY and USD
	Currency string `json:"currency" description:"currency"`
	// cpu cost with above currency unit for per core per hour
	CpuPerCorePerHour float64 `json:"cpu_per_core_per_hour,omitempty" description:"cpu price"`
	// mem cost with above currency unit for per GB per hour
	MemPerGigabytesPerHour float64 `json:"mem_per_gigabytes_per_hour,omitempty" description:"mem price"`
	// ingress network traffic cost with above currency unit for per MB per hour
	IngressNetworkTrafficPerMegabytesPerHour float64 `json:"ingress_network_traffic_per_megabytes_per_hour,omitempty" description:"ingress price"`
	// egress network traffice cost with above currency unit for per MB per hour
	EgressNetworkTrafficPerMegabytesPerHour float64 `json:"egress_network_traffic_per_megabytes_per_hour,omitempty" description:"egress price"`
	// pvc cost with above currency unit for per GB per hour
	PvcPerGigabytesPerHour float64 `json:"pvc_per_gigabytes_per_hour,omitempty" description:"pvc price"`
}

type PriceResponse struct {
	RetentionDay string `json:"retention_day"`
	PriceInfo    `json:",inline"`
}

type PodStatistic struct {
	CPUUsage            float64 `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64 `json:"memory_usage_wo_cache" description:"memory_usage_wo_cache"`
	NetBytesTransmitted float64 `json:"net_bytes_transmitted" desription:"net_bytes_transmitted"`
	NetBytesReceived    float64 `json:"net_bytes_received" description:"net_bytes_received"`
	PVCBytesTotal       float64 `json:"pvc_bytes_total" description:"pvc_bytes_total"`
}

type PodsStats map[string]*PodStatistic

func (ps *PodsStats) Set(podName, meterName string, value float64) {
	if _, ok := (*ps)[podName]; !ok {
		(*ps)[podName] = &PodStatistic{}
	}
	switch meterName {
	case "meter_pod_cpu_usage":
		(*ps)[podName].CPUUsage = value
	case "meter_pod_memory_usage_wo_cache":
		(*ps)[podName].MemoryUsageWoCache = value
	case "meter_pod_net_bytes_transmitted":
		(*ps)[podName].NetBytesTransmitted = value
	case "meter_pod_net_bytes_received":
		(*ps)[podName].NetBytesReceived = value
	case "meter_pod_pvc_bytes_total":
		(*ps)[podName].PVCBytesTotal = value
	}
}

type OpenPitrixStatistic struct {
	AppStatistic
}

type AppStatistic struct {
	CPUUsage            float64                          `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                          `json:"memory_usage_wo_cache" description:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                          `json:"net_bytes_transmitted" description:"net_bytes_transmitted"`
	NetBytesReceived    float64                          `json:"net_bytes_received" description:"net_bytes_received"`
	PVCBytesTotal       float64                          `json:"pvc_bytes_total" description:"pvc_bytes_total"`
	Deploys             map[string]*DeploymentStatistic  `json:"deployments" description:"deployment statistic"`
	Statefulsets        map[string]*StatefulsetStatistic `json:"statefulsets" description:"statefulset statistic"`
	Daemonsets          map[string]*DaemonsetStatistic   `json:"daemonsets" description:"daemonsets statistics"`
}

func (as *AppStatistic) GetDeployStats(name string) *DeploymentStatistic {
	if as.Deploys == nil {
		as.Deploys = make(map[string]*DeploymentStatistic)
	}
	if as.Deploys[name] == nil {
		as.Deploys[name] = &DeploymentStatistic{}
	}
	return as.Deploys[name]
}

func (as *AppStatistic) GetDaemonStats(name string) *DaemonsetStatistic {
	if as.Daemonsets == nil {
		as.Daemonsets = make(map[string]*DaemonsetStatistic)
	}
	if as.Daemonsets[name] == nil {
		as.Daemonsets[name] = &DaemonsetStatistic{}
	}
	return as.Daemonsets[name]
}

func (as *AppStatistic) GetStatefulsetStats(name string) *StatefulsetStatistic {
	if as.Statefulsets == nil {
		as.Statefulsets = make(map[string]*StatefulsetStatistic)
	}
	if as.Statefulsets[name] == nil {
		as.Statefulsets[name] = &StatefulsetStatistic{}
	}
	return as.Statefulsets[name]
}

func (as *AppStatistic) Aggregate() {
	if as.Deploys == nil && as.Statefulsets == nil && as.Daemonsets == nil {
		return
	}

	// aggregate deployment stats
	for _, deployObj := range as.Deploys {
		for _, podObj := range deployObj.Pods {
			deployObj.CPUUsage += podObj.CPUUsage
			deployObj.MemoryUsageWoCache += podObj.MemoryUsageWoCache
			deployObj.NetBytesTransmitted += podObj.NetBytesTransmitted
			deployObj.NetBytesReceived += podObj.NetBytesReceived
			deployObj.PVCBytesTotal += podObj.PVCBytesTotal
		}
		as.CPUUsage += deployObj.CPUUsage
		as.MemoryUsageWoCache += deployObj.MemoryUsageWoCache
		as.NetBytesTransmitted += deployObj.NetBytesTransmitted
		as.NetBytesReceived += deployObj.NetBytesReceived
		as.PVCBytesTotal += deployObj.PVCBytesTotal
	}

	// aggregate statfulset stats
	for _, statfulObj := range as.Statefulsets {
		for _, podObj := range statfulObj.Pods {
			statfulObj.CPUUsage += podObj.CPUUsage
			statfulObj.MemoryUsageWoCache += podObj.MemoryUsageWoCache
			statfulObj.NetBytesTransmitted += podObj.NetBytesTransmitted
			statfulObj.NetBytesReceived += podObj.NetBytesReceived
			statfulObj.PVCBytesTotal += podObj.PVCBytesTotal
		}
		as.CPUUsage += statfulObj.CPUUsage
		as.MemoryUsageWoCache += statfulObj.MemoryUsageWoCache
		as.NetBytesTransmitted += statfulObj.NetBytesTransmitted
		as.NetBytesReceived += statfulObj.NetBytesReceived
		as.PVCBytesTotal += statfulObj.PVCBytesTotal
	}

	// aggregate daemonset stats
	for _, daemonsetObj := range as.Daemonsets {
		for _, podObj := range daemonsetObj.Pods {
			daemonsetObj.CPUUsage += podObj.CPUUsage
			daemonsetObj.MemoryUsageWoCache += podObj.MemoryUsageWoCache
			daemonsetObj.NetBytesTransmitted += podObj.NetBytesTransmitted
			daemonsetObj.NetBytesReceived += podObj.NetBytesReceived
			daemonsetObj.PVCBytesTotal += podObj.PVCBytesTotal
		}
		as.CPUUsage += daemonsetObj.CPUUsage
		as.MemoryUsageWoCache += daemonsetObj.MemoryUsageWoCache
		as.NetBytesTransmitted += daemonsetObj.NetBytesTransmitted
		as.NetBytesReceived += daemonsetObj.NetBytesReceived
		as.PVCBytesTotal += daemonsetObj.PVCBytesTotal
	}

	return
}

type ServiceStatistic struct {
	CPUUsage            float64                  `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                  `json:"memory_usage_wo_cache" desription:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                  `json:"net_bytes_transmitted" description:"net_bytes_transmitted"`
	NetBytesReceived    float64                  `json:"net_bytes_received" description:"net_bytes_received"`
	Pods                map[string]*PodStatistic `json:"pods" description:"pod statistic"`
}

func (ss *ServiceStatistic) SetPodStats(name string, podStat *PodStatistic) {
	if ss.Pods == nil {
		ss.Pods = make(map[string]*PodStatistic)
	}
	ss.Pods[name] = podStat
}

func (ss *ServiceStatistic) GetPodStats(name string) *PodStatistic {
	if ss.Pods == nil {
		ss.Pods = make(map[string]*PodStatistic)
	}
	if ss.Pods[name] == nil {
		ss.Pods[name] = &PodStatistic{}
	}
	return ss.Pods[name]
}

func (ss *ServiceStatistic) Aggregate() {
	if ss.Pods == nil {
		return
	}

	for key := range ss.Pods {
		ss.CPUUsage += ss.GetPodStats(key).CPUUsage
		ss.MemoryUsageWoCache += ss.GetPodStats(key).MemoryUsageWoCache
		ss.NetBytesTransmitted += ss.GetPodStats(key).NetBytesTransmitted
		ss.NetBytesReceived += ss.GetPodStats(key).NetBytesReceived
	}
}

type DeploymentStatistic struct {
	CPUUsage            float64                  `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                  `json:"memory_usage_wo_cache" description:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                  `json:"net_bytes_transmitted" desciption:"net_bytes_transmitted"`
	NetBytesReceived    float64                  `json:"net_bytes_received" description:"net_bytes_received"`
	PVCBytesTotal       float64                  `json:"pvc_bytes_total" description:"pvc_bytes_total"`
	Pods                map[string]*PodStatistic `json:"pods" description:"pod statistic"`
}

func (ds *DeploymentStatistic) GetPodStats(name string) *PodStatistic {
	if ds.Pods == nil {
		ds.Pods = make(map[string]*PodStatistic)
	}
	if ds.Pods[name] == nil {
		ds.Pods[name] = &PodStatistic{}
	}
	return ds.Pods[name]
}

func (ds *DeploymentStatistic) SetPodStats(name string, podStat *PodStatistic) {
	if ds.Pods == nil {
		ds.Pods = make(map[string]*PodStatistic)
	}
	ds.Pods[name] = podStat
}

func (ds *DeploymentStatistic) Aggregate() {
	if ds.Pods == nil {
		return
	}

	for key := range ds.Pods {
		ds.CPUUsage += ds.GetPodStats(key).CPUUsage
		ds.MemoryUsageWoCache += ds.GetPodStats(key).MemoryUsageWoCache
		ds.NetBytesTransmitted += ds.GetPodStats(key).NetBytesTransmitted
		ds.NetBytesReceived += ds.GetPodStats(key).NetBytesReceived
		ds.PVCBytesTotal += ds.GetPodStats(key).PVCBytesTotal
	}
}

type StatefulsetStatistic struct {
	CPUUsage            float64                  `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                  `json:"memory_usage_wo_cache" description:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                  `json:"net_bytes_transmitted" description:"net_bytes_transmitted"`
	NetBytesReceived    float64                  `json:"net_bytes_received" description:"net_bytes_received"`
	PVCBytesTotal       float64                  `json:"pvc_bytes_total" description:"pvc_bytes_total"`
	Pods                map[string]*PodStatistic `json:"pods" description:"pod statistic"`
}

func (ss *StatefulsetStatistic) GetPodStats(name string) *PodStatistic {
	if ss.Pods == nil {
		ss.Pods = make(map[string]*PodStatistic)
	}
	if ss.Pods[name] == nil {
		ss.Pods[name] = &PodStatistic{}
	}
	return ss.Pods[name]
}

func (ss *StatefulsetStatistic) SetPodStats(name string, podStat *PodStatistic) {
	if ss.Pods == nil {
		ss.Pods = make(map[string]*PodStatistic)
	}
	ss.Pods[name] = podStat
}

func (ss *StatefulsetStatistic) Aggregate() {
	if ss.Pods == nil {
		return
	}

	for key := range ss.Pods {
		ss.CPUUsage += ss.GetPodStats(key).CPUUsage
		ss.MemoryUsageWoCache += ss.GetPodStats(key).MemoryUsageWoCache
		ss.NetBytesTransmitted += ss.GetPodStats(key).NetBytesTransmitted
		ss.NetBytesReceived += ss.GetPodStats(key).NetBytesReceived
		ss.PVCBytesTotal += ss.GetPodStats(key).PVCBytesTotal
	}
}

type DaemonsetStatistic struct {
	CPUUsage            float64                  `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                  `json:"memory_usage_wo_cache" description:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                  `json:"net_bytes_transmitted" description:"net_bytes_transmitted"`
	NetBytesReceived    float64                  `json:"net_bytes_received" description:"net_bytes_received"`
	PVCBytesTotal       float64                  `json:"pvc_bytes_total" description:"pvc_bytes_total"`
	Pods                map[string]*PodStatistic `json:"pods" description:"pod statistic"`
}

func (ds *DaemonsetStatistic) GetPodStats(name string) *PodStatistic {
	if ds.Pods == nil {
		ds.Pods = make(map[string]*PodStatistic)
	}
	if ds.Pods[name] == nil {
		ds.Pods[name] = &PodStatistic{}
	}
	return ds.Pods[name]
}

func (ds *DaemonsetStatistic) SetPodStats(name string, podStat *PodStatistic) {
	if ds.Pods == nil {
		ds.Pods = make(map[string]*PodStatistic)
	}
	ds.Pods[name] = podStat
}

func (ds *DaemonsetStatistic) Aggregate() {
	if ds.Pods == nil {
		return
	}
	for key := range ds.Pods {
		ds.CPUUsage += ds.GetPodStats(key).CPUUsage
		ds.MemoryUsageWoCache += ds.GetPodStats(key).MemoryUsageWoCache
		ds.NetBytesTransmitted += ds.GetPodStats(key).NetBytesTransmitted
		ds.NetBytesReceived += ds.GetPodStats(key).NetBytesReceived
		ds.PVCBytesTotal += ds.GetPodStats(key).PVCBytesTotal
	}
}

type ResourceStatistic struct {
	// openpitrix statistic
	OpenPitrixs map[string]*OpenPitrixStatistic `json:"openpitrixs" description:"openpitrix statistic"`

	// app crd statistic
	Apps map[string]*AppStatistic `json:"apps" description:"app statistic"`

	// k8s workload only which exclude app and op
	Deploys      map[string]*DeploymentStatistic  `json:"deployments" description:"deployment statistic"`
	Statefulsets map[string]*StatefulsetStatistic `json:"statefulsets" description:"statefulset statistic"`
	Daemonsets   map[string]*DaemonsetStatistic   `json:"daemonsets" description:"daemonsets statistics"`
}

func (rs *ResourceStatistic) GetOpenPitrixStats(name string) *OpenPitrixStatistic {
	if rs.OpenPitrixs == nil {
		rs.OpenPitrixs = make(map[string]*OpenPitrixStatistic)
	}
	if rs.OpenPitrixs[name] == nil {
		rs.OpenPitrixs[name] = &OpenPitrixStatistic{}
	}
	return rs.OpenPitrixs[name]
}

func (rs *ResourceStatistic) GetAppStats(name string) *AppStatistic {
	if rs.Apps == nil {
		rs.Apps = make(map[string]*AppStatistic)
	}
	if rs.Apps[name] == nil {
		rs.Apps[name] = &AppStatistic{}
	}
	return rs.Apps[name]
}

func (rs *ResourceStatistic) GetDeployStats(name string) *DeploymentStatistic {
	if rs.Deploys == nil {
		rs.Deploys = make(map[string]*DeploymentStatistic)
	}
	if rs.Deploys[name] == nil {
		rs.Deploys[name] = &DeploymentStatistic{}
	}
	return rs.Deploys[name]
}

func (rs *ResourceStatistic) GetStatefulsetStats(name string) *StatefulsetStatistic {
	if rs.Statefulsets == nil {
		rs.Statefulsets = make(map[string]*StatefulsetStatistic)
	}
	if rs.Statefulsets[name] == nil {
		rs.Statefulsets[name] = &StatefulsetStatistic{}
	}
	return rs.Statefulsets[name]
}

func (rs *ResourceStatistic) GetDaemonsetStats(name string) *DaemonsetStatistic {
	if rs.Daemonsets == nil {
		rs.Daemonsets = make(map[string]*DaemonsetStatistic)
	}
	if rs.Daemonsets[name] == nil {
		rs.Daemonsets[name] = &DaemonsetStatistic{}
	}
	return rs.Daemonsets[name]
}
