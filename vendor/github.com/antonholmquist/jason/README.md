
![logo](https://us-bucket-host.s3.amazonaws.com/jason/jason_cropped_4.svg)

Jason is an easy-to-use JSON library for Go.

[![Build Status](https://img.shields.io/travis/antonholmquist/jason.svg?style=flat)](https://travis-ci.org/antonholmquist/jason) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/antonholmquist/jason) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/antonholmquist/jason/master/LICENSE)

# About

Jason is designed to be convenient for reading arbitrary JSON while still honoring the strictness of the language. Inspired by other libraries and improved to work well for common use cases. It currently focuses on reading JSON data rather than creating it. [API Documentation](http://godoc.org/github.com/antonholmquist/jason) can be found on godoc.org.

## Install

```shell
go get github.com/antonholmquist/jason
```

## Import

```go
import (
  "github.com/antonholmquist/jason"
)
```

## Data types

The following golang values are used to represent JSON data types. It is consistent with how `encoding/json` uses primitive types.

- `bool`, for JSON booleans
- `json.Number/float64/int64`, for JSON numbers
- `string`, for JSON strings
- `[]*Value`, for JSON arrays
- `map[string]*Value`, for JSON objects
- `nil` for JSON null

## Examples

### Create from bytes

Create object from bytes. Returns an error if the bytes are not valid JSON.

```go
v, err := jason.NewObjectFromBytes(b)

```

If the root object is not an array, use this method instead. It can then be cased to the expected type with one of the As-Methods.

```go
v, err := jason.NewValueFromBytes(b)

```

### Create from a reader (like a http response)

Create value from a io.reader. Returns an error if the string couldn't be parsed.

```go
v, err := jason.NewObjectFromReader(res.Body)

```

### Read values

Reading values is easy. If the key path is invalid or type doesn't match, it will return an error and the default value.

```go
name, err := v.GetString("name")
age, err := v.GetInt64("age")
verified, err := v.GetBoolean("verified")
education, err := v.GetObject("education")
friends, err := v.GetObjectArray("friends")
interests, err := v.GetStringArray("interests")

```

### Read nested values

Reading nested values is easy. If the path is invalid or type doesn't match, it will return the default value and an error.

```go
name, err := v.GetString("person", "name")
age, err := v.GetInt64("person", "age")
verified, err := v.GetBoolean("person", "verified")
education, err := v.GetObject("person", "education")
friends, err := v.GetObjectArray("person", "friends")

```

### Loop through array

Looping through an array is done with `GetValueArray()` or `GetObjectArray()`. It returns an error if the value at that keypath is null (or something else than an array).

```go
friends, err := person.GetObjectArray("friends")
for _, friend := range friends {
  name, err := friend.GetString("name")
  age, err := friend.GetNumber("age")
}
```

### Loop through object

Looping through an object is easy. `GetObject()` returns an error if the value at that keypath is null (or something else than an object).

```go
person, err := person.GetObject("person")
for key, value := range person.Map() {
  ...
}
```

## Sample App

Example project:

```go
package main

import (
  "github.com/antonholmquist/jason"
  "log"
)

func main() {

  exampleJSON := `{
    "name": "Walter White",
    "age": 51,
    "children": [
      "junior",
      "holly"
    ],
    "other": {
      "occupation": "chemist",
      "years": 23
    }
  }`

  v, _ := jason.NewObjectFromBytes([]byte(exampleJSON))

  name, _ := v.GetString("name")
  age, _ := v.GetNumber("age")
  occupation, _ := v.GetString("other", "occupation")
  years, _ := v.GetNumber("other", "years")

  log.Println("age:", age)
  log.Println("name:", name)
  log.Println("occupation:", occupation)
  log.Println("years:", years)

  children, _ := v.GetStringArray("children")
  for i, child := range children {
    log.Printf("child %d: %s", i, child)
  }

  others, _ := v.GetObject("other")

  for _, value := range others.Map() {

    s, sErr := value.String()
    n, nErr := value.Number()

    if sErr == nil {
      log.Println("string value: ", s)
    } else if nErr == nil {
      log.Println("number value: ", n)
    }
  }
}

```

## Documentation

Documentation can be found a godoc:

https://godoc.org/github.com/antonholmquist/jason

## Test
To run the project tests:

```shell
go test
```

## Compatibility

Go 1.1 and up.

## Where does the name come from?

I remebered it from an email one of our projects managers sent a couple of years ago.

> "Don't worry. We can handle both XML and Jason"

## Author

Anton Holmquist, http://twitter.com/antonholmquist
