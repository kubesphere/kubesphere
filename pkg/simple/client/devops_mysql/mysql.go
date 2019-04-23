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

package devops_mysql

import (
	"flag"
	"github.com/gocraft/dbr"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/db"
	"sync"
	"time"
)

var (
	dbClientOnce sync.Once
	dsn          string
	dbClient     *db.Database
)

func init() {
	flag.StringVar(&dsn, "devops-database-connection", "root@tcp(127.0.0.1:3306)/devops", "data source name")
}

var defaultEventReceiver = db.EventReceiver{}

func OpenDatabase() *db.Database {
	dbClientOnce.Do(func() {
		conn, err := dbr.Open("mysql", dsn+"?parseTime=1&multiStatements=1&charset=utf8mb4&collation=utf8mb4_unicode_ci", &defaultEventReceiver)
		if err != nil {
			glog.Fatal(err)
		}
		conn.SetMaxIdleConns(100)
		conn.SetMaxOpenConns(100)
		conn.SetConnMaxLifetime(10 * time.Second)
		dbClient = &db.Database{
			Session: conn.NewSession(nil),
		}
		err = dbClient.Ping()
		if err != nil {
			glog.Error(err)
		}
	})
	return dbClient
}
