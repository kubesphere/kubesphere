csvutil [![PkgGoDev](https://pkg.go.dev/badge/github.com/jszwec/csvutil@v1.4.0?tab=doc)](https://pkg.go.dev/github.com/jszwec/csvutil?tab=doc) ![Go](https://github.com/jszwec/csvutil/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/jszwec/csvutil)](https://goreportcard.com/report/github.com/jszwec/csvutil) [![codecov](https://codecov.io/gh/jszwec/csvutil/branch/master/graph/badge.svg)](https://codecov.io/gh/jszwec/csvutil)
=================

<p align="center">
  <img style="float: right;" src="https://user-images.githubusercontent.com/3941256/33054906-52b4bc08-ce4a-11e7-9651-b70c5a47c921.png"/ width=200>
</p>

Package csvutil provides fast, idiomatic, and dependency free mapping between CSV and Go (golang) values.

This package is not a CSV parser, it is based on the [Reader](https://godoc.org/github.com/jszwec/csvutil#Reader) and [Writer](https://godoc.org/github.com/jszwec/csvutil#Writer)
interfaces which are implemented by eg. std Go (golang) [csv package](https://golang.org/pkg/encoding/csv). This gives a possibility
of choosing any other CSV writer or reader which may be more performant.

Installation
------------

    go get github.com/jszwec/csvutil

Requirements
-------------

* Go1.7+

Index
------

1. [Examples](#examples)
	1. [Unmarshal](#examples_unmarshal)
	2. [Marshal](#examples_marshal)
	3. [Unmarshal and metadata](#examples_unmarshal_and_metadata)
	4. [But my CSV file has no header...](#examples_but_my_csv_has_no_header)
	5. [Decoder.Map - data normalization](#examples_decoder_map)
	6. [Different separator/delimiter](#examples_different_separator)
	7. [Custom Types](#examples_custom_types)
	8. [Custom time.Time format](#examples_time_format)
	9. [Custom struct tags](#examples_struct_tags)
	10. [Slice and Map fields](#examples_slice_and_map_field)
	11. [Nested/Embedded structs](#examples_nested_structs)
	12. [Inline tag](#examples_inlined_structs)
2. [Performance](#performance)
	1. [Unmarshal](#performance_unmarshal)
	2. [Marshal](#performance_marshal)

Example <a name="examples"></a>
--------

### Unmarshal <a name="examples_unmarshal"></a>

Nice and easy Unmarshal is using the Go std [csv.Reader](https://golang.org/pkg/encoding/csv/#Reader) with its default options. Use [Decoder](https://godoc.org/github.com/jszwec/csvutil#Decoder) for streaming and more advanced use cases.

```go
	var csvInput = []byte(`
name,age,CreatedAt
jacek,26,2012-04-01T15:00:00Z
john,,0001-01-01T00:00:00Z`,
	)

	type User struct {
		Name      string `csv:"name"`
		Age       int    `csv:"age,omitempty"`
		CreatedAt time.Time
	}

	var users []User
	if err := csvutil.Unmarshal(csvInput, &users); err != nil {
		fmt.Println("error:", err)
	}

	for _, u := range users {
		fmt.Printf("%+v\n", u)
	}

	// Output:
	// {Name:jacek Age:26 CreatedAt:2012-04-01 15:00:00 +0000 UTC}
	// {Name:john Age:0 CreatedAt:0001-01-01 00:00:00 +0000 UTC}
```

### Marshal <a name="examples_marshal"></a>

Marshal is using the Go std [csv.Writer](https://golang.org/pkg/encoding/csv/#Writer) with its default options. Use [Encoder](https://godoc.org/github.com/jszwec/csvutil#Encoder) for streaming or to use a different Writer.

```go
	type Address struct {
		City    string
		Country string
	}

	type User struct {
		Name string
		Address
		Age       int `csv:"age,omitempty"`
		CreatedAt time.Time
	}

	users := []User{
		{
			Name:      "John",
			Address:   Address{"Boston", "USA"},
			Age:       26,
			CreatedAt: time.Date(2010, 6, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			Name:    "Alice",
			Address: Address{"SF", "USA"},
		},
	}

	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))

	// Output:
	// Name,City,Country,age,CreatedAt
	// John,Boston,USA,26,2010-06-02T12:00:00Z
	// Alice,SF,USA,,0001-01-01T00:00:00Z
```

### Unmarshal and metadata <a name="examples_unmarshal_and_metadata"></a>

It may happen that your CSV input will not always have the same header. In addition
to your base fields you may get extra metadata that you would still like to store.
[Decoder](https://godoc.org/github.com/jszwec/csvutil#Decoder) provides 
[Unused](https://godoc.org/github.com/jszwec/csvutil#Decoder.Unused) method, which after each call to 
[Decode](https://godoc.org/github.com/jszwec/csvutil#Decoder.Decode) can report which header indexes 
were not used during decoding. Based on that, it is possible to handle and store all these extra values.

```go
	type User struct {
		Name      string            `csv:"name"`
		City      string            `csv:"city"`
		Age       int               `csv:"age"`
		OtherData map[string]string `csv:"-"`
	}

	csvReader := csv.NewReader(strings.NewReader(`
name,age,city,zip
alice,25,la,90005
bob,30,ny,10005`))

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}

	header := dec.Header()
	var users []User
	for {
		u := User{OtherData: make(map[string]string)}

		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		for _, i := range dec.Unused() {
			u.OtherData[header[i]] = dec.Record()[i]
		}
		users = append(users, u)
	}

	fmt.Println(users)

	// Output:
	// [{alice la 25 map[zip:90005]} {bob ny 30 map[zip:10005]}]
```

### But my CSV file has no header... <a name="examples_but_my_csv_has_no_header"></a>

Some CSV files have no header, but if you know how it should look like, it is
possible to define a struct and generate it. All that is left to do, is to pass
it to a decoder.

```go
	type User struct {
		ID   int
		Name string
		Age  int `csv:",omitempty"`
		City string
	}

	csvReader := csv.NewReader(strings.NewReader(`
1,John,27,la
2,Bob,,ny`))

	// in real application this should be done once in init function.
	userHeader, err := csvutil.Header(User{}, "csv")
	if err != nil {
		log.Fatal(err)
	}

	dec, err := csvutil.NewDecoder(csvReader, userHeader...)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	for {
		var u User
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}

	fmt.Printf("%+v", users)

	// Output:
	// [{ID:1 Name:John Age:27 City:la} {ID:2 Name:Bob Age:0 City:ny}]
```

### Decoder.Map - data normalization <a name="examples_decoder_map"></a>

The Decoder's [Map](https://godoc.org/github.com/jszwec/csvutil#Decoder.Map) function is a powerful tool that can help clean up or normalize
the incoming data before the actual decoding takes place.

Lets say we want to decode some floats and the csv input contains some NaN values, but these values are represented by the 'n/a' string. An attempt to decode 'n/a' into float will end up with error, because strconv.ParseFloat expects 'NaN'. Knowing that, we can implement a Map function that will normalize our 'n/a' string and turn it to 'NaN' only for float types.

```go
	dec, err := NewDecoder(r)
	if err != nil {
		log.Fatal(err)
	}

	dec.Map = func(field, column string, v interface{}) string {
		if _, ok := v.(float64); ok && field == "n/a" {
			return "NaN"
		}
		return field
	}
```

Now our float64 fields will be decoded properly into NaN. What about float32, float type aliases and other NaN formats? Look at the full example [here](https://gist.github.com/jszwec/2bb94f8f3612e0162eb16003701f727e).

### Different separator/delimiter <a name="examples_different_separator"></a>

Some files may use different value separators, for example TSV files would use `\t`. The following examples show how to set up a Decoder and Encoder for such use case.

#### Decoder:
```go
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'

	dec, err := NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	for {
		var u User
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}

```

#### Encoder:
```go
	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
        w.Comma = '\t'
	enc := csvutil.NewEncoder(w)

	for _, u := range users {
		if err := enc.Encode(u); err != nil {
			log.Fatal(err)
		}
        }

	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
```

### Custom Types and Overrides <a name="examples_custom_types"></a>

There are multiple ways to customize or override your type's behavior.

1. a type implements [csvutil.Marshaler](https://pkg.go.dev/github.com/jszwec/csvutil#Marshaler) and/or [csvutil.Unmarshaler](https://pkg.go.dev/github.com/jszwec/csvutil#Unmarshaler)
```go
type Foo int64

func (f Foo) MarshalCSV() ([]byte, error) {
	return strconv.AppendInt(nil, int64(f), 16), nil
}

func (f *Foo) UnmarshalCSV(data []byte) error {
	i, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*f = Foo(i)
	return nil
}
```
2. a type implements [encoding.TextUnmarshaler](https://golang.org/pkg/encoding/#TextUnmarshaler) and/or [encoding.TextMarshaler](https://golang.org/pkg/encoding/#TextMarshaler)
```go
type Foo int64

func (f Foo) MarshalText() ([]byte, error) {
	return strconv.AppendInt(nil, int64(f), 16), nil
}

func (f *Foo) UnmarshalText(data []byte) error {
	i, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*f = Foo(i)
	return nil
}
```
3. a type is registered using [Encoder.Register](https://pkg.go.dev/github.com/jszwec/csvutil#Encoder.Register) and/or [Decoder.Register](https://pkg.go.dev/github.com/jszwec/csvutil#Decoder.Register)
```go
type Foo int64

enc.Register(func(f Foo) ([]byte, error) {
	return strconv.AppendInt(nil, int64(f), 16), nil
})

dec.Register(func(data []byte, f *Foo) error {
	v, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*f = Foo(v)
	return nil
})
```
4. a type implements an interface that was registered using [Encoder.Register](https://pkg.go.dev/github.com/jszwec/csvutil#Encoder.Register) and/or [Decoder.Register](https://pkg.go.dev/github.com/jszwec/csvutil#Decoder.Register)
```go
type Foo int64

func (f Foo) String() string {
	return strconv.FormatInt(int64(f), 16)
}

func (f *Foo) Scan(state fmt.ScanState, verb rune) error {
	// too long; look here: https://github.com/jszwec/csvutil/blob/master/example_decoder_register_test.go#L19
}

enc.Register(func(s fmt.Stringer) ([]byte, error) {
	return []byte(s.String()), nil
})

dec.Register(func(data []byte, s fmt.Scanner) error {
	_, err := fmt.Sscan(string(data), s)
	return err
})
```

The order of precedence for both Encoder and Decoder is:
1. type is registered
2. type implements an interface that was registered
3. csvutil.{Un,M}arshaler
4. encoding.Text{Un,M}arshaler

For more examples look [here](https://pkg.go.dev/github.com/jszwec/csvutil?readme=expanded#pkg-examples)

### Custom time.Time format <a name="examples_time_format"></a>

Type [time.Time](https://golang.org/pkg/time/#Time) can be used as is in the struct fields by both Decoder and Encoder
due to the fact that both have builtin support for [encoding.TextUnmarshaler](https://golang.org/pkg/encoding/#TextUnmarshaler) and [encoding.TextMarshaler](https://golang.org/pkg/encoding/#TextMarshaler). This means that by default
Time has a specific format; look at [MarshalText](https://golang.org/pkg/time/#Time.MarshalText) and [UnmarshalText](https://golang.org/pkg/time/#Time.UnmarshalText). There are two ways to override it, which one you choose depends on your use case:

1. Via Register func (based on encoding/json)
```go
const format = "2006/01/02 15:04:05"

marshalTime := func(t time.Time) ([]byte, error) {
	return t.AppendFormat(nil, format), nil
}

unmarshalTime := func(data []byte, t *time.Time) error {
	tt, err := time.Parse(format, string(data))
	if err != nil {
		return err
	}
	*t = tt
	return nil
}

enc := csvutil.NewEncoder(w)
enc.Register(marshalTime)

dec, err := csvutil.NewDecoder(r)
if err != nil {
	return err
}
dec.Register(unmarshalTime)
```

2. With custom type:
```go
type Time struct {
	time.Time
}

const format = "2006/01/02 15:04:05"

func (t Time) MarshalCSV() ([]byte, error) {
	var b [len(format)]byte
	return t.AppendFormat(b[:0], format), nil
}

func (t *Time) UnmarshalCSV(data []byte) error {
	tt, err := time.Parse(format, string(data))
	if err != nil {
		return err
	}
	*t = Time{Time: tt}
	return nil
}
```

### Custom struct tags <a name="examples_struct_tags"></a>

Like in other Go encoding packages struct field tags can be used to set
custom names or options. By default encoders and decoders are looking at `csv` tag.
However, this can be overriden by manually setting the Tag field.

```go
	type Foo struct {
		Bar int `custom:"bar"`
	}
```

```go
	dec, err := csvutil.NewDecoder(r)
	if err != nil {
		log.Fatal(err)
	}
	dec.Tag = "custom"
```

```go
	enc := csvutil.NewEncoder(w)
	enc.Tag = "custom"
```

### Slice and Map fields <a name="examples_slice_and_map_field"></a>

There is no default encoding/decoding support for slice and map fields because there is no CSV spec for such values.
In such case, it is recommended to create a custom type alias and implement Marshaler and Unmarshaler interfaces.
Please note that slice and map aliases behave differently than aliases of other types - there is no need for type casting.

```go
	type Strings []string

	func (s Strings) MarshalCSV() ([]byte, error) {
		return []byte(strings.Join(s, ",")), nil // strings.Join takes []string but it will also accept Strings
	}

	type StringMap map[string]string

	func (sm StringMap) MarshalCSV() ([]byte, error) {
		return []byte(fmt.Sprint(sm)), nil
	}

	func main() {
		b, err := csvutil.Marshal([]struct {
			Strings Strings   `csv:"strings"`
			Map     StringMap `csv:"map"`
		}{
			{[]string{"a", "b"}, map[string]string{"a": "1"}}, // no type casting is required for slice and map aliases
			{Strings{"c", "d"}, StringMap{"b": "1"}},
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", b)

		// Output:
		// strings,map
		// "a,b",map[a:1]
		// "c,d",map[b:1]
	}
```

### Nested/Embedded structs <a name="examples_nested_structs"></a>

Both Encoder and Decoder support nested or embedded structs.

Playground: https://play.golang.org/p/ZySjdVkovbf

```go
package main

import (
	"fmt"

	"github.com/jszwec/csvutil"
)

type Address struct {
	Street string `csv:"street"`
	City   string `csv:"city"`
}

type User struct {
	Name string `csv:"name"`
	Address
}

func main() {
	users := []User{
		{
			Name: "John",
			Address: Address{
				Street: "Boylston",
				City:   "Boston",
			},
		},
	}

	b, err := csvutil.Marshal(users)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", b)

	var out []User
	if err := csvutil.Unmarshal(b, &out); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", out)

	// Output:
	//
	// name,street,city
	// John,Boylston,Boston
	//
	// [{Name:John Address:{Street:Boylston City:Boston}}]
}
```

### Inline tag <a name="examples_inlined_structs"></a>

Fields with inline tag behave similarly to embedded struct fields. However,
it gives a possibility to specify the prefix for all underlying fields. This
can be useful when one structure can define multiple CSV columns because they 
are different from each other only by a certain prefix. Look at the example below.

Playground: https://play.golang.org/p/jyEzeskSnj7

```go
package main

import (
	"fmt"

	"github.com/jszwec/csvutil"
)

func main() {
	type Address struct {
		Street string `csv:"street"`
		City   string `csv:"city"`
	}

	type User struct {
		Name        string  `csv:"name"`
		Address     Address `csv:",inline"`
		HomeAddress Address `csv:"home_address_,inline"`
		WorkAddress Address `csv:"work_address_,inline"`
		Age         int     `csv:"age,omitempty"`
	}

	users := []User{
		{
			Name:        "John",
			Address:     Address{"Washington", "Boston"},
			HomeAddress: Address{"Boylston", "Boston"},
			WorkAddress: Address{"River St", "Cambridge"},
			Age:         26,
		},
	}

	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Printf("%s\n", b)

	// Output:
	// name,street,city,home_address_street,home_address_city,work_address_street,work_address_city,age
	// John,Washington,Boston,Boylston,Boston,River St,Cambridge,26
}
```

Performance
------------

csvutil provides the best encoding and decoding performance with small memory usage.

### Unmarshal <a name="performance_unmarshal"></a>

[benchmark code](https://gist.github.com/jszwec/e8515e741190454fa3494bcd3e1f100f)

#### csvutil:
```
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-12         	  280696	      4516 ns/op	    7332 B/op	      26 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-12       	   95750	     11517 ns/op	    8356 B/op	      35 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-12      	   14997	     83146 ns/op	   18532 B/op	     125 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-12     	    1485	    750143 ns/op	  121094 B/op	    1025 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-12    	     154	   7587205 ns/op	 1136662 B/op	   10025 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-12   	      14	  76126616 ns/op	11808744 B/op	  100025 allocs/op
```

#### gocsv:
```
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-12           	  141330	      7499 ns/op	    7795 B/op	      97 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_records-12         	   54252	     21664 ns/op	   13891 B/op	     307 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-12        	    6920	    159662 ns/op	   72644 B/op	    2380 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-12       	     752	   1556083 ns/op	  650248 B/op	   23083 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-12      	      72	  17086623 ns/op	 7017469 B/op	  230092 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-12     	       7	 163610749 ns/op	75004923 B/op	 2300105 allocs/op
```

#### easycsv:
```
BenchmarkUnmarshal/easycsv.ReadAll/1_record-12           	  101527	     10662 ns/op	    8855 B/op	      81 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10_records-12         	   23325	     51437 ns/op	   24072 B/op	     391 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100_records-12        	    2402	    447296 ns/op	  170538 B/op	    3454 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-12       	     272	   4370854 ns/op	 1595683 B/op	   34057 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-12      	      24	  47502457 ns/op	18861808 B/op	  340068 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-12     	       3	 468974170 ns/op	189427066 B/op	 3400082 allocs/op
```

### Marshal <a name="performance_marshal"></a>

[benchmark code](https://gist.github.com/jszwec/31980321e1852ebb5615a44ccf374f17)

#### csvutil:
```
BenchmarkMarshal/csvutil.Marshal/1_record-12         	  279558	      4390 ns/op	    9952 B/op	      12 allocs/op
BenchmarkMarshal/csvutil.Marshal/10_records-12       	   82478	     15608 ns/op	   10800 B/op	      21 allocs/op
BenchmarkMarshal/csvutil.Marshal/100_records-12      	   10275	    117288 ns/op	   28208 B/op	     112 allocs/op
BenchmarkMarshal/csvutil.Marshal/1000_records-12     	    1075	   1147473 ns/op	  168508 B/op	    1014 allocs/op
BenchmarkMarshal/csvutil.Marshal/10000_records-12    	     100	  11985382 ns/op	 1525973 B/op	   10017 allocs/op
BenchmarkMarshal/csvutil.Marshal/100000_records-12   	       9	 113640813 ns/op	22455873 B/op	  100021 allocs/op
```

#### gocsv:
```
BenchmarkMarshal/gocsv.Marshal/1_record-12           	  203052	      6077 ns/op	    5914 B/op	      81 allocs/op
BenchmarkMarshal/gocsv.Marshal/10_records-12         	   50132	     24585 ns/op	    9284 B/op	     360 allocs/op
BenchmarkMarshal/gocsv.Marshal/100_records-12        	    5480	    212008 ns/op	   51916 B/op	    3151 allocs/op
BenchmarkMarshal/gocsv.Marshal/1000_records-12       	     514	   2053919 ns/op	  444506 B/op	   31053 allocs/op
BenchmarkMarshal/gocsv.Marshal/10000_records-12      	      52	  21066666 ns/op	 4332377 B/op	  310064 allocs/op
BenchmarkMarshal/gocsv.Marshal/100000_records-12     	       5	 207408929 ns/op	51169419 B/op	 3100077 allocs/op
```
