// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	v1 "github.com/open-policy-agent/opa/v1/ast"
)

// Builtins is the registry of built-in functions supported by OPA.
// Call RegisterBuiltin to add a new built-in.
var Builtins = v1.Builtins

// RegisterBuiltin adds a new built-in function to the registry.
func RegisterBuiltin(b *Builtin) {
	v1.RegisterBuiltin(b)
}

// DefaultBuiltins is the registry of built-in functions supported in OPA
// by default. When adding a new built-in function to OPA, update this
// list.
var DefaultBuiltins = v1.DefaultBuiltins

// BuiltinMap provides a convenient mapping of built-in names to
// built-in definitions.
var BuiltinMap = v1.BuiltinMap

// Deprecated: Builtins can now be directly annotated with the
// Nondeterministic property, and when set to true, will be ignored
// for partial evaluation.
var IgnoreDuringPartialEval = v1.IgnoreDuringPartialEval

/**
 * Unification
 */

// Equality represents the "=" operator.
var Equality = v1.Equality

/**
 * Assignment
 */

// Assign represents the assignment (":=") operator.
var Assign = v1.Assign

// Member represents the `in` (infix) operator.
var Member = v1.Member

// MemberWithKey represents the `in` (infix) operator when used
// with two terms on the lhs, i.e., `k, v in obj`.
var MemberWithKey = v1.MemberWithKey

var GreaterThan = v1.GreaterThan

var GreaterThanEq = v1.GreaterThanEq

// LessThan represents the "<" comparison operator.
var LessThan = v1.LessThan

var LessThanEq = v1.LessThanEq

var NotEqual = v1.NotEqual

// Equal represents the "==" comparison operator.
var Equal = v1.Equal

var Plus = v1.Plus

var Minus = v1.Minus

var Multiply = v1.Multiply

var Divide = v1.Divide

var Round = v1.Round

var Ceil = v1.Ceil

var Floor = v1.Floor

var Abs = v1.Abs

var Rem = v1.Rem

/**
 * Bitwise
 */

var BitsOr = v1.BitsOr

var BitsAnd = v1.BitsAnd

var BitsNegate = v1.BitsNegate

var BitsXOr = v1.BitsXOr

var BitsShiftLeft = v1.BitsShiftLeft

var BitsShiftRight = v1.BitsShiftRight

/**
 * Sets
 */

var And = v1.And

// Or performs a union operation on sets.
var Or = v1.Or

var Intersection = v1.Intersection

var Union = v1.Union

/**
 * Aggregates
 */

var Count = v1.Count

var Sum = v1.Sum

var Product = v1.Product

var Max = v1.Max

var Min = v1.Min

/**
 * Sorting
 */

var Sort = v1.Sort

/**
 * Arrays
 */

var ArrayConcat = v1.ArrayConcat

var ArraySlice = v1.ArraySlice

var ArrayReverse = v1.ArrayReverse

/**
 * Conversions
 */

var ToNumber = v1.ToNumber

/**
 * Regular Expressions
 */

var RegexMatch = v1.RegexMatch

var RegexIsValid = v1.RegexIsValid

var RegexFindAllStringSubmatch = v1.RegexFindAllStringSubmatch

var RegexTemplateMatch = v1.RegexTemplateMatch

var RegexSplit = v1.RegexSplit

// RegexFind takes two strings and a number, the pattern, the value and number of match values to
// return, -1 means all match values.
var RegexFind = v1.RegexFind

// GlobsMatch takes two strings regexp-style strings and evaluates to true if their
// intersection matches a non-empty set of non-empty strings.
// Examples:
//   - "a.a." and ".b.b" -> true.
//   - "[a-z]*" and [0-9]+" -> not true.
var GlobsMatch = v1.GlobsMatch

/**
 * Strings
 */

var AnyPrefixMatch = v1.AnyPrefixMatch

var AnySuffixMatch = v1.AnySuffixMatch

var Concat = v1.Concat

var FormatInt = v1.FormatInt

