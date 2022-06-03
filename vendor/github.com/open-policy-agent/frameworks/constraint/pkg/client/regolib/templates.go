package regolib

import "text/template"

var TargetLib = template.Must(template.New("TargetLib").Parse(targetLibSrc))
