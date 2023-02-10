// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	ghodss "github.com/ghodss/yaml"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/util"
)

func builtinJSONMarshal(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	asJSON, err := ast.JSON(operands[0].Value)
	if err != nil {
		return err
	}

	bs, err := json.Marshal(asJSON)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(string(bs)))
}

func builtinJSONUnmarshal(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	var x interface{}

	if err := util.UnmarshalJSON([]byte(str), &x); err != nil {
		return err
	}
	v, err := ast.InterfaceToValue(x)
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(v))
}

func builtinJSONIsValid(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return iter(ast.BooleanTerm(false))
	}

	return iter(ast.BooleanTerm(json.Valid([]byte(str))))
}

func builtinBase64Encode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(base64.StdEncoding.EncodeToString([]byte(str))))
}

func builtinBase64Decode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	result, err := base64.StdEncoding.DecodeString(string(str))
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(ast.String(result)))
}

func builtinBase64IsValid(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return iter(ast.BooleanTerm(false))
	}

	_, err = base64.StdEncoding.DecodeString(string(str))
	return iter(ast.BooleanTerm(err == nil))
}

func builtinBase64UrlEncode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(base64.URLEncoding.EncodeToString([]byte(str))))
}

func builtinBase64UrlEncodeNoPad(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	return iter(ast.StringTerm(base64.RawURLEncoding.EncodeToString([]byte(str))))
}

func builtinBase64UrlDecode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s := string(str)

	// Some base64url encoders omit the padding at the end, so this case
	// corrects such representations using the method given in RFC 7515
	// Appendix C: https://tools.ietf.org/html/rfc7515#appendix-C
	if !strings.HasSuffix(s, "=") {
		switch len(s) % 4 {
		case 0:
		case 2:
			s += "=="
		case 3:
			s += "="
		default:
			return fmt.Errorf("illegal base64url string: %s", s)
		}
	}
	result, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(ast.String(result)))
}

func builtinURLQueryEncode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	return iter(ast.StringTerm(url.QueryEscape(string(str))))
}

func builtinURLQueryDecode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s, err := url.QueryUnescape(string(str))
	if err != nil {
		return err
	}
	return iter(ast.StringTerm(s))
}

var encodeObjectErr = builtins.NewOperandErr(1, "values must be string, array[string], or set[string]")

func builtinURLQueryEncodeObject(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	asJSON, err := ast.JSON(operands[0].Value)
	if err != nil {
		return err
	}

	inputs, ok := asJSON.(map[string]interface{})
	if !ok {
		return builtins.NewOperandTypeErr(1, operands[0].Value, "object")
	}

	query := url.Values{}

	for k, v := range inputs {
		switch vv := v.(type) {
		case string:
			query.Set(k, vv)
		case []interface{}:
			for _, val := range vv {
				strVal, ok := val.(string)
				if !ok {
					return encodeObjectErr
				}
				query.Add(k, strVal)
			}
		default:
			return encodeObjectErr
		}
	}

	return iter(ast.StringTerm(query.Encode()))
}

func builtinURLQueryDecodeObject(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	query, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	queryParams, err := url.ParseQuery(string(query))
	if err != nil {
		return err
	}

	queryObject := ast.NewObject()
	for k, v := range queryParams {
		paramsArray := make([]*ast.Term, len(v))
		for i, param := range v {
			paramsArray[i] = ast.StringTerm(param)
		}
		queryObject.Insert(ast.StringTerm(k), ast.ArrayTerm(paramsArray...))
	}

	return iter(ast.NewTerm(queryObject))
}

func builtinYAMLMarshal(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	asJSON, err := ast.JSON(operands[0].Value)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(asJSON); err != nil {
		return err
	}

	bs, err := ghodss.JSONToYAML(buf.Bytes())
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(string(bs)))
}

func builtinYAMLUnmarshal(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	bs, err := ghodss.YAMLToJSON([]byte(str))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(bs)
	decoder := util.NewJSONDecoder(buf)
	var val interface{}
	err = decoder.Decode(&val)
	if err != nil {
		return err
	}
	v, err := ast.InterfaceToValue(val)
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(v))
}

func builtinYAMLIsValid(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return iter(ast.BooleanTerm(false))
	}

	var x interface{}
	err = ghodss.Unmarshal([]byte(str), &x)
	return iter(ast.BooleanTerm(err == nil))
}

func builtinHexEncode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	return iter(ast.StringTerm(hex.EncodeToString([]byte(str))))
}

func builtinHexDecode(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	val, err := hex.DecodeString(string(str))
	if err != nil {
		return err
	}
	return iter(ast.NewTerm(ast.String(val)))
}

func init() {
	RegisterBuiltinFunc(ast.JSONMarshal.Name, builtinJSONMarshal)
	RegisterBuiltinFunc(ast.JSONUnmarshal.Name, builtinJSONUnmarshal)
	RegisterBuiltinFunc(ast.JSONIsValid.Name, builtinJSONIsValid)
	RegisterBuiltinFunc(ast.Base64Encode.Name, builtinBase64Encode)
	RegisterBuiltinFunc(ast.Base64Decode.Name, builtinBase64Decode)
	RegisterBuiltinFunc(ast.Base64IsValid.Name, builtinBase64IsValid)
	RegisterBuiltinFunc(ast.Base64UrlEncode.Name, builtinBase64UrlEncode)
	RegisterBuiltinFunc(ast.Base64UrlEncodeNoPad.Name, builtinBase64UrlEncodeNoPad)
	RegisterBuiltinFunc(ast.Base64UrlDecode.Name, builtinBase64UrlDecode)
	RegisterBuiltinFunc(ast.URLQueryDecode.Name, builtinURLQueryDecode)
	RegisterBuiltinFunc(ast.URLQueryEncode.Name, builtinURLQueryEncode)
	RegisterBuiltinFunc(ast.URLQueryEncodeObject.Name, builtinURLQueryEncodeObject)
	RegisterBuiltinFunc(ast.URLQueryDecodeObject.Name, builtinURLQueryDecodeObject)
	RegisterBuiltinFunc(ast.YAMLMarshal.Name, builtinYAMLMarshal)
	RegisterBuiltinFunc(ast.YAMLUnmarshal.Name, builtinYAMLUnmarshal)
	RegisterBuiltinFunc(ast.YAMLIsValid.Name, builtinYAMLIsValid)
	RegisterBuiltinFunc(ast.HexEncode.Name, builtinHexEncode)
	RegisterBuiltinFunc(ast.HexDecode.Name, builtinHexDecode)
}