var IndexOf = v1.IndexOf

var IndexOfN = v1.IndexOfN

var Substring = v1.Substring

var Contains = v1.Contains

var StringCount = v1.StringCount

var StartsWith = v1.StartsWith

var EndsWith = v1.EndsWith

var Lower = v1.Lower

var Upper = v1.Upper

var Split = v1.Split

var Replace = v1.Replace

var ReplaceN = v1.ReplaceN

var RegexReplace = v1.RegexReplace

var Trim = v1.Trim

var TrimLeft = v1.TrimLeft

var TrimPrefix = v1.TrimPrefix

var TrimRight = v1.TrimRight

var TrimSuffix = v1.TrimSuffix

var TrimSpace = v1.TrimSpace

var Sprintf = v1.Sprintf

var StringReverse = v1.StringReverse

var RenderTemplate = v1.RenderTemplate

/**
 * Numbers
 */

// RandIntn returns a random number 0 - n
// Marked non-deterministic because it relies on RNG internally.
var RandIntn = v1.RandIntn

var NumbersRange = v1.NumbersRange

var NumbersRangeStep = v1.NumbersRangeStep

/**
 * Units
 */

var UnitsParse = v1.UnitsParse

var UnitsParseBytes = v1.UnitsParseBytes

//
/**
 * Type
 */

// UUIDRFC4122 returns a version 4 UUID string.
// Marked non-deterministic because it relies on RNG internally.
var UUIDRFC4122 = v1.UUIDRFC4122

var UUIDParse = v1.UUIDParse

/**
 * JSON
 */

var JSONFilter = v1.JSONFilter

var JSONRemove = v1.JSONRemove

var JSONPatch = v1.JSONPatch

var ObjectSubset = v1.ObjectSubset

var ObjectUnion = v1.ObjectUnion

var ObjectUnionN = v1.ObjectUnionN

var ObjectRemove = v1.ObjectRemove

var ObjectFilter = v1.ObjectFilter

var ObjectGet = v1.ObjectGet

var ObjectKeys = v1.ObjectKeys

/*
 *  Encoding
 */

var JSONMarshal = v1.JSONMarshal

var JSONMarshalWithOptions = v1.JSONMarshalWithOptions

var JSONUnmarshal = v1.JSONUnmarshal

var JSONIsValid = v1.JSONIsValid

var Base64Encode = v1.Base64Encode

var Base64Decode = v1.Base64Decode

var Base64IsValid = v1.Base64IsValid

var Base64UrlEncode = v1.Base64UrlEncode

var Base64UrlEncodeNoPad = v1.Base64UrlEncodeNoPad

var Base64UrlDecode = v1.Base64UrlDecode

var URLQueryDecode = v1.URLQueryDecode

var URLQueryEncode = v1.URLQueryEncode

var URLQueryEncodeObject = v1.URLQueryEncodeObject

var URLQueryDecodeObject = v1.URLQueryDecodeObject

var YAMLMarshal = v1.YAMLMarshal

var YAMLUnmarshal = v1.YAMLUnmarshal

// YAMLIsValid verifies the input string is a valid YAML document.
var YAMLIsValid = v1.YAMLIsValid

var HexEncode = v1.HexEncode

var HexDecode = v1.HexDecode

/**
 * Tokens
 */

var JWTDecode = v1.JWTDecode

var JWTVerifyRS256 = v1.JWTVerifyRS256

var JWTVerifyRS384 = v1.JWTVerifyRS384

var JWTVerifyRS512 = v1.JWTVerifyRS512

var JWTVerifyPS256 = v1.JWTVerifyPS256

var JWTVerifyPS384 = v1.JWTVerifyPS384

var JWTVerifyPS512 = v1.JWTVerifyPS512

var JWTVerifyES256 = v1.JWTVerifyES256

var JWTVerifyES384 = v1.JWTVerifyES384

var JWTVerifyES512 = v1.JWTVerifyES512

var JWTVerifyHS256 = v1.JWTVerifyHS256

