// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"encoding/json"
	"fmt"
	"strings"

	gqlast "github.com/open-policy-agent/opa/internal/gqlparser/ast"
	gqlparser "github.com/open-policy-agent/opa/internal/gqlparser/parser"
	gqlvalidator "github.com/open-policy-agent/opa/internal/gqlparser/validator"

	// Side-effecting import. Triggers GraphQL library's validation rule init() functions.
	_ "github.com/open-policy-agent/opa/internal/gqlparser/validator/rules"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
)

// Parses a GraphQL schema, and returns the GraphQL AST for the schema.
func parseSchema(schema string) (*gqlast.SchemaDocument, error) {
	// NOTE(philipc): We don't include the "built-in schema defs" from the
	// underlying graphql parsing library here, because those definitions
	// generate enormous AST blobs. In the future, if there is demand for
	// a "full-spec" version of schema ASTs, we may need to provide a
	// version of this function that includes the built-in schema
	// definitions.
	schemaAST, err := gqlparser.ParseSchema(&gqlast.Source{Input: schema})
	if err != nil {
		errorParts := strings.SplitN(err.Error(), ":", 4)
		msg := strings.TrimLeft(errorParts[3], " ")
		return nil, fmt.Errorf("%s in GraphQL string at location %s:%s", msg, errorParts[1], errorParts[2])
	}
	return schemaAST, nil
}

// Parses a GraphQL query, and returns the GraphQL AST for the query.
func parseQuery(query string) (*gqlast.QueryDocument, error) {
	queryAST, err := gqlparser.ParseQuery(&gqlast.Source{Input: query})
	if err != nil {
		errorParts := strings.SplitN(err.Error(), ":", 4)
		msg := strings.TrimLeft(errorParts[3], " ")
		return nil, fmt.Errorf("%s in GraphQL string at location %s:%s", msg, errorParts[1], errorParts[2])
	}
	return queryAST, nil
}

// Validates a GraphQL query against a schema, and returns an error.
// In this case, we get a wrappered error list type, and pluck out
// just the first error message in the list.
func validateQuery(schema *gqlast.Schema, query *gqlast.QueryDocument) error {
	// Validate the query against the schema, erroring if there's an issue.
	err := gqlvalidator.Validate(schema, query)
	if err != nil {
		// We use strings.TrimSuffix to remove the '.' characters that the library
		// authors include on most of their validation errors. This should be safe,
		// since variable names in their error messages are usually quoted, and
		// this affects only the last character(s) in the string.
		// NOTE(philipc): We know the error location will be in the query string,
		// because schema validation always happens before this function is called.
		errorParts := strings.SplitN(err.Error(), ":", 4)
		msg := strings.TrimSuffix(strings.TrimLeft(errorParts[3], " "), ".\n")
		return fmt.Errorf("%s in GraphQL query string at location %s:%s", msg, errorParts[1], errorParts[2])
	}
	return nil
}

func getBuiltinSchema() *gqlast.SchemaDocument {
	schema, err := gqlparser.ParseSchema(gqlvalidator.Prelude)
	if err != nil {
		panic(fmt.Errorf("Error in gqlparser Prelude (should be impossible): %w", err))
	}
	return schema
}

// NOTE(philipc): This function expects *validated* schema documents, and will break
// if it is fed arbitrary structures.
func mergeSchemaDocuments(docA *gqlast.SchemaDocument, docB *gqlast.SchemaDocument) *gqlast.SchemaDocument {
	ast := &gqlast.SchemaDocument{}
	ast.Merge(docA)
	ast.Merge(docB)
	return ast
}

