package filter

import "io/fs"

type LoaderFilter func(abspath string, info fs.FileInfo, depth int) bool