var JWTVerifyHS384 = v1.JWTVerifyHS384

var JWTVerifyHS512 = v1.JWTVerifyHS512

// Marked non-deterministic because it relies on time internally.
var JWTDecodeVerify = v1.JWTDecodeVerify

// Marked non-deterministic because it relies on RNG internally.
var JWTEncodeSignRaw = v1.JWTEncodeSignRaw

// Marked non-deterministic because it relies on RNG internally.
var JWTEncodeSign = v1.JWTEncodeSign

/**
 * Time
 */

// Marked non-deterministic because it relies on time directly.
var NowNanos = v1.NowNanos

var ParseNanos = v1.ParseNanos

var ParseRFC3339Nanos = v1.ParseRFC3339Nanos

var ParseDurationNanos = v1.ParseDurationNanos

var Format = v1.Format

var Date = v1.Date

var Clock = v1.Clock

var Weekday = v1.Weekday

var AddDate = v1.AddDate

var Diff = v1.Diff

/**
 * Crypto.
 */

var CryptoX509ParseCertificates = v1.CryptoX509ParseCertificates

var CryptoX509ParseAndVerifyCertificates = v1.CryptoX509ParseAndVerifyCertificates

var CryptoX509ParseAndVerifyCertificatesWithOptions = v1.CryptoX509ParseAndVerifyCertificatesWithOptions

var CryptoX509ParseCertificateRequest = v1.CryptoX509ParseCertificateRequest

var CryptoX509ParseKeyPair = v1.CryptoX509ParseKeyPair
var CryptoX509ParseRSAPrivateKey = v1.CryptoX509ParseRSAPrivateKey

var CryptoParsePrivateKeys = v1.CryptoParsePrivateKeys

var CryptoMd5 = v1.CryptoMd5

var CryptoSha1 = v1.CryptoSha1

var CryptoSha256 = v1.CryptoSha256

var CryptoHmacMd5 = v1.CryptoHmacMd5

var CryptoHmacSha1 = v1.CryptoHmacSha1

var CryptoHmacSha256 = v1.CryptoHmacSha256

var CryptoHmacSha512 = v1.CryptoHmacSha512

var CryptoHmacEqual = v1.CryptoHmacEqual

/**
 * Graphs.
 */

var WalkBuiltin = v1.WalkBuiltin

var ReachableBuiltin = v1.ReachableBuiltin

var ReachablePathsBuiltin = v1.ReachablePathsBuiltin

/**
 * Type
 */

var IsNumber = v1.IsNumber

var IsString = v1.IsString

var IsBoolean = v1.IsBoolean

var IsArray = v1.IsArray

var IsSet = v1.IsSet

var IsObject = v1.IsObject

var IsNull = v1.IsNull

/**
 * Type Name
 */

// TypeNameBuiltin returns the type of the input.
var TypeNameBuiltin = v1.TypeNameBuiltin

/**
 * HTTP Request
 */

// Marked non-deterministic because HTTP request results can be non-deterministic.
var HTTPSend = v1.HTTPSend

/**
 * GraphQL
 */

// GraphQLParse returns a pair of AST objects from parsing/validation.
var GraphQLParse = v1.GraphQLParse

// GraphQLParseAndVerify returns a boolean and a pair of AST object from parsing/validation.
var GraphQLParseAndVerify = v1.GraphQLParseAndVerify

// GraphQLParseQuery parses the input GraphQL query and returns a JSON
// representation of its AST.
var GraphQLParseQuery = v1.GraphQLParseQuery

// GraphQLParseSchema parses the input GraphQL schema and returns a JSON
// representation of its AST.
var GraphQLParseSchema = v1.GraphQLParseSchema

// GraphQLIsValid returns true if a GraphQL query is valid with a given
// schema, and returns false for all other inputs.
var GraphQLIsValid = v1.GraphQLIsValid

// GraphQLSchemaIsValid returns true if the input is valid GraphQL schema,
// and returns false for all other inputs.
var GraphQLSchemaIsValid = v1.GraphQLSchemaIsValid

