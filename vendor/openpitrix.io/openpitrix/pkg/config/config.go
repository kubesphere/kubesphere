// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/koding/multiconfig"

	"openpitrix.io/openpitrix/pkg/logger"
)

type Config struct {
	Log         LogConfig
	Grpc        GrpcConfig
	Mysql       MysqlConfig
	Etcd        EtcdConfig
	IAM         IAMConfig
	Attachment  AttachmentConfig
	DisableGops bool `default:"false"`
}

type IAMConfig struct {
	SecretKey              string        `default:"OpenPitrix-lC4LipAXPYsuqw5F"`
	ExpireTime             time.Duration `default:"2h"`
	RefreshTokenExpireTime time.Duration `default:"336h"` // default is 2 week
}

type AttachmentConfig struct {
	AccessKey  string `default:"openpitrixminioaccesskey"`
	SecretKey  string `default:"openpitrixminiosecretkey"`
	Endpoint   string `default:"http://openpitrix-minio:9000"`
	BucketName string `default:"openpitrix-attachment"`
}

type LogConfig struct {
	Level string `default:"info"` // debug, info, warn, error, fatal
}

type GrpcConfig struct {
	ShowErrorCause bool `default:"false"` // show grpc error cause to frontend
}

type EtcdConfig struct {
	Endpoints string `default:"openpitrix-etcd:2379"` // Example: "localhost:2379,localhost:22379,localhost:32379"
}

type MysqlConfig struct {
	Host     string `default:"openpitrix-db"`
	Port     string `default:"3306"`
	User     string `default:"root"`
	Password string `default:"password"`
	Database string `default:"openpitrix"`
	Disable  bool   `default:"false"`
}

func (m *MysqlConfig) GetUrl() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", m.User, m.Password, m.Host, m.Port, m.Database)
}

func PrintUsage() {
	fmt.Fprintf(os.Stdout, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprint(os.Stdout, "\nSupported environment variables:\n")
	e := newLoader("openpitrix")
	e.PrintEnvs(new(Config))
	fmt.Println("")
}

func GetFlagSet() *flag.FlagSet {
	flag.CommandLine.Usage = PrintUsage
	return flag.CommandLine
}

func ParseFlag() {
	GetFlagSet().Parse(os.Args[1:])
}

var conf Config

func loadConf() *Config {
	config := new(Config)
	m := &multiconfig.DefaultLoader{}
	m.Loader = multiconfig.MultiLoader(newLoader("openpitrix"))
	m.Validator = multiconfig.MultiValidator(
		&multiconfig.RequiredValidator{},
	)
	err := m.Load(config)
	if err != nil {
		logger.Critical(nil, "Failed to load config: %+v", err)
		panic(err)
	}
	logger.SetLevelByString(config.Log.Level)
	logger.Debug(nil, "GetConf: %+v", config)

	return config
}

func init() {
	conf = *loadConf()
}

func GetConf() *Config {
	ParseFlag()
	var c = new(Config)
	*c = conf
	return c
}
