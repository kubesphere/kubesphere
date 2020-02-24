/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package iam

import (
	"k8s.io/api/rbac/v1"
	"time"
)

const (
	ConfigPath      = "/etc/kubesphere/iam"
	KindTokenReview = "TokenReview"
)

type User struct {
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Lang        string    `json:"lang,omitempty"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"create_time"`
	Groups      []string  `json:"groups,omitempty"`
	Password    string    `json:"password,omitempty"`
}

type Action struct {
	Name  string          `json:"name"`
	Rules []v1.PolicyRule `json:"rules"`
}

type Rule struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

type SimpleRule struct {
	Name    string   `json:"name" description:"rule name"`
	Actions []string `json:"actions" description:"actions"`
}

type RoleList struct {
	ClusterRoles []*v1.ClusterRole `json:"clusterRole" description:"cluster role list"`
	Roles        []*v1.Role        `json:"roles" description:"role list"`
}
