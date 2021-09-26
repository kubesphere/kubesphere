package regolib

import "text/template"

// TargetLib is the a parsed text template for Rego tests.
var TargetLib = template.Must(template.New("TargetLib").Parse(targetLibSrc))
