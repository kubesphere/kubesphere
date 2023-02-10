// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/providers/aws"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var awsRequiredConfigKeyNames = ast.NewSet(
	ast.StringTerm("aws_service"),
	ast.StringTerm("aws_access_key"),
	ast.StringTerm("aws_secret_access_key"),
	ast.StringTerm("aws_region"),
)

func stringFromTerm(t *ast.Term) string {
	if v, ok := t.Value.(ast.String); ok {
		return string(v)
	}
	return ""
}

func getReqBodyBytes(body, rawBody *ast.Term) ([]byte, error) {
	var out []byte

	switch {
	case rawBody != nil:
		out = []byte(stringFromTerm(rawBody))
	case body != nil:
		bodyVal := body.Value
		bodyValInterface, err := ast.JSON(bodyVal)
		if err != nil {
			return nil, err
		}
		bodyValBytes, err := json.Marshal(bodyValInterface)
		if err != nil {
			return nil, err
		}
		out = bodyValBytes
	default:
		out = []byte("")
	}

	return out, nil
}

func objectToMap(o ast.Object) map[string][]string {
	var out map[string][]string
	o.Foreach(func(k, v *ast.Term) {
		ks := stringFromTerm(k)
		vs := stringFromTerm(v)
		out[ks] = []string{vs}
	})
	return out
}

// Note(philipc): This is roughly the same approach used for http.send.
func validateAWSAuthParameters(o ast.Object) error {
	awsKeys := ast.NewSet(o.Keys()...)

	missingKeys := awsRequiredConfigKeyNames.Diff(awsKeys)
	if missingKeys.Len() != 0 {
		return builtins.NewOperandErr(2, "missing required AWS config parameters(s): %v", missingKeys)
	}

	invalidKeys := ast.NewSet()
	awsRequiredConfigKeyNames.Foreach(func(t *ast.Term) {
		if v := o.Get(t); v != nil {
			if _, ok := v.Value.(ast.String); !ok {
				invalidKeys.Add(t)
			}
		}
	})
	if invalidKeys.Len() != 0 {
		return builtins.NewOperandErr(2, "invalid values for required AWS config parameters(s): %v", invalidKeys)
	}

	return nil
}

func builtinAWSSigV4SignReq(ctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Request object.
	reqObj, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// AWS SigV4 config info object.
	awsConfigObj, err := builtins.ObjectOperand(operands[1].Value, 1)
	if err != nil {
		return err
	}
	// Make sure our required keys exist!
	err = validateAWSAuthParameters(awsConfigObj)
	if err != nil {
		return err
	}
	service := stringFromTerm(awsConfigObj.Get(ast.StringTerm("aws_service")))
	awsCreds := aws.CredentialsFromObject(awsConfigObj)

	// Timestamp for signing.
	var signingTimestamp time.Time
	timestamp, err := builtins.NumberOperand(operands[2].Value, 1)
	if err != nil {
		return err
	}

	ts, ok := timestamp.Int64()
	if !ok {
		return builtins.NewOperandErr(3, "could not convert time_ns value into a unix timestamp")
	}

	signingTimestamp = time.Unix(0, ts)
	if err != nil {
		return err
	}

	// Make sure our required keys exist!
	// This check is stricter than required, but better to break here than downstream.
	_, err = validateHTTPRequestOperand(operands[0], 1)
	if err != nil {
		return err
	}

	// Prepare required fields from the HTTP request object.
	var theURL *url.URL
	var method string
	reqURL := reqObj.Get(ast.StringTerm("url"))
	reqMethod := reqObj.Get(ast.StringTerm("method"))

	headers := ast.NewObject()
	headersTerm := reqObj.Get(ast.StringTerm("headers"))
	if headersTerm != nil {
		var ok bool
		headers, ok = headersTerm.Value.(ast.Object)
		if !ok {
			return builtins.NewOperandTypeErr(0, headersTerm.Value, "object")
		}
	}

	// Check types on the request parameters.
	invalidParameters := ast.NewSet()
	if _, ok := reqURL.Value.(ast.String); !ok {
		invalidParameters.Add(ast.StringTerm("url"))
	}
	if _, ok := reqMethod.Value.(ast.String); !ok {
		invalidParameters.Add(ast.StringTerm("method"))
	}
	if invalidParameters.Len() > 0 {
		return builtins.NewOperandErr(1, "invalid values for required request parameters(s): %v", invalidParameters)
	}

	theURL, err = url.Parse(stringFromTerm(reqURL))
	if err != nil {
		return err
	}
	method = stringFromTerm(reqMethod)

	bodyTerm := reqObj.Get(ast.StringTerm("body"))
	rawBodyTerm := reqObj.Get(ast.StringTerm("raw_body"))
	body, err := getReqBodyBytes(bodyTerm, rawBodyTerm)
	if err != nil {
		return err
	}

	// Sign the request object's headers, and reconstruct the headers map.
	authHeader, signedHeadersMap := aws.SignV4(objectToMap(headers), method, theURL, body, service, awsCreds, signingTimestamp)
	signedHeadersObj := ast.NewObject()
	signedHeadersObj.Insert(ast.StringTerm("Authorization"), ast.StringTerm(authHeader))
	for k, v := range signedHeadersMap {
		signedHeadersObj.Insert(ast.StringTerm(k), ast.StringTerm(v))
	}

	// Create new request object with updated headers.
	out := reqObj.Copy()
	out.Insert(ast.StringTerm("headers"), ast.NewTerm(signedHeadersObj))

	return iter(ast.NewTerm(out))
}

func init() {
	RegisterBuiltinFunc(ast.ProvidersAWSSignReqObj.Name, builtinAWSSigV4SignReq)
}