/**
 * JSON Schema
 */

// JSONSchemaVerify returns empty string if the input is valid JSON schema
// and returns error string for all other inputs.
var JSONSchemaVerify = v1.JSONSchemaVerify

// JSONMatchSchema returns empty array if the document matches the JSON schema,
// and returns non-empty array with error objects otherwise.
var JSONMatchSchema = v1.JSONMatchSchema

/**
 * Cloud Provider Helper Functions
 */

var ProvidersAWSSignReqObj = v1.ProvidersAWSSignReqObj

/**
 * Rego
 */

var RegoParseModule = v1.RegoParseModule

var RegoMetadataChain = v1.RegoMetadataChain

// RegoMetadataRule returns the metadata for the active rule
var RegoMetadataRule = v1.RegoMetadataRule

/**
 * OPA
 */

// Marked non-deterministic because of unpredictable config/environment-dependent results.
var OPARuntime = v1.OPARuntime

/**
 * Trace
 */

var Trace = v1.Trace

/**
 * Glob
 */

var GlobMatch = v1.GlobMatch

var GlobQuoteMeta = v1.GlobQuoteMeta

/**
 * Networking
 */

var NetCIDRIntersects = v1.NetCIDRIntersects

var NetCIDRExpand = v1.NetCIDRExpand

var NetCIDRContains = v1.NetCIDRContains

var NetCIDRContainsMatches = v1.NetCIDRContainsMatches

var NetCIDRMerge = v1.NetCIDRMerge

var NetCIDRIsValid = v1.NetCIDRIsValid

// Marked non-deterministic because DNS resolution results can be non-deterministic.
var NetLookupIPAddr = v1.NetLookupIPAddr

/**
 * Semantic Versions
 */

var SemVerIsValid = v1.SemVerIsValid

var SemVerCompare = v1.SemVerCompare

/**
 * Printing
 */

// Print is a special built-in function that writes zero or more operands
// to a message buffer. The caller controls how the buffer is displayed. The
// operands may be of any type. Furthermore, unlike other built-in functions,
// undefined operands DO NOT cause the print() function to fail during
// evaluation.
var Print = v1.Print

// InternalPrint represents the internal implementation of the print() function.
// The compiler rewrites print() calls to refer to the internal implementation.
var InternalPrint = v1.InternalPrint

/**
 * Deprecated built-ins.
 */

// SetDiff has been replaced by the minus built-in.
var SetDiff = v1.SetDiff

// NetCIDROverlap has been replaced by the `net.cidr_contains` built-in.
var NetCIDROverlap = v1.NetCIDROverlap

// CastArray checks the underlying type of the input. If it is array or set, an array
// containing the values is returned. If it is not an array, an error is thrown.
var CastArray = v1.CastArray

// CastSet checks the underlying type of the input.
// If it is a set, the set is returned.
// If it is an array, the array is returned in set form (all duplicates removed)
// If neither, an error is thrown
var CastSet = v1.CastSet

// CastString returns input if it is a string; if not returns error.
// For formatting variables, see sprintf
var CastString = v1.CastString

// CastBoolean returns input if it is a boolean; if not returns error.
var CastBoolean = v1.CastBoolean

// CastNull returns null if input is null; if not returns error.
var CastNull = v1.CastNull

// CastObject returns the given object if it is null; throws an error otherwise
var CastObject = v1.CastObject

// RegexMatchDeprecated declares `re_match` which has been deprecated. Use `regex.match` instead.
var RegexMatchDeprecated = v1.RegexMatchDeprecated

// All takes a list and returns true if all of the items
// are true. A collection of length 0 returns true.
var All = v1.All

// Any takes a collection and returns true if any of the items
// is true. A collection of length 0 returns false.
var Any = v1.Any

// Builtin represents a built-in function supported by OPA. Every built-in
// function is uniquely identified by a name.
type Builtin = v1.Builtin
