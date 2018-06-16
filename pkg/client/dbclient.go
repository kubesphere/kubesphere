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

package client

import (
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/options"
)

var dbClient *gorm.DB

const database = "kubesphere"

func NewDBClient() *gorm.DB {

	if dbClient != nil {
		return dbClient
	}

	user := options.ServerOptions.GetMysqlUser()
	passwd := options.ServerOptions.GetMysqlPassword()
	addr := options.ServerOptions.GetMysqlAddr()
	conn := fmt.Sprintf("%s:%s@tcp(%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local", user, passwd, addr)
	db, err := gorm.Open("mysql", conn)

	if err != nil {
		glog.Error(err)
		panic(err)
	}

	db.Exec(fmt.Sprintf("create database  if not exists %s;", database))

	conn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, passwd, addr, database)
	db, err = gorm.Open("mysql", conn)

	if err != nil {
		glog.Error(err)
		panic(err)
	}
	dbClient = db
	return dbClient
}
