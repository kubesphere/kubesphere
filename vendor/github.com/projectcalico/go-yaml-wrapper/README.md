[![Build Status](https://semaphoreci.com/api/v1/calico/go-yaml-wrapper/branches/calico/badge.svg)](https://semaphoreci.com/calico/go-yaml-wrapper)

# YAML marshaling and unmarshaling support for Go

This is a fork of `github.com/ghodss/yaml` to provide modified YAML and JSON
parsing for use with libcalico-go and calicoctl.  The modifications include:
  -  Swapping support of Float32 with Float64 (since calico does not use Float32)
  -  Providing the ability to perform strict unmarshaling (i.e. erroring if a 
     field in the document was not in the struct)
  -  Slightly modified error messages to be less go-lang oriented and more
     user facing.
  -  Indication of field names or values

