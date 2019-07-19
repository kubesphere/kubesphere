

go-hashids [![Build Status](https://ci.appveyor.com/api/projects/status/1s8yeafycpa2vdaq?svg=true)](https://ci.appveyor.com/project/speps/go-hashids) [![GoDoc](https://godoc.org/github.com/speps/go-hashids?status.svg)](https://godoc.org/github.com/speps/go-hashids)
==========

Go (golang) v1 implementation of http://www.hashids.org
under MIT License (same as the original implementations)

Original implementations by [Ivan Akimov](https://github.com/ivanakimov)

### Setup
<pre>go get github.com/speps/go-hashids</pre>

CLI tool :

<pre>go get github.com/speps/go-hashids/cmd/hashid</pre>

### Example
```go
package main

import "fmt"
import "github.com/speps/go-hashids"

func main() {
	hd := hashids.NewData()
	hd.Salt = "this is my salt"
	hd.MinLength = 30
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{45, 434, 1313, 99})
	fmt.Println(e)
	d, _ := h.DecodeWithError(e)
	fmt.Println(d)
}
```

### Thanks to all the contributors

* [Harm Aarts](https://github.com/haarts)
* [Christoffer G. Thomsen](https://github.com/cgt)
* [Peter Hellberg](https://github.com/peterhellberg)
* [Rémy Oudompheng](https://github.com/remyoudompheng)
* [Mart Roosmaa](https://github.com/roosmaa)
* [Jakub Kramarz](https://github.com/jkramarz)
* [Zou Xifeng](https://github.com/zouxifeng)
* [Per Persson](https://github.com/md2perpe)
* [Baiju Muthukadan](https://github.com/baijum)
* [Pablo de la Concepción Sanz](https://github.com/pconcepcion)
* [Olivier Mengué](https://github.com/dolmen)
* [Matthew Valimaki](https://github.com/matthewvalimaki)
* [Cody Maloney](https://github.com/cmaloney)

Let me know if I forgot anyone of course.

### Changelog

2017/05/09

* Changed API
	* `New` methods now return errors
	* Added sanity check in `Decode` that makes sure that the salt is consistent

2014/09/13

* Updated to Hashids v1.0.0 (should be compatible with other implementations, let me know if not, was checked against the Javascript version)
* Changed API
	* Encrypt/Decrypt are now Encode/Decode
	* HashID is now constructed from HashIDData containing alphabet, salt and minimum length
