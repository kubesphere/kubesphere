/*
Copyright 2020 The KubeSphere Authors.
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

package openpitrix

import (
	"github.com/go-openapi/strfmt"
	"openpitrix.io/openpitrix/pkg/pb"
	"time"
)

func convertApp(in *pb.App) *App {

	if in == nil {
		return nil
	}

	categorySet := make(AppCategorySet, 0)

	for _, item := range in.CategorySet {
		category := convertResourceCategory(item)
		categorySet = append(categorySet, category)
	}

	out := App{
		CategorySet: categorySet,
	}

	if in.Abstraction != nil {
		out.Abstraction = in.Abstraction.Value
	}
	if in.Active != nil {
		out.Active = in.Active.Value
	}
	if in.AppId != nil {
		out.AppId = in.AppId.Value
	}
	if in.AppVersionTypes != nil {
		out.AppVersionTypes = in.AppVersionTypes.Value
	}
	if in.ChartName != nil {
		out.ChartName = in.ChartName.Value
	}
	if in.CompanyJoinTime != nil {
		date := strfmt.DateTime(time.Unix(in.CompanyJoinTime.Seconds, 0))
		out.CompanyJoinTime = &date
	}
	if in.CompanyName != nil {
		out.CompanyName = in.CompanyName.Value
	}
	if in.CompanyProfile != nil {
		out.CompanyProfile = in.CompanyProfile.Value
	}
	if in.CompanyWebsite != nil {
		out.CompanyWebsite = in.CompanyWebsite.Value
	}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.CompanyWebsite != nil {
		out.CompanyWebsite = in.CompanyWebsite.Value
	}
	if in.Description != nil {
		out.Description = in.Description.Value
	}
	if in.Home != nil {
		out.Home = in.Home.Value
	}
	if in.Icon != nil {
		out.Icon = in.Icon.Value
	}
	if in.Isv != nil {
		out.Isv = in.Isv.Value
	}
	if in.Keywords != nil {
		out.Keywords = in.Keywords.Value
	}
	if in.LatestAppVersion != nil {
		out.LatestAppVersion = convertAppVersion(in.LatestAppVersion)
	}
	if in.Name != nil {
		out.Name = in.Name.Value
	}
	if in.Owner != nil {
		out.Owner = in.Owner.Value
	}
	if in.Readme != nil {
		out.Readme = in.Readme.Value
	}
	if in.RepoId != nil {
		out.RepoId = in.RepoId.Value
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.Sources != nil {
		out.Sources = in.Sources.Value
	}
	if in.Screenshots != nil {
		out.Screenshots = in.Screenshots.Value
	}
	if in.Tos != nil {
		out.Tos = in.Tos.Value
	}
	if in.UpdateTime != nil {
		date := strfmt.DateTime(time.Unix(in.UpdateTime.Seconds, 0))
		out.UpdateTime = &date
	}

	return &out
}

func convertAppVersion(in *pb.AppVersion) *AppVersion {
	if in == nil {
		return nil
	}
	out := AppVersion{}
	if in.AppId != nil {
		out.AppId = in.AppId.Value
	}
	if in.Active != nil {
		out.Active = in.Active.Value
	}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.Description != nil {
		out.Description = in.Description.Value
	}
	if in.Home != nil {
		out.Home = in.Home.Value
	}
	if in.Icon != nil {
		out.Icon = in.Icon.Value
	}
	if in.Maintainers != nil {
		out.Maintainers = in.Maintainers.Value
	}
	if in.Message != nil {
		out.Message = in.Message.Value
	}
	if in.Keywords != nil {
		out.Keywords = in.Keywords.Value
	}
	if in.Name != nil {
		out.Name = in.Name.Value
	}
	if in.Owner != nil {
		out.Owner = in.Owner.Value
	}
	if in.PackageName != nil {
		out.PackageName = in.PackageName.Value
	}
	if in.Readme != nil {
		out.Readme = in.Readme.Value
	}
	if in.ReviewId != nil {
		out.ReviewId = in.ReviewId.Value
	}
	if in.Screenshots != nil {
		out.Screenshots = in.Screenshots.Value
	}
	if in.Sources != nil {
		out.Sources = in.Sources.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.Sequence != nil {
		out.Sequence = int64(in.Sequence.Value)
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}
	if in.Type != nil {
		out.Type = in.Type.Value
	}
	if in.UpdateTime != nil {
		date := strfmt.DateTime(time.Unix(in.UpdateTime.Seconds, 0))
		out.UpdateTime = &date
	}
	if in.VersionId != nil {
		out.VersionId = in.VersionId.Value
	}

	return &out

}

func convertResourceCategory(in *pb.ResourceCategory) *ResourceCategory {
	if in == nil {
		return nil
	}
	out := ResourceCategory{}

	if in.CategoryId != nil {
		out.CategoryId = in.CategoryId.Value
	}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.Locale != nil {
		out.Locale = in.Locale.Value
	}
	if in.Name != nil {
		out.Name = in.Name.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}

	return &out
}

func convertCategory(in *pb.Category) *Category {
	if in == nil {
		return nil
	}
	out := Category{}

	if in.CategoryId != nil {
		out.CategoryID = in.CategoryId.Value
	}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.Locale != nil {
		out.Locale = in.Locale.Value
	}
	if in.Name != nil {
		out.Name = in.Name.Value
	}
	if in.Description != nil {
		out.Description = in.Description.Value
	}
	if in.Icon != nil {
		out.Icon = in.Icon.Value
	}
	if in.Owner != nil {
		out.Owner = in.Owner.Value
	}
	if in.UpdateTime != nil {
		date := strfmt.DateTime(time.Unix(in.UpdateTime.Seconds, 0))
		out.UpdateTime = &date
	}

	return &out
}

func convertAttachment(in *pb.Attachment) *Attachment {
	if in == nil {
		return nil
	}
	out := Attachment{}

	out.AttachmentID = in.AttachmentId

	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.AttachmentContent != nil {
		out.AttachmentContent = make(map[string]strfmt.Base64)
		for k, v := range in.AttachmentContent {
			out.AttachmentContent[k] = v
		}
	}

	return &out
}

func convertRepo(in *pb.Repo) *Repo {
	if in == nil {
		return nil
	}
	out := Repo{}

	if in.RepoId != nil {
		out.RepoId = in.RepoId.Value
	}
	if in.Name != nil {
		out.Name = in.Name.Value
	}
	if in.AppDefaultStatus != nil {
		out.AppDefaultStatus = in.AppDefaultStatus.Value
	}
	if in.Credential != nil {
		out.Credential = in.Credential.Value
	}

	categorySet := make(RepoCategorySet, 0)

	for _, item := range in.CategorySet {
		category := convertResourceCategory(item)
		categorySet = append(categorySet, category)
	}

	out.CategorySet = categorySet

	if in.Controller != nil {
		out.Credential = in.Credential.Value
	}

	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}

	if in.Description != nil {
		out.Description = in.Description.Value
	}

	labelSet := make(RepoLabels, 0)

	for _, item := range in.Labels {
		label := convertRepoLabel(item)
		labelSet = append(labelSet, label)
	}

	out.Labels = labelSet

	if in.Owner != nil {
		out.Owner = in.Owner.Value
	}
	if in.Providers != nil {
		out.Providers = in.Providers
	}
	if in.RepoId != nil {
		out.RepoId = in.RepoId.Value
	}

	selectorSet := make(RepoSelectors, 0)

	for _, item := range in.Selectors {
		selector := convertRepoSelector(item)
		selectorSet = append(selectorSet, selector)
	}

	out.Selectors = selectorSet

	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		out.StatusTime = strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
	}
	if in.Type != nil {
		out.Type = in.Type.Value
	}
	if in.Url != nil {
		out.URL = in.Url.Value
	}
	if in.Visibility != nil {
		out.Visibility = in.Visibility.Value
	}
	return &out
}

func convertRepoLabel(in *pb.RepoLabel) *RepoLabel {
	if in == nil {
		return nil
	}
	out := RepoLabel{}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.LabelKey != nil {
		out.LabelKey = in.LabelKey.Value
	}
	if in.LabelValue != nil {
		out.LabelValue = in.LabelValue.Value
	}
	return &out
}

func convertRepoSelector(in *pb.RepoSelector) *RepoSelector {
	if in == nil {
		return nil
	}
	out := RepoSelector{}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.SelectorKey != nil {
		out.SelectorKey = in.SelectorKey.Value
	}
	if in.SelectorValue != nil {
		out.SelectorValue = in.SelectorValue.Value
	}
	return &out
}

func convertAppVersionAudit(in *pb.AppVersionAudit) *AppVersionAudit {
	if in == nil {
		return nil
	}
	out := AppVersionAudit{}
	if in.AppId != nil {
		out.AppId = in.AppId.Value
	}
	if in.AppName != nil {
		out.AppName = in.AppName.Value
	}
	if in.Message != nil {
		out.Message = in.Message.Value
	}
	if in.Operator != nil {
		out.Operator = in.Operator.Value
	}
	if in.OperatorType != nil {
		out.OperatorType = in.OperatorType.Value
	}
	if in.ReviewId != nil {
		out.ReviewId = in.ReviewId.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}
	if in.VersionId != nil {
		out.VersionId = in.VersionId.Value
	}
	if in.VersionName != nil {
		out.VersionName = in.VersionName.Value
	}
	if in.VersionType != nil {
		out.VersionType = in.VersionType.Value
	}
	return &out
}

func convertAppVersionReview(in *pb.AppVersionReview) *AppVersionReview {
	if in == nil {
		return nil
	}
	out := AppVersionReview{}
	if in.AppId != nil {
		out.AppId = in.AppId.Value
	}
	if in.AppName != nil {
		out.AppName = in.AppName.Value
	}
	if in.Phase != nil {
		out.Phase = make(AppVersionReviewPhaseOAIGen)
		for k, v := range in.Phase {
			out.Phase[k] = *convertAppVersionReviewPhase(v)
		}
	}
	if in.ReviewId != nil {
		out.ReviewId = in.ReviewId.Value
	}
	if in.Reviewer != nil {
		out.Reviewer = in.Reviewer.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		out.StatusTime = strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
	}
	if in.VersionId != nil {
		out.VersionID = in.VersionId.Value
	}
	if in.VersionName != nil {
		out.VersionName = in.VersionName.Value
	}
	if in.VersionType != nil {
		out.VersionType = in.VersionType.Value
	}
	return &out
}

func convertAppVersionReviewPhase(in *pb.AppVersionReviewPhase) *AppVersionReviewPhase {
	if in == nil {
		return nil
	}
	out := AppVersionReviewPhase{}
	if in.Message != nil {
		out.Message = in.Message.Value
	}
	if in.OperatorType != nil {
		out.OperatorType = in.OperatorType.Value
	}
	if in.ReviewTime != nil {
		date := strfmt.DateTime(time.Unix(in.ReviewTime.Seconds, 0))
		out.ReviewTime = &date
	}
	if in.Operator != nil {
		out.Operator = in.Operator.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}
	return &out
}

func convertRepoEvent(in *pb.RepoEvent) *RepoEvent {
	if in == nil {
		return nil
	}
	out := RepoEvent{}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.Owner != nil {
		out.Owner = in.Owner.Value
	}
	if in.RepoEventId != nil {
		out.RepoEventId = in.RepoEventId.Value
	}
	if in.RepoId != nil {
		out.RepoId = in.RepoId.Value
	}
	if in.Result != nil {
		out.Result = in.Result.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}

	return &out
}

func convertCluster(in *pb.Cluster) *Cluster {
	if in == nil {
		return nil
	}
	out := Cluster{}
	if in.AdditionalInfo != nil {
		out.AdditionalInfo = in.AdditionalInfo.Value
	}
	if in.AppId != nil {
		out.AppId = in.AppId.Value
	}
	if in.ClusterId != nil {
		out.ClusterId = in.ClusterId.Value
	}
	if in.ClusterType != nil {
		out.ClusterType = int64(in.ClusterType.Value)
	}
	if in.CreateTime != nil {
		date := strfmt.DateTime(time.Unix(in.CreateTime.Seconds, 0))
		out.CreateTime = &date
	}
	if in.Debug != nil {
		out.Debug = in.Debug.Value
	}
	if in.Description != nil {
		out.Description = in.Description.Value
	}
	if in.Endpoints != nil {
		out.Endpoints = in.Endpoints.Value
	}
	if in.Env != nil {
		out.Env = in.Env.Value
	}
	if in.FrontgateId != nil {
		out.FrontgateId = in.FrontgateId.Value
	}
	if in.GlobalUuid != nil {
		out.GlobalUUID = in.GlobalUuid.Value
	}
	if in.MetadataRootAccess != nil {
		out.MetadataRootAccess = in.MetadataRootAccess.Value
	}
	if in.Name != nil {
		out.Name = in.Name.Value
	}
	if in.Owner != nil {
		out.Owner = in.Owner.Value
	}
	if in.RuntimeId != nil {
		out.RuntimeId = in.RuntimeId.Value
	}
	if in.Status != nil {
		out.Status = in.Status.Value
	}
	if in.StatusTime != nil {
		date := strfmt.DateTime(time.Unix(in.StatusTime.Seconds, 0))
		out.StatusTime = &date
	}
	if in.SubnetId != nil {
		out.SubnetId = in.SubnetId.Value
	}
	if in.TransitionStatus != nil {
		out.TransitionStatus = in.TransitionStatus.Value
	}
	if in.UpgradeStatus != nil {
		out.UpgradeStatus = in.UpgradeStatus.Value
	}
	if in.UpgradeTime != nil {
		date := strfmt.DateTime(time.Unix(in.UpgradeTime.Seconds, 0))
		out.UpgradeTime = &date
	}
	if in.VersionId != nil {
		out.VersionId = in.VersionId.Value
	}
	if in.VpcId != nil {
		out.VpcId = in.VpcId.Value
	}
	if in.Zone != nil {
		out.Zone = in.Zone.Value
	}
	return &out
}