// Converts a SchemaDocument into a gqlast.Schema object that can be used for validation.
// It merges in the builtin schema typedefs exactly as gqltop.LoadSchema did internally.
func convertSchema(schemaDoc *gqlast.SchemaDocument) (*gqlast.Schema, error) {
	// Merge builtin schema + schema we were provided.
	builtinsSchemaDoc := getBuiltinSchema()
	mergedSchemaDoc := mergeSchemaDocuments(builtinsSchemaDoc, schemaDoc)
	schema, err := gqlvalidator.ValidateSchemaDocument(mergedSchemaDoc)
	if err != nil {
		return nil, fmt.Errorf("Error in gqlparser SchemaDocument to Schema conversion: %w", err)
	}
	return schema, nil
}

// Converts an ast.Object into a gqlast.QueryDocument object.
func objectToQueryDocument(value ast.Object) (*gqlast.QueryDocument, error) {
	// Convert ast.Term to interface{} for JSON encoding below.
	asJSON, err := ast.JSON(value)
	if err != nil {
		return nil, err
	}
	// Marshal to JSON.
	bs, err := json.Marshal(asJSON)
	if err != nil {
		return nil, err
	}
	// Unmarshal from JSON -> gqlast.QueryDocument.
	var result gqlast.QueryDocument
	err = json.Unmarshal(bs, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Converts an ast.Object into a gqlast.SchemaDocument object.
func objectToSchemaDocument(value ast.Object) (*gqlast.SchemaDocument, error) {
	// Convert ast.Term to interface{} for JSON encoding below.
	asJSON, err := ast.JSON(value)
	if err != nil {
		return nil, err
	}
	// Marshal to JSON.
	bs, err := json.Marshal(asJSON)
	if err != nil {
		return nil, err
	}
	// Unmarshal from JSON -> gqlast.SchemaDocument.
	var result gqlast.SchemaDocument
	err = json.Unmarshal(bs, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Recursively traverses an AST that has been run through InterfaceToValue,
// and prunes away the fields with null or empty values, and all `Position`
// structs.
// NOTE(philipc): We currently prune away null values to reduce the level
// of clutter in the returned AST objects. In the future, if there is demand
// for ASTs that have a more regular/fixed structure, we may need to provide
// a "raw" version of the AST, where we still prune away the `Position`
// structs, but leave in the null fields.
func pruneIrrelevantGraphQLASTNodes(value ast.Value) ast.Value {
	// We iterate over the Value we've been provided, and recurse down
	// in the case of complex types, such as Arrays/Objects.
	// We are guaranteed to only have to deal with standard JSON types,
	// so this is much less ugly than what we'd need for supporting every
	// extant ast type!
	switch x := value.(type) {
	case *ast.Array:
		result := ast.NewArray()
		// Iterate over the array's elements, and do the following:
		// - Drop any Nulls
		// - Drop any any empty object/array value (after running the pruner)
		for i := range x.Len() {
			vTerm := x.Elem(i)
			switch v := vTerm.Value.(type) {
			case ast.Null:
				continue
			case *ast.Array:
				// Safe, because we knew the type before going to prune it.
				va := pruneIrrelevantGraphQLASTNodes(v).(*ast.Array)
				if va.Len() > 0 {
					result = result.Append(ast.NewTerm(va))
				}
			case ast.Object:
				// Safe, because we knew the type before going to prune it.
				vo := pruneIrrelevantGraphQLASTNodes(v).(ast.Object)
				if vo.Len() > 0 {
					result = result.Append(ast.NewTerm(vo))
				}
			default:
				result = result.Append(vTerm)
			}
		}
		return result
	case ast.Object:
		result := ast.NewObject()
		// Iterate over our object's keys, and do the following:
		// - Drop "Position".
		// - Drop any key with a Null value.
		// - Drop any key with an empty object/array value (after running the pruner)
		keys := x.Keys()
		for _, k := range keys {
			// We drop the "Position" objects because we don't need the
			// source-backref/location info they provide for policy rules.
			// Note that keys are ast.Strings.
			if ast.String("Position").Equal(k.Value) {
				continue
			}
			vTerm := x.Get(k)
			switch v := vTerm.Value.(type) {
			case ast.Null:
				continue
			case *ast.Array:
				// Safe, because we knew the type before going to prune it.
				va := pruneIrrelevantGraphQLASTNodes(v).(*ast.Array)
				if va.Len() > 0 {
					result.Insert(k, ast.NewTerm(va))
				}
			case ast.Object:
				// Safe, because we knew the type before going to prune it.
				vo := pruneIrrelevantGraphQLASTNodes(v).(ast.Object)
				if vo.Len() > 0 {
					result.Insert(k, ast.NewTerm(vo))
				}
			default:
				result.Insert(k, vTerm)
			}
		}
		return result
	default:
		return x
	}
}

// Reports errors from parsing/validation.
func builtinGraphQLParse(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var queryDoc *gqlast.QueryDocument
	var schemaDoc *gqlast.SchemaDocument
	var schemaASTValue ast.Value
	var querySchema ast.Value
	var err error

	// Parse/translate query if it's a string/object.
	switch x := operands[0].Value.(type) {
	case ast.String:
		queryDoc, err = parseQuery(string(x))
	case ast.Object:
		queryDoc, err = objectToQueryDocument(x)
	default:
		// Error if wrong type.
		return builtins.NewOperandTypeErr(0, x, "string", "object")
	}
	if err != nil {
		return err
	}

	schemaCacheKey, schema := cacheGetSchema(bctx, operands[1])
	schemaASTCacheKey, querySchema := cacheGetSchemaAST(bctx, operands[1])
	if schema == nil || querySchema == nil {
		// Parse/translate schema if it's a string/object.
		switch x := operands[1].Value.(type) {
		case ast.String:
			schemaDoc, err = parseSchema(string(x))
		case ast.Object:
			schemaDoc, err = objectToSchemaDocument(x)
		default:
			// Error if wrong type.
			return builtins.NewOperandTypeErr(1, x, "string", "object")
		}
		if err != nil {
			return err
		}

		// Convert SchemaDoc to Object before validating and converting it to a Schema
		// This precludes inclusion of extra definitions from the default GraphQL schema
		if querySchema == nil {
			schemaASTValue, err = ast.InterfaceToValue(schemaDoc)
			if err != nil {
				return err
			}
			querySchema = pruneIrrelevantGraphQLASTNodes(schemaASTValue.(ast.Object))
			cacheInsertSchemaAST(bctx, schemaASTCacheKey, querySchema)
		}

		// Validate the query against the schema, erroring if there's an issue.
		if schema == nil {
			schema, err = convertSchema(schemaDoc)
			if err != nil {
				return err
			}
			cacheInsertSchema(bctx, schemaCacheKey, schema)
		}

	}
	// Transform the ASTs into Objects.
	queryASTValue, err := ast.InterfaceToValue(queryDoc)
	if err != nil {
		return err
	}

	if err := validateQuery(schema, queryDoc); err != nil {
		return err
	}

	// Recursively remove irrelevant AST structures.
	queryResult := pruneIrrelevantGraphQLASTNodes(queryASTValue.(ast.Object))

	// Construct return value.
	verified := ast.ArrayTerm(
		ast.NewTerm(queryResult),
		ast.NewTerm(querySchema),
	)

	return iter(verified)
}

// Returns default value when errors occur.
func builtinGraphQLParseAndVerify(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var queryDoc *gqlast.QueryDocument
	var schemaDoc *gqlast.SchemaDocument
	var schemaASTValue ast.Value
	var querySchema ast.Value
	var err error

	unverified := ast.ArrayTerm(
		ast.InternedBooleanTerm(false),
		ast.NewTerm(ast.NewObject()),
		ast.NewTerm(ast.NewObject()),
	)

	// Parse/translate query if it's a string/object.
	switch x := operands[0].Value.(type) {
	case ast.String:
		queryDoc, err = parseQuery(string(x))
	case ast.Object:
		queryDoc, err = objectToQueryDocument(x)
	default:
		// Error if wrong type.
		return iter(unverified)
	}
	if err != nil {
		return iter(unverified)
	}

	// Transform the ASTs into Objects.
	queryASTValue, err := ast.InterfaceToValue(queryDoc)
	if err != nil {
		return iter(unverified)
	}

	schemaCacheKey, schema := cacheGetSchema(bctx, operands[1])
	schemaASTCacheKey, querySchema := cacheGetSchemaAST(bctx, operands[1])
	if schema == nil || querySchema == nil {
		// Parse/translate schema if it's a string/object.
		switch x := operands[1].Value.(type) {
		case ast.String:
			schemaDoc, err = parseSchema(string(x))
		case ast.Object:
			schemaDoc, err = objectToSchemaDocument(x)
		default:
			// Error if wrong type.
			return iter(unverified)
		}
		if err != nil {
			return iter(unverified)
		}

		// Convert SchemaDoc to Object before validating and converting it to a Schema
		// This precludes inclusion of extra definitions from the default GraphQL schema
		if querySchema == nil {
			schemaASTValue, err = ast.InterfaceToValue(schemaDoc)
			if err != nil {
				return iter(unverified)
			}
			querySchema = pruneIrrelevantGraphQLASTNodes(schemaASTValue.(ast.Object))
			cacheInsertSchemaAST(bctx, schemaASTCacheKey, querySchema)
		}

		if schema == nil {
			schema, err = convertSchema(schemaDoc)
			if err != nil {
				return iter(unverified)
			}
			cacheInsertSchema(bctx, schemaCacheKey, schema)
		}

	}

	// Validate the query against the schema, erroring if there's an issue.
	if err := validateQuery(schema, queryDoc); err != nil {
		return iter(unverified)
	}

	// Recursively remove irrelevant AST structures.
	queryResult := pruneIrrelevantGraphQLASTNodes(queryASTValue.(ast.Object))

	// Construct return value.
	verified := ast.ArrayTerm(
		ast.InternedBooleanTerm(true),
		ast.NewTerm(queryResult),
		ast.NewTerm(querySchema),
	)

	return iter(verified)
}

func builtinGraphQLParseQuery(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	raw, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Get the highly-nested AST struct, along with any errors generated.
	query, err := parseQuery(string(raw))
	if err != nil {
		return err
	}

	// Transform the AST into an Object.
	value, err := ast.InterfaceToValue(query)
	if err != nil {
		return err
	}

	// Recursively remove irrelevant AST structures.
	result := pruneIrrelevantGraphQLASTNodes(value.(ast.Object))

	return iter(ast.NewTerm(result))
}

func builtinGraphQLParseSchema(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	schemaDocCacheKey, schemaDoc := cacheGetSchemaDoc(bctx, operands[0])
	if schemaDoc == nil {
		raw, err := builtins.StringOperand(operands[0].Value, 1)
		if err != nil {
			return err
		}

		// Get the highly-nested AST struct, along with any errors generated.
		schemaDoc, err = parseSchema(string(raw))
		if err != nil {
			return err
		}
		// Note SchemaDoc is not validated
		cacheInsertSchemaDoc(bctx, schemaDocCacheKey, schemaDoc)
	}

	schemaASTCacheKey, schemaAST := cacheGetSchemaAST(bctx, operands[0])
	if schemaAST == nil {

		// Transform the AST into an Object.
		value, err := ast.InterfaceToValue(schemaDoc)
		if err != nil {
			return err
		}

		// Recursively remove irrelevant AST structures.
		schemaAST = pruneIrrelevantGraphQLASTNodes(value.(ast.Object))
		cacheInsertSchemaAST(bctx, schemaASTCacheKey, schemaAST)
	}
	return iter(ast.NewTerm(schemaAST))
}

func builtinGraphQLIsValid(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var queryDoc *gqlast.QueryDocument
	var schemaDoc *gqlast.SchemaDocument
	var schema *gqlast.Schema
	var err error

	switch x := operands[0].Value.(type) {
	case ast.String:
		queryDoc, err = parseQuery(string(x))
	case ast.Object:
		queryDoc, err = objectToQueryDocument(x)
	default:
		// Error if wrong type.
		return iter(ast.InternedBooleanTerm(false))
	}
	if err != nil {
		return iter(ast.InternedBooleanTerm(false))
	}

	schemaCacheKey, schema := cacheGetSchema(bctx, operands[1])
	if schema == nil {
		switch x := operands[1].Value.(type) {
		case ast.String:
			schemaDoc, err = parseSchema(string(x))
		case ast.Object:
			schemaDoc, err = objectToSchemaDocument(x)
		default:
			// Error if wrong type.
			return iter(ast.InternedBooleanTerm(false))
		}
		if err != nil {
			return iter(ast.InternedBooleanTerm(false))
		}

		// Validate the query against the schema, erroring if there's an issue.
		schema, err = convertSchema(schemaDoc)
		if err != nil {
			return iter(ast.InternedBooleanTerm(false))
		}
		cacheInsertSchema(bctx, schemaCacheKey, schema)
	}

	if err := validateQuery(schema, queryDoc); err != nil {
		return iter(ast.InternedBooleanTerm(false))
	}

	// If we got this far, the GraphQL query passed validation.
	return iter(ast.InternedBooleanTerm(true))
}

func builtinGraphQLSchemaIsValid(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var err error

	// Schemas are only cached if they are valid
	schemaCacheKey, schema := cacheGetSchema(bctx, operands[0])
	if schema == nil {
		var schemaDoc *gqlast.SchemaDocument
		var validatedSchema *gqlast.Schema

		switch x := operands[0].Value.(type) {
		case ast.String:
			schemaDoc, err = parseSchema(string(x))
		case ast.Object:
			schemaDoc, err = objectToSchemaDocument(x)
		default:
			// Error if wrong type.
			return iter(ast.InternedBooleanTerm(false))
		}
		if err != nil {
			return iter(ast.InternedBooleanTerm(false))
		}
		// Validate the schema, this determines the result
		// and whether there is a schema to cache
		validatedSchema, err = convertSchema(schemaDoc)
		if err == nil {
			cacheInsertSchema(bctx, schemaCacheKey, validatedSchema)
		}
	}

	return iter(ast.InternedBooleanTerm(err == nil))
}

// Insert Schema into cache
func cacheInsertSchema(bctx BuiltinContext, key string, schema *gqlast.Schema) {
	if bctx.InterQueryBuiltinValueCache == nil || key == "" {
		return
	}
	cacheKey := ast.String(key)
	c := bctx.InterQueryBuiltinValueCache.GetCache(gqlCacheName)
	if c == nil {
		return
	}
	c.Insert(cacheKey, schema)
}

// Insert SchemaAST into cache
func cacheInsertSchemaAST(bctx BuiltinContext, key string, schemaAST ast.Value) {
	if bctx.InterQueryBuiltinValueCache == nil || key == "" {
		return
	}
	cacheKeyAST := ast.String(key)
	c := bctx.InterQueryBuiltinValueCache.GetCache(gqlCacheName)
	if c == nil {
		return
	}
	c.Insert(cacheKeyAST, schemaAST)
}

// Insert SchemaDocument into cache
func cacheInsertSchemaDoc(bctx BuiltinContext, key string, schemaDoc *gqlast.SchemaDocument) {
	if bctx.InterQueryBuiltinValueCache == nil || key == "" {
		return
	}
	cacheKey := ast.String(key)
	c := bctx.InterQueryBuiltinValueCache.GetCache(gqlCacheName)
	if c == nil {
		return
	}
	c.Insert(cacheKey, schemaDoc)
}

// Returns the cache key and a Schema if this key already exists in the cache
func cacheGetSchema(bctx BuiltinContext, t *ast.Term) (string, *gqlast.Schema) {
	if bctx.InterQueryBuiltinValueCache != nil {
		if c := bctx.InterQueryBuiltinValueCache.GetCache(gqlCacheName); c != nil {
			if key, keyOk := cacheKeyWithPrefix(bctx, t, "gql_schema-"); keyOk {
				if val, ok := c.Get(ast.String(key)); ok {
					if schema, isSchema := val.(*gqlast.Schema); isSchema {
						return key, schema
					}
				}
				return key, nil
			}
		}
	}
	return "", nil
}

// Returns the cache key and a SchemaDocument if this key already exists in the cache
// Note: the SchemaDocument is not a validated Schema
func cacheGetSchemaDoc(bctx BuiltinContext, t *ast.Term) (string, *gqlast.SchemaDocument) {
	if bctx.InterQueryBuiltinValueCache != nil {
		if c := bctx.InterQueryBuiltinValueCache.GetCache(gqlCacheName); c != nil {
			if key, keyOk := cacheKeyWithPrefix(bctx, t, "gql_schema_doc-"); keyOk {
				if val, ok := c.Get(ast.String(key)); ok {
					if schemaDoc, isSchemaDoc := val.(*gqlast.SchemaDocument); isSchemaDoc {
						return key, schemaDoc
					}
				}
				return key, nil
			}
		}
	}
	return "", nil
}

// Returns the cache key and a SchemaDocument if this key already exists in the cache
// Note: the AST should be pruned
func cacheGetSchemaAST(bctx BuiltinContext, t *ast.Term) (string, ast.Value) {
	if bctx.InterQueryBuiltinValueCache != nil {
		if c := bctx.InterQueryBuiltinValueCache.GetCache(gqlCacheName); c != nil {
			if key, keyOk := cacheKeyWithPrefix(bctx, t, "gql_schema_ast-"); keyOk {
				if val, ok := c.Get(ast.String(key)); ok {
					if schemaAST, isSchemaAST := val.(ast.Value); isSchemaAST {
						return key, schemaAST
					}
				}
				return key, nil
			}
		}
	}
	return "", nil
}

// Compute a constant size key for use with the cache
func cacheKeyWithPrefix(bctx BuiltinContext, t *ast.Term, prefix string) (string, bool) {
	var cacheKey ast.String
	var ok = false

	if bctx.InterQueryBuiltinValueCache != nil {
		switch t.Value.(type) {
		case ast.String:
			err := builtinCryptoSha256(bctx, []*ast.Term{t}, func(term *ast.Term) error {
				cacheKey = term.Value.(ast.String)
				return nil
			})
			ok = (len(cacheKey) > 0) && (err == nil)
		case ast.Object:
			objTerm := ast.StringTerm(t.String())
			err := builtinCryptoSha256(bctx, []*ast.Term{objTerm}, func(term *ast.Term) error {
				cacheKey = term.Value.(ast.String)
				return nil
			})
			ok = (len(cacheKey) > 0) && (err == nil)
		default:
			ok = false
		}
	}

	return prefix + string(cacheKey), ok
}

const gqlCacheName = "graphql"

func init() {

	var defaultCacheEntries int = 10
	var graphqlCacheConfig = cache.NamedValueCacheConfig{
		MaxNumEntries: &defaultCacheEntries,
	}
	cache.RegisterDefaultInterQueryBuiltinValueCacheConfig(gqlCacheName, &graphqlCacheConfig)

	RegisterBuiltinFunc(ast.GraphQLParse.Name, builtinGraphQLParse)
	RegisterBuiltinFunc(ast.GraphQLParseAndVerify.Name, builtinGraphQLParseAndVerify)
	RegisterBuiltinFunc(ast.GraphQLParseQuery.Name, builtinGraphQLParseQuery)
	RegisterBuiltinFunc(ast.GraphQLParseSchema.Name, builtinGraphQLParseSchema)
	RegisterBuiltinFunc(ast.GraphQLIsValid.Name, builtinGraphQLIsValid)
	RegisterBuiltinFunc(ast.GraphQLSchemaIsValid.Name, builtinGraphQLSchemaIsValid)
}
