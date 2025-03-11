// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

// Helper: sets of vertices can be represented as Arrays or Sets.
func foreachVertex(collection *ast.Term, f func(*ast.Term)) {
	switch v := collection.Value.(type) {
	case ast.Set:
		v.Foreach(f)
	case *ast.Array:
		v.Foreach(f)
	}
}

// numberOfEdges returns the number of elements of an array or a set (of edges)
func numberOfEdges(collection *ast.Term) int {
	switch v := collection.Value.(type) {
	case ast.Set:
		return v.Len()
	case *ast.Array:
		return v.Len()
	}

	return 0
}

func builtinReachable(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Error on wrong types for args.
	graph, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	var queue []*ast.Term
	switch initial := operands[1].Value.(type) {
	case *ast.Array, ast.Set:
		foreachVertex(ast.NewTerm(initial), func(t *ast.Term) {
			queue = append(queue, t)
		})
	default:
		return builtins.NewOperandTypeErr(2, initial, "{array, set}")
	}

	// This is the set of nodes we have reached.
	reached := ast.NewSet()

	// Keep going as long as we have nodes in the queue.
	for len(queue) > 0 {
		// Get the edges for this node.  If the node was not in the graph,
		// `edges` will be `nil` and we can ignore it.
		node := queue[0]
		if edges := graph.Get(node); edges != nil {
			// Add all the newly discovered neighbors.
			foreachVertex(edges, func(neighbor *ast.Term) {
				if !reached.Contains(neighbor) {
					queue = append(queue, neighbor)
				}
			})
			// Mark the node as reached.
			reached.Add(node)
		}
		queue = queue[1:]
	}

	return iter(ast.NewTerm(reached))
}

// pathBuilder is called recursively to build a Set of paths that are reachable from the root
func pathBuilder(graph ast.Object, root *ast.Term, path []*ast.Term, edgeRslt ast.Set, reached ast.Set) {
	paths := []*ast.Term{}

	if edges := graph.Get(root); edges != nil {
		path = append(path, root)

		if numberOfEdges(edges) >= 1 {

			foreachVertex(edges, func(neighbor *ast.Term) {

				if reached.Contains(neighbor) {
					// If we've already reached this node, return current path (avoid infinite recursion)
					paths = append(paths, path...)
					edgeRslt.Add(ast.ArrayTerm(paths...))
				} else {
					reached.Add(root)
					pathBuilder(graph, neighbor, path, edgeRslt, reached)

				}

			})

		} else {
			paths = append(paths, path...)
			edgeRslt.Add(ast.ArrayTerm(paths...))

		}
	} else {
		// Node is nonexistent (not in graph). Commit the current path (without adding this root)
		paths = append(paths, path...)
		edgeRslt.Add(ast.ArrayTerm(paths...))

	}

}

func builtinReachablePaths(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var traceResult = ast.NewSet()
	// Error on wrong types for args.
	graph, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// This is a queue that holds all nodes we still need to visit.  It is
	// initialised to the initial set of nodes we start out with.
	var queue []*ast.Term
	switch initial := operands[1].Value.(type) {
	case *ast.Array, ast.Set:
		foreachVertex(ast.NewTerm(initial), func(t *ast.Term) {
			queue = append(queue, t)
		})
	default:
		return builtins.NewOperandTypeErr(2, initial, "{array, set}")
	}

	for _, node := range queue {
		// Find reachable paths from edges in root node in queue and append arrays to the results set
		if edges := graph.Get(node); edges != nil {
			if numberOfEdges(edges) >= 1 {
				foreachVertex(edges, func(neighbor *ast.Term) {
					pathBuilder(graph, neighbor, []*ast.Term{node}, traceResult, ast.NewSet(node))
				})
			} else {
				traceResult.Add(ast.ArrayTerm(node))
			}
		}
	}

	return iter(ast.NewTerm(traceResult))
}

func init() {
	RegisterBuiltinFunc(ast.ReachableBuiltin.Name, builtinReachable)
	RegisterBuiltinFunc(ast.ReachablePathsBuiltin.Name, builtinReachablePaths)
}
