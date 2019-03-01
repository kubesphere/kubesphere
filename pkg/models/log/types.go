package log

import (
	"kubesphere.io/kubesphere/pkg/client"
	"time"
)

type FluentbitCRDResult struct {
	Status int                  `json:"status"`
	CRD    client.FluentBitSpec `json:"CRD,omitempty"`
}

type FluentbitCRDDeleteResult struct {
	Status int `json:"status"`
}

type FluentbitSettingsResult struct {
	Status int    `json:"status"`
	Enable string `json:"Enable,omitempty"`
}

type FluentbitFilter struct {
	Type       string `json:"type"`
	Field      string `json:"field"`
	Expression string `json:"expression"`
}

type FluentbitFiltersResult struct {
	Status  int               `json:"status"`
	Filters []FluentbitFilter `json:"filters,omitempty"`
}

type FluentbitOutputsResult struct {
	Status  int                   `json:"status"`
	Outputs []client.OutputPlugin `json:"outputs,omitempty"`
}

type OutputDBBinding struct {
	Id         uint   `gorm:"primary_key;auto_increment;unique"`
	Type       string `gorm:"not null"`
	Name       string `gorm:"not null"`
	Parameters string `gorm:"not null"`
	Internal   bool
	Enable     bool      `gorm:"not null"`
	Updatetime time.Time `gorm:"not null"`
}
