// Copyright 2017 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package db

import (
	"context"
	"sync"
	"time"

	"github.com/gocraft/dbr"

	"openpitrix.io/openpitrix/pkg/config"
	"openpitrix.io/openpitrix/pkg/logger"
)

type key int

var dbMap = sync.Map{}
var dbKey key

type Database struct {
	Conn *dbr.Connection
}

func OpenDatabase(cfg config.MysqlConfig) (*Database, error) {
	// https://github.com/go-sql-driver/mysql/issues/9
	conn, err := dbr.Open("mysql", cfg.GetUrl()+"?parseTime=1&multiStatements=1&charset=utf8mb4&collation=utf8mb4_unicode_ci", nil)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(100)
	conn.SetMaxOpenConns(100)
	conn.SetConnMaxLifetime(10 * time.Second)

	db := &Database{
		Conn: conn,
	}
	dbMap.Store(cfg, db)
	return db, nil
}

func NewContext(ctx context.Context, cfg config.MysqlConfig) context.Context {
	return context.WithValue(ctx, dbKey, cfg)
}

func FromContext(ctx context.Context) (*Database, bool) {
	cfg := ctx.Value(dbKey).(config.MysqlConfig)
	var err error
	db, ok := dbMap.Load(cfg)
	if !ok {
		db, err = OpenDatabase(cfg)
		if err != nil {
			logger.Critical(ctx, "Failed to open database: %+v", err)
			return nil, false
		}
	}
	return db.(*Database), true
}

func (db *Database) New(ctx context.Context) *Conn {
	actualDb, ok := FromContext(ctx)
	var conn *dbr.Connection
	if ok || db == nil {
		conn = actualDb.Conn
	} else {
		conn = db.Conn
	}
	return &Conn{
		Session: conn.NewSession(&EventReceiver{ctx}),
		ctx:     ctx,
	}
}
