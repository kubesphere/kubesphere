/*

 Copyright 2019 The KubeSphere Authors.

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
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("resource-gen: ")
	flag.Parse()

	fileSet := token.NewFileSet()
	target := os.Getenv("GOFILE")
	log.Println("start parse", target)
	f, err := parser.ParseFile(fileSet, target, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	imports := make(map[string]string, 0)
	lister := make(map[string][2]string, 0)

	for _, i := range f.Imports {
		if i.Name != nil {
			imports[i.Name.Name] = i.Path.Value
		} else {
			j := strings.LastIndex(i.Path.Value, "/")
			if j > 0 {
				imports[i.Path.Value[j+1:]] = i.Path.Value
			} else {
				imports[i.Path.Value] = i.Path.Value
			}
		}
	}
	for _, v := range f.Scope.Objects {

		if v.Kind != ast.Var || !strings.HasSuffix(v.Name, "Lister") {
			continue
		}

		if d, ok := v.Decl.(*ast.ValueSpec); ok {
			if t, ok := d.Type.(*ast.SelectorExpr); ok {
				if strings.HasSuffix(t.Sel.Name, "Lister") {
					lister[t.Sel.Name] = [2]string{t.X.(*ast.Ident).Name, v.Name}
				}
			}
		}
	}

	src := genString(imports, lister)

	baseName := fmt.Sprintf("getter.go")
	outputName := filepath.Join(".", strings.ToLower(baseName))
	err = ioutil.WriteFile(outputName, src, 0644)

	if err != nil {
		log.Fatalln(err)
	}
}

func genString(imports map[string]string, lister map[string][2]string) []byte {
	const strTmp = `
	package {{.pkg}}
    import (
    {{range $index,$str :=.imports}}
	 {{$str}}
	{{end}}
    )

    {{range $l,$m :=.lister}}
    func Get{{$l}}() {{index $m 0}}.{{$l}} {
	return {{index $m 1}}
    }
	{{end}}
	`
	pkgName := os.Getenv("GOPACKAGE")

	i := make([]string, 0)
	for _, v := range lister {
		if imports[v[0]] == "" {
			log.Panicln("import xxxxxx")
		} else {
			str := fmt.Sprintln(v[0], imports[v[0]])
			if !hasString(i, str) {
				i = append(i, fmt.Sprintln(v[0], imports[v[0]]))
			}
		}
	}

	data := map[string]interface{}{
		"pkg":     pkgName,
		"imports": i,
		"lister":  lister,
	}
	//利用模板库，生成代码文件
	t, err := template.New("").Parse(strTmp)
	if err != nil {
		log.Fatal(err)
	}
	buff := bytes.NewBufferString("")
	err = t.Execute(buff, data)
	if err != nil {
		log.Fatal(err)
	}
	//进行格式化
	src, err := format.Source(buff.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return src
}

func hasString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
