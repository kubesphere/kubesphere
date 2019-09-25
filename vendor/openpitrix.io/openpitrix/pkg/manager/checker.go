// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package manager

import (
	"context"

	"github.com/fatih/structs"
	"github.com/golang/protobuf/ptypes/wrappers"

	"openpitrix.io/openpitrix/pkg/gerr"
	"openpitrix.io/openpitrix/pkg/util/stringutil"
)

type checker struct {
	ctx          context.Context
	req          Request
	required     []string
	stringChosen map[string][]string
}

func NewChecker(ctx context.Context, req Request) *checker {
	return &checker{
		ctx:          ctx,
		req:          req,
		required:     []string{},
		stringChosen: make(map[string][]string),
	}
}

func (c *checker) Required(params ...string) *checker {
	c.required = append(c.required, params...)
	return c
}

func (c *checker) checkRequired(param string, value interface{}) error {
	if len(c.required) > 0 && stringutil.StringIn(param, c.required) {
		switch v := value.(type) {
		case string:
			if v == "" {
				return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorMissingParameter, param)
			}
		case *wrappers.StringValue:
			if v == nil || v.GetValue() == "" {
				return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorMissingParameter, param)
			}
		case *wrappers.BytesValue:
			if v == nil || len(v.GetValue()) == 0 {
				return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorMissingParameter, param)
			}
		case []byte:
			if len(v) == 0 {
				return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorMissingParameter, param)
			}
		case []string:
			var values []string
			for _, v := range v {
				if v != "" {
					values = append(values, v)
				}
			}
			if len(values) == 0 {
				return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorMissingParameter, param)
			}
		}
	}
	return nil
}

func (c *checker) StringChosen(param string, chosen []string) *checker {
	if exist, ok := c.stringChosen[param]; ok {
		c.stringChosen[param] = append(exist, chosen...)
	} else {
		c.stringChosen[param] = chosen
	}
	return c
}

func (c *checker) checkStringChosen(param string, value interface{}) error {
	if len(c.stringChosen) > 0 {
		if chosen, ok := c.stringChosen[param]; ok {
			switch v := value.(type) {
			case string:
				if !stringutil.StringIn(v, chosen) {
					return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorUnsupportedParameterValue, param, v)
				}
			case *wrappers.StringValue:
				if v != nil {
					if !stringutil.StringIn(v.GetValue(), chosen) {
						return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorUnsupportedParameterValue, param, v.GetValue())
					}
				}
			case []string:
				for _, s := range v {
					if !stringutil.StringIn(s, chosen) {
						return gerr.New(c.ctx, gerr.InvalidArgument, gerr.ErrorUnsupportedParameterValue, param, s)
					}
				}
			}
		}
	}
	return nil
}

func (c *checker) chainChecker(param string, value interface{}, checks ...func(string, interface{}) error) error {
	var err error
	for _, c := range checks {
		err = c(param, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *checker) Exec() error {
	for _, field := range structs.Fields(c.req) {
		param := getFieldName(field)
		value := field.Value()

		err := c.chainChecker(param, value,
			c.checkRequired,
			c.checkStringChosen,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
