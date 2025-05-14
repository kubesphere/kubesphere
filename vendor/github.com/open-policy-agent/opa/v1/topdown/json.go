// Copyright 2019 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"

	"github.com/open-policy-agent/opa/internal/edittree"
)

func builtinJSONRemove(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Expect an object and a string or array/set of strings
	_, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Build a list of json pointers to remove
	paths, err := getJSONPaths(operands[1].Value)
	if err != nil {
		return err
	}

	newObj, err := jsonRemove(operands[0], ast.NewTerm(pathsToObject(paths)))
	if err != nil {
		return err
	}

	if newObj == nil {
		return nil
	}

	return iter(newObj)
}

// jsonRemove returns a new term that is the result of walking
// through a and omitting removing any values that are in b but
// have ast.Null values (ie leaf nodes for b).
func jsonRemove(a *ast.Term, b *ast.Term) (*ast.Term, error) {
	if b == nil {
		// The paths diverged, return a
		return a, nil
	}

	var bObj ast.Object
	switch bValue := b.Value.(type) {
	case ast.Object:
		bObj = bValue
	case ast.Null:
		// Means we hit a leaf node on "b", dont add the value for a
		return nil, nil
	default:
		// The paths diverged, return a
		return a, nil
	}

	switch aValue := a.Value.(type) {
	case ast.String, ast.Number, ast.Boolean, ast.Null:
		return a, nil
	case ast.Object:
		newObj := ast.NewObject()
		err := aValue.Iter(func(k *ast.Term, v *ast.Term) error {
			// recurse and add the diff of sub objects as needed
			diffValue, err := jsonRemove(v, bObj.Get(k))
			if err != nil || diffValue == nil {
				return err
			}
			newObj.Insert(k, diffValue)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return ast.NewTerm(newObj), nil
	case ast.Set:
		newSet := ast.NewSet()
		err := aValue.Iter(func(v *ast.Term) error {
			// recurse and add the diff of sub objects as needed
			diffValue, err := jsonRemove(v, bObj.Get(v))
			if err != nil || diffValue == nil {
				return err
			}
			newSet.Add(diffValue)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return ast.NewTerm(newSet), nil
	case *ast.Array:
		// When indexes are removed we shift left to close empty spots in the array
		// as per the JSON patch spec.
		newArray := ast.NewArray()
		for i := range aValue.Len() {
			v := aValue.Elem(i)
			// recurse and add the diff of sub objects as needed
			// Note: Keys in b will be strings for the index, eg path /a/1/b => {"a": {"1": {"b": null}}}
			diffValue, err := jsonRemove(v, bObj.Get(ast.StringTerm(strconv.Itoa(i))))
			if err != nil {
				return nil, err
			}
			if diffValue != nil {
				newArray = newArray.Append(diffValue)
			}
		}
		return ast.NewTerm(newArray), nil
	default:
		return nil, fmt.Errorf("invalid value type %T", a)
	}
}

func builtinJSONFilter(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Ensure we have the right parameters, expect an object and a string or array/set of strings
	obj, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Build a list of filter strings
	filters, err := getJSONPaths(operands[1].Value)
	if err != nil {
		return err
	}

	// Actually do the filtering
	filterObj := pathsToObject(filters)
	r, err := obj.Filter(filterObj)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(r))
}

func getJSONPaths(operand ast.Value) ([]ast.Ref, error) {
	var paths []ast.Ref

	switch v := operand.(type) {
	case *ast.Array:
		for i := range v.Len() {
			filter, err := parsePath(v.Elem(i))
			if err != nil {
				return nil, err
			}
			paths = append(paths, filter)
		}
	case ast.Set:
		err := v.Iter(func(f *ast.Term) error {
			filter, err := parsePath(f)
			if err != nil {
				return err
			}
			paths = append(paths, filter)
			return nil
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, builtins.NewOperandTypeErr(2, v, "set", "array")
	}

	return paths, nil
}

func parsePath(path *ast.Term) (ast.Ref, error) {
	// paths can either be a `/` separated json path or
	// an array or set of values
	var pathSegments ast.Ref
	switch p := path.Value.(type) {
	case ast.String:
		if p == "" {
			return ast.Ref{}, nil
		}
		parts := strings.Split(strings.TrimLeft(string(p), "/"), "/")
		for _, part := range parts {
			part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
			pathSegments = append(pathSegments, ast.StringTerm(part))
		}
	case *ast.Array:
		p.Foreach(func(term *ast.Term) {
			pathSegments = append(pathSegments, term)
		})
	default:
		return nil, builtins.NewOperandErr(2, "must be one of {set, array} containing string paths or array of path segments but got %v", ast.ValueName(p))
	}

	return pathSegments, nil
}

func pathsToObject(paths []ast.Ref) ast.Object {
	root := ast.NewObject()

	for _, path := range paths {
		node := root
		var done bool

		// If the path is an empty JSON path, skip all further processing.
		if len(path) == 0 {
			done = true
		}

		// Otherwise, we should have 1+ path segments to work with.
		for i := 0; i < len(path)-1 && !done; i++ {

			k := path[i]
			child := node.Get(k)

			if child == nil {
				obj := ast.NewObject()
				node.Insert(k, ast.NewTerm(obj))
				node = obj
				continue
			}

			switch v := child.Value.(type) {
			case ast.Null:
				done = true
			case ast.Object:
				node = v
			default:
				panic("unreachable")
			}
		}

		if !done {
			node.Insert(path[len(path)-1], ast.InternedNullTerm)
		}
	}

	return root
}

type jsonPatch struct {
	op    string
	path  *ast.Term
	from  *ast.Term
	value *ast.Term
}

func getPatch(o ast.Object) (jsonPatch, error) {
	validOps := map[string]struct{}{"add": {}, "remove": {}, "replace": {}, "move": {}, "copy": {}, "test": {}}
	var out jsonPatch
	var ok bool
	getAttribute := func(attr string) (*ast.Term, error) {
		if term := o.Get(ast.StringTerm(attr)); term != nil {
			return term, nil
		}

		return nil, fmt.Errorf("missing '%s' attribute", attr)
	}

	opTerm, err := getAttribute("op")
	if err != nil {
		return out, err
	}
	op, ok := opTerm.Value.(ast.String)
	if !ok {
		return out, errors.New("attribute 'op' must be a string")
	}
	out.op = string(op)
	if _, found := validOps[out.op]; !found {
		out.op = ""
		return out, fmt.Errorf("unrecognized op '%s'", string(op))
	}

	pathTerm, err := getAttribute("path")
	if err != nil {
		return out, err
	}
	out.path = pathTerm

	// Only fetch the "from" parameter for move/copy ops.
	switch out.op {
	case "move", "copy":
		fromTerm, err := getAttribute("from")
		if err != nil {
			return out, err
		}
		out.from = fromTerm
	}

	// Only fetch the "value" parameter for add/replace/test ops.
	switch out.op {
	case "add", "replace", "test":
		valueTerm, err := getAttribute("value")
		if err != nil {
			return out, err
		}
		out.value = valueTerm
	}

	return out, nil
}

func applyPatches(source *ast.Term, operations *ast.Array) (*ast.Term, error) {
	et := edittree.NewEditTree(source)
	for i := range operations.Len() {
		object, ok := operations.Elem(i).Value.(ast.Object)
		if !ok {
			return nil, errors.New("must be an array of JSON-Patch objects, but at least one element is not an object")
		}
		patch, err := getPatch(object)
		if err != nil {
			return nil, err
		}
		path, err := parsePath(patch.path)
		if err != nil {
			return nil, err
		}

		switch patch.op {
		case "add":
			_, err = et.InsertAtPath(path, patch.value)
			if err != nil {
				return nil, err
			}
		case "remove":
			_, err = et.DeleteAtPath(path)
			if err != nil {
				return nil, err
			}
		case "replace":
			_, err = et.DeleteAtPath(path)
			if err != nil {
				return nil, err
			}
			_, err = et.InsertAtPath(path, patch.value)
			if err != nil {
				return nil, err
			}
		case "move":
			from, err := parsePath(patch.from)
			if err != nil {
				return nil, err
			}
			chunk, err := et.RenderAtPath(from)
			if err != nil {
				return nil, err
			}
			_, err = et.DeleteAtPath(from)
			if err != nil {
				return nil, err
			}
			_, err = et.InsertAtPath(path, chunk)
			if err != nil {
				return nil, err
			}
		case "copy":
			from, err := parsePath(patch.from)
			if err != nil {
				return nil, err
			}
			chunk, err := et.RenderAtPath(from)
			if err != nil {
				return nil, err
			}
			_, err = et.InsertAtPath(path, chunk)
			if err != nil {
				return nil, err
			}
		case "test":
			chunk, err := et.RenderAtPath(path)
			if err != nil {
				return nil, err
			}
			if !chunk.Equal(patch.value) {
				return nil, fmt.Errorf("value from EditTree != patch value.\n\nExpected: %v\n\nFound: %v", patch.value, chunk)
			}
		}
	}
	final := et.Render()
	// TODO: Nil check here?
	return final, nil
}

func builtinJSONPatch(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// JSON patch supports arrays, objects as well as values as the target.
	target := ast.NewTerm(operands[0].Value)

	// Expect an array of operations.
	operations, err := builtins.ArrayOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	patched, err := applyPatches(target, operations)
	if err != nil {
		return nil
	}
	return iter(patched)
}

func init() {
	RegisterBuiltinFunc(ast.JSONFilter.Name, builtinJSONFilter)
	RegisterBuiltinFunc(ast.JSONRemove.Name, builtinJSONRemove)
	RegisterBuiltinFunc(ast.JSONPatch.Name, builtinJSONPatch)
}
