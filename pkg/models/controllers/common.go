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

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	checkPeriod = 30 * time.Minute
	sleepPeriod = 15 * time.Second
)

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

func makeHttpRequest(method, url, data string) ([]byte, error) {
	var req *http.Request

	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(data))
	}

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		err := fmt.Errorf("Request to %s failed, method: %s, reason: %s ", url, method, err)
		glog.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		err = errors.New(string(body))
	}
	return body, err
}

func handleCrash(ctl Controller) {
	close(ctl.chanAlive())
	if err := recover(); err != nil {
		glog.Errorf("panic occur in %s controller's listAndWatch function, reason: %s", ctl.Name(), err)
		return
	}
}

func hasSynced(ctl Controller) bool {
	totalInDb := ctl.CountWithConditions("")
	totalInK8s := ctl.total()

	if totalInDb == totalInK8s {
		return true
	}

	return false
}

func checkAndResync(ctl Controller, stopChan chan struct{}) {
	defer close(stopChan)

	lastTime := time.Now()

	for {
		select {
		case <-ctl.chanStop():
			return
		default:
			if time.Now().Sub(lastTime) < checkPeriod {
				time.Sleep(sleepPeriod)
				break
			}

			lastTime = time.Now()
			if !hasSynced(ctl) {
				glog.Errorf("the data in db and kubernetes is inconsistent, resync %s controller", ctl.Name())
				close(stopChan)
				stopChan = make(chan struct{})
				go ctl.sync(stopChan)
			}
		}
	}
}

func listAndWatch(ctl Controller, wg *sync.WaitGroup) {
	defer handleCrash(ctl)
	defer ctl.CloseDB()
	defer wg.Done()
	stopChan := make(chan struct{})

	go ctl.sync(stopChan)

	checkAndResync(ctl, stopChan)
}
