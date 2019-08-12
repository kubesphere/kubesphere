// Copyright 2019 The KubeSphere Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package models

var SearchWordColumnTable = []string{
	TableResourceType,
	TableResourceFilter,
	TableMetric,
	TablePolicy,
	TableRule,
	TableAlert,
	TableHistory,
	TableComment,
}

// columns that can be search through sql 'like' operator
var SearchColumns = map[string][]string{
	TableResourceType: {
		RtColId, RtColName,
	},
	TableResourceFilter: {
		RfColId, RfColName, RfColStatus, RfColTypeId,
	},
	TableMetric: {
		MtColId, MtColName, MtColStatus, MtColTypeId,
	},
	TablePolicy: {
		PlColId, PlColName, PlColDescription, PlColCreator, PlColTypeId,
	},
	TableRule: {
		RlColId, RlColName, RlColDisabled, RlColMonitorPeriods, RlColSeverity, RlColMetricsType, RlColConditionType, RlColThresholds, RlColUnit, RlColConsecutiveCount, RlColInhibit, RlColPolicyId, RlColMetricId,
	},
	TableAlert: {
		AlColId, AlColName, AlColDisabled, AlColRunningStatus, AlColPolicyId, AlColRsFilterId, AlColExecutorId,
	},
	TableHistory: {
		HsColId, HsColName, HsColEvent, HsColContent, HsColNotificationId, HsColAlertId, HsColRuleId, HsColResourceName,
	},
	TableComment: {
		CmColId, CmColAddresser, CmColContent, CmColHistoryId,
	},
	TableAction: {
		AcColId, AcColName, AcColTriggerStatus, AcColTriggerAction, AcColPolicyId, AcColNfAddressListId,
	},
}

// columns that can be search through sql '=' operator
var IndexedColumns = map[string][]string{
	TableResourceType: {
		RtColId, RtColName,
	},
	TableResourceFilter: {
		RfColId, RfColName, RfColStatus, RfColTypeId,
	},
	TableMetric: {
		MtColId, MtColName, MtColStatus, MtColTypeId,
	},
	TablePolicy: {
		PlColId, PlColName, PlColDescription, PlColCreator, PlColTypeId,
	},
	TableRule: {
		RlColId, RlColName, RlColDisabled, RlColMonitorPeriods, RlColSeverity, RlColMetricsType, RlColConditionType, RlColThresholds, RlColUnit, RlColConsecutiveCount, RlColInhibit, RlColPolicyId, RlColMetricId,
	},
	TableAlert: {
		AlColId, AlColName, AlColDisabled, AlColRunningStatus, AlColPolicyId, AlColRsFilterId, AlColExecutorId,
	},
	TableHistory: {
		HsColId, HsColName, HsColEvent, HsColContent, HsColNotificationId, HsColAlertId, HsColRuleId, HsColResourceName,
	},
	TableComment: {
		CmColId, CmColAddresser, CmColContent, CmColHistoryId,
	},
	TableAction: {
		AcColId, AcColName, AcColTriggerStatus, AcColTriggerAction, AcColPolicyId, AcColNfAddressListId,
	},
}
