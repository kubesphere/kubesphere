/*
Copyright 2019 The Kubernetes Authors.
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

package typed

import (
	"sigs.k8s.io/structured-merge-diff/fieldpath"
	"sigs.k8s.io/structured-merge-diff/schema"
	"sigs.k8s.io/structured-merge-diff/value"
)

type removingWalker struct {
	value    *value.Value
	schema   *schema.Schema
	toRemove *fieldpath.Set
}

func removeItemsWithSchema(value *value.Value, toRemove *fieldpath.Set, schema *schema.Schema, typeRef schema.TypeRef) {
	w := &removingWalker{
		value:    value,
		schema:   schema,
		toRemove: toRemove,
	}
	resolveSchema(schema, typeRef, value, w)
}

// doLeaf should be called on leaves before descending into children, if there
// will be a descent. It modifies w.inLeaf.
func (w *removingWalker) doLeaf() ValidationErrors { return nil }

func (w *removingWalker) doScalar(t schema.Scalar) ValidationErrors { return nil }

func (w *removingWalker) doList(t schema.List) (errs ValidationErrors) {
	l := w.value.ListValue

	// If list is null, empty, or atomic just return
	if l == nil || len(l.Items) == 0 || t.ElementRelationship == schema.Atomic {
		return nil
	}

	newItems := []value.Value{}
	for i := range l.Items {
		item := l.Items[i]
		// Ignore error because we have already validated this list
		pe, _ := listItemToPathElement(t, i, item)
		path, _ := fieldpath.MakePath(pe)
		if w.toRemove.Has(path) {
			continue
		}
		if subset := w.toRemove.WithPrefix(pe); !subset.Empty() {
			removeItemsWithSchema(&l.Items[i], subset, w.schema, t.ElementType)
		}
		newItems = append(newItems, l.Items[i])
	}
	l.Items = newItems
	if len(l.Items) == 0 {
		w.value.ListValue = nil
		w.value.Null = true
	}
	return nil
}

func (w *removingWalker) doMap(t schema.Map) ValidationErrors {
	m := w.value.MapValue

	// If map is null, empty, or atomic just return
	if m == nil || len(m.Items) == 0 || t.ElementRelationship == schema.Atomic {
		return nil
	}

	fieldTypes := map[string]schema.TypeRef{}
	for _, structField := range t.Fields {
		fieldTypes[structField.Name] = structField.Type
	}

	newMap := &value.Map{}
	for i := range m.Items {
		item := m.Items[i]
		pe := fieldpath.PathElement{FieldName: &item.Name}
		path, _ := fieldpath.MakePath(pe)
		fieldType := t.ElementType
		if ft, ok := fieldTypes[item.Name]; ok {
			fieldType = ft
		} else {
			if w.toRemove.Has(path) {
				continue
			}
		}
		if subset := w.toRemove.WithPrefix(pe); !subset.Empty() {
			removeItemsWithSchema(&m.Items[i].Value, subset, w.schema, fieldType)
		}
		newMap.Set(item.Name, m.Items[i].Value)
	}
	w.value.MapValue = newMap
	if len(w.value.MapValue.Items) == 0 {
		w.value.MapValue = nil
		w.value.Null = true
	}
	return nil
}

func (*removingWalker) errorf(_ string, _ ...interface{}) ValidationErrors { return nil }
