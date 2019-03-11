/*

 Copyright 2019 The KubeSphere Authors.

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

package client

import (
	"flag"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	dbClientOnce sync.Once
	dbClient     *gorm.DB
	dsn          string
)

func init() {
	flag.StringVar(&dsn, "database-connection", "root@tcp(localhost:3306)/kubesphere?charset=utf8&parseTime=True", "data source name")
}

func DBClient() *gorm.DB {
	dbClientOnce.Do(func() {
		var err error
		dbClient, err = gorm.Open("mysql", dsn)

		if err != nil {
			log.Fatalln(err)
		}

		if err := dbClient.DB().Ping(); err != nil {
			log.Fatalln(err)
		}
	})

	return dbClient

}
