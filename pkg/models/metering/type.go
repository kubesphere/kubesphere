package metering

type PriceInfo struct {
	Currency                                  string  `json:"currency" description:"currency"`
	CpuPerCorePerHour                         float64 `json:"cpu_per_core_per_hour,omitempty" description:"cpu price"`
	MemPerGigabytesPerHour                    float64 `json:"mem_per_gigabytes_per_hour,omitempty" description:"mem price"`
	IngressNetworkTrafficPerGiagabytesPerHour float64 `json:"ingress_network_traffic_per_giagabytes_per_hour,omitempty" description:"ingress price"`
	EgressNetworkTrafficPerGiagabytesPerHour  float64 `json:"egress_network_traffic_per_gigabytes_per_hour,omitempty" description:"egress price"`
	PvcPerGigabytesPerHour                    float64 `json:"pvc_per_gigabytes_per_hour,omitempty" description:"pvc price"`
}

// currently init method fill illegal value to hint that metering config file was not mounted yet
func (p *PriceInfo) Init() {
	p.Currency = ""
	p.CpuPerCorePerHour = -1
	p.MemPerGigabytesPerHour = -1
	p.IngressNetworkTrafficPerGiagabytesPerHour = -1
	p.EgressNetworkTrafficPerGiagabytesPerHour = -1
	p.PvcPerGigabytesPerHour = -1
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

type AppStatistic struct {
	CPUUsage            float64                      `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                      `json:"memory_usage_wo_cache" description:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                      `json:"net_bytes_transmitted" description:"net_bytes_transmitted"`
	NetBytesReceived    float64                      `json:"net_bytes_received" description:"net_bytes_received"`
	PVCBytesTotal       float64                      `json:"pvc_bytes_total" description:"pvc_bytes_total"`
	Services            map[string]*ServiceStatistic `json:"services" description:"services"`
}

func (as *AppStatistic) GetServiceStats(name string) *ServiceStatistic {
	if as.Services == nil {
		as.Services = make(map[string]*ServiceStatistic)
	}
	if as.Services[name] == nil {
		as.Services[name] = &ServiceStatistic{}
	}
	return as.Services[name]
}

func (as *AppStatistic) Aggregate() {
	if as.Services == nil {
		return
	}

	// remove duplicate pods which were selected by different svc
	podsMap := make(map[string]struct{})
	for _, svcObj := range as.Services {
		for podName, podObj := range svcObj.Pods {
			if _, ok := podsMap[podName]; ok {
				continue
			} else {
				podsMap[podName] = struct{}{}
			}
			as.CPUUsage += podObj.CPUUsage
			as.MemoryUsageWoCache += podObj.MemoryUsageWoCache
			as.NetBytesTransmitted += podObj.NetBytesTransmitted
			as.NetBytesReceived += podObj.NetBytesReceived
			as.PVCBytesTotal += podObj.PVCBytesTotal
		}
	}
}

type ServiceStatistic struct {
	CPUUsage            float64                  `json:"cpu_usage" description:"cpu_usage"`
	MemoryUsageWoCache  float64                  `json:"memory_usage_wo_cache" desription:"memory_usage_wo_cache"`
	NetBytesTransmitted float64                  `json:"net_bytes_transmitted" description:"net_bytes_transmitted"`
	NetBytesReceived    float64                  `json:"net_bytes_received" description:"net_bytes_received"`
	Pods                map[string]*PodStatistic `json:"pods" description:"pod statistic"`
}

func (ss *ServiceStatistic) SetPodStats(name string, podStat *PodStatistic) error {
	if ss.Pods == nil {
		ss.Pods = make(map[string]*PodStatistic)
	}
	ss.Pods[name] = podStat
	return nil
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

func (ds *DeploymentStatistic) SetPodStats(name string, podStat *PodStatistic) error {
	if ds.Pods == nil {
		ds.Pods = make(map[string]*PodStatistic)
	}
	ds.Pods[name] = podStat
	return nil
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

func (ss *StatefulsetStatistic) SetPodStats(name string, podStat *PodStatistic) error {
	if ss.Pods == nil {
		ss.Pods = make(map[string]*PodStatistic)
	}
	ss.Pods[name] = podStat
	return nil
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

func (ds *DaemonsetStatistic) SetPodStats(name string, podStat *PodStatistic) error {
	if ds.Pods == nil {
		ds.Pods = make(map[string]*PodStatistic)
	}
	ds.Pods[name] = podStat
	return nil
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
	Apps         map[string]*AppStatistic         `json:"apps" description:"app statistic"`
	Services     map[string]*ServiceStatistic     `json:"services" description:"service statistic"`
	Deploys      map[string]*DeploymentStatistic  `json:"deployments" description:"deployment statistic"`
	Statefulsets map[string]*StatefulsetStatistic `json:"statefulsets" description:"statefulset statistic"`
	Daemonsets   map[string]*DaemonsetStatistic   `json:"daemonsets" description:"daemonsets statistics"`
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

func (rs *ResourceStatistic) GetServiceStats(name string) *ServiceStatistic {
	if rs.Services == nil {
		rs.Services = make(map[string]*ServiceStatistic)
	}
	if rs.Services[name] == nil {
		rs.Services[name] = &ServiceStatistic{}
	}
	return rs.Services[name]
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
