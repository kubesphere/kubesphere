/*
Copyright 2018 The KubeSphere Authors.

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

package controllers

import "github.com/jinzhu/gorm"

func listWithConditions(db *gorm.DB, total *int, object, list interface{}, conditions string, paging *Paging, order string) {
	if len(conditions) == 0 {
		db.Model(object).Count(total)
	} else {
		db.Model(object).Where(conditions).Count(total)
	}

	if paging != nil {
		if len(conditions) > 0 {
			db.Where(conditions).Order(order).Limit(paging.Limit).Offset(paging.Offset).Find(list)
		} else {
			db.Order(order).Limit(paging.Limit).Offset(paging.Offset).Find(list)
		}

	} else {
		if len(conditions) > 0 {
			db.Where(conditions).Order(order).Find(list)
		} else {
			db.Order(order).Find(list)
		}
	}
}

func countWithConditions(db *gorm.DB, conditions string, object interface{}) int {
	var count int
	if len(conditions) == 0 {
		db.Model(object).Count(&count)
	} else {
		db.Model(object).Where(conditions).Count(&count)
	}
	return count
}
