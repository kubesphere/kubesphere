package openpitrix

import (
	"fmt"
	"net/http"
	"time"
)

type Cluster struct {
	ClusterID       string        `json:"cluster_id"`
	Name            string        `json:"name"`
	AppID           string        `json:"app_id"`
	VersionID       string        `json:"version_id"`
	Status          string        `json:"status"`
	UpdateTime      time.Time     `json:"status_time"`
	CreateTime      time.Time     `json:"create_time"`
	RunTimeId       string        `json:"runtime_id"`
	Description     string        `json:"description"`
	ClusterRoleSets []ClusterRole `json:"cluster_role_set"`
}

type ClusterRole struct {
	ClusterID string `json:"cluster_id"`
	Role      string `json:"role"`
}

type ClusterList struct {
	Total    int       `json:"total_count"`
	Clusters []Cluster `json:"cluster_set"`
}

type VersionList struct {
	Total    int       `json:"total_count"`
	Versions []version `json:"app_version_set"`
}

type version struct {
	Name      string `json:"name"`
	VersionID string `json:"version_id"`
}

type runtime struct {
	RuntimeID string `json:"runtime_id"`
	Zone      string `json:"zone"`
}

type runtimeList struct {
	Total    int       `json:"total_count"`
	Runtimes []runtime `json:"runtime_set"`
}

type app struct {
	AppId     string `json:"app_id"`
	Name      string `json:"name"`
	ChartName string `json:"chart_name"`
	RepoId    string `json:"repo_id"`
}

type repo struct {
	RepoId string `json:"repo_id"`
	Name   string `json:"name"`
	Url    string `json:"url"`
}

type appList struct {
	Total int   `json:"total_count"`
	Apps  []app `json:"app_set"`
}

type repoList struct {
	Total int    `json:"total_count"`
	Repos []repo `json:"repo_set"`
}

type CreateClusterRequest struct {
	AppId     string `json:"app_id" description:"ID of app to run in cluster, e.g. app-AA3A3y3zEgEM"`
	VersionId string `json:"version_id" description:"app version, e.g. appv-154gXYx5RKRp"`
	RuntimeId string `json:"runtime_id" description:"ID of runtime, e.g. runtime-wWwXL0LzWqEr"`
	Conf      string `json:"conf" description:"conf a json string, include cpu, memory info of cluster"`
}

type DeleteClusterRequest struct {
	ClusterId []string `json:"cluster_id" description:"cluster ID"`
}

type RunTime struct {
	RuntimeId         string `json:"runtime_id"`
	RuntimeUrl        string `json:"runtime_url"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	Zone              string `json:"zone"`
	RuntimeCredential string `json:"runtime_credential"`
}

type Interface interface {
	CreateRuntime(runtime *RunTime) error
	DeleteRuntime(runtimeId string) error
}
type cluster struct {
	Status    string `json:"status"`
	ClusterId string `json:"cluster_id"`
}

type Error struct {
	status  int
	message string
}

func (e Error) Error() string {
	return fmt.Sprintf("status: %d,message: %s", e.status, e.message)
}

type OpenPitrixClient struct {
	client    *http.Client
	apiServer string
	token     string
}
