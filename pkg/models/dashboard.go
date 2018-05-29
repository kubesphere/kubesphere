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

package models

import (
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/client"
	"fmt"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	influxClient "github.com/influxdata/influxdb/client/v2"
	"time"
	"strings"

	"strconv"
	"encoding/json"
)


type PodStatistics struct {
	Total     string       `json:"total"`
	Running   string       `json:"running"`
	Timestamp time.Time `json:"timestamp"`
}

const STATUS = "Running"
const COMPLETED  = "Completed"
const PODSTATISTICS = "pods_count"
const DAY  = "d"

// Get List All Namespaces Pods
func GetAllPods() (pods PodStatistics, err error) {

	namespaces := GetNameSpaces()

	k8sclient := client.NewK8sClient()



	var total int
	var running int
	if len(namespaces) > 0 {

		for _, namespace := range namespaces {

			nspods, err := k8sclient.CoreV1().Pods(namespace).List(meta_v1.ListOptions{})

			if err != nil {

				glog.Fatal(err)

				return pods, err
			}
			// sum all namespace pods

			if len(nspods.Items) > 0 {

				for _, item := range nspods.Items {


					if item.Status.Phase != COMPLETED{

						total++

						if item.Status.Phase == STATUS {

							running++
						}
					}

				}

			}

		}

	}

	pods.Total = strconv.Itoa(total)
	pods.Running = strconv.Itoa(running)
	pods.Timestamp = time.Now()

	return pods, nil

}

//Send Statistic data to influxdb

func StorePodsStatis() {

	host := fmt.Sprintf("http://%s:%d", "monitoring-influxdb", 8086)

	// Create a new HTTPClient
	client, err := influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr: host,
	})
	if err != nil {
		glog.Fatal(err)
	}
	defer client.Close()

	// Create a new point batch
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  "k8s",
		Precision: "s",
	})
	if err != nil {
		glog.Fatal(err)
	}

	pods, err := GetAllPods()

	if err != nil {
		glog.Fatal(err)
	}

	dateTime := time.Now().Format("2006-01-02 15:04:05")

	date := strings.Split(dateTime, " ")

	tags := map[string]string{"date": date[0]}

	total,_ := strconv.Atoi(pods.Total)
	running,_ := strconv.Atoi(pods.Running)

	values := map[string]interface{}{

		"total":    total,
		"running":  running,
	}

	pt, err := influxClient.NewPoint(PODSTATISTICS, tags, values, time.Now())

	if err != nil {
		glog.Fatal(err)

	}

	bp.AddPoint(pt)

	// Write the batch
	if err := client.Write(bp); err != nil {
		glog.Fatal(err)
	}

	glog.Infoln("Success store one point:", pt)

	// Close client resources
	if err := client.Close(); err != nil {
		glog.Fatal(err)
	}

}

// query statistic data from influx
/***
 querytype:d h

 */

func QueryPodsData(querytype string) (dataList []PodStatistics,err error){


	host := fmt.Sprintf("http://%s:%d", "monitoring-influxdb", 8086)

	// Create a new HTTPClient
	client, err := influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr: host,
	})
	if err != nil {
		glog.Fatal(err)
	}
	defer client.Close()

	glog.Infoln("success make connect with influx")

	var cmd string

	if querytype == DAY {

		cmd = fmt.Sprintf("SELECT mean(\"total\") AS \"total\" , mean(\"running\") as \"running\" " +
			"FROM pods_count WHERE time > now() - 7d  GROUP BY time(1d) FILL(previous)")

	}else {

		cmd = fmt.Sprintf("SELECT mean(\"total\") as \"total\", mean(\"running\") as running " +
			"FROM pods_count WHERE time > now() - 7h  GROUP BY time(1h) FILL(previous)")
	}

	query := influxClient.NewQuery(cmd,"k8s","s")

	if data, err := client.Query(query); err == nil && data.Error() == nil {

		if len(data.Results[0].Series) > 0{


			for _,row := range data.Results[0].Series{


				for _,cols := range row.Values{

					var pod PodStatistics

					for index, col := range cols{

						colName := row.Columns[index]

						switch colName {

						case "total":
							if col != nil{

								pod.Total = col.(json.Number).String()

							}else {

								pod.Total = "0"

							}

							break

						case "running":

							if col != nil{
								pod.Running = col.(json.Number).String()

							}else {

								pod.Running = "0"

							}

							break

						default:

							tmp,_:= strconv.ParseInt(col.(json.Number).String(),10,64)
							timestamp := time.Unix(tmp,0)
							pod.Timestamp = timestamp

							break

						}

					}

					dataList = append(dataList, pod)

				}

			}

		}



	} else {


		return dataList,err
	}

	return dataList[1:],nil

}

