# changes to the go-restful-openapi package

# v2+ versions are using the Go module of go-restful v3+

## v1.4.0 + v2.2.0
    - Allow maps as top level types and support maps to slices (#63)

## v1.3.0 + v2.1.0
    - add json.Number handling (PR #61)
    - add type alias support for primitives (PR #61)

## v1.2.0

    - handle map[string][]byte (#59)

## v1.1.0 (v0.14.1)

    - Add Host field to Config which is copied into Swagger object
    - Enable CORS by default as per the documentation (#58)
    - add go module
    - update dependencies

## v0.13.0

    - Do not use 200 as default response, instead use the one explicitly defined.
    - support time.Duration
    - Fix Parameter 'AllowableValues' to populate swagger definition

## v0.12.0

    - add support for time.Duration
    - Populate the swagger definition with the parameter's 'AllowableValues' as an enum (#53)
    - Fix for #19 MapModelTypeNameFunc has incomplete behavior
    - Merge paths with existing paths from other webServices (#48)
    - prevent array param.Type be overwritten in the else case below (#47)

## v0.11.0

    - Register pointer to array/slice of primitives as such rather than as reference to the primitive type definition. (#46)
    - Add support for map types using "additional properties" (#44) 

## <= v0.10.0

See `git log`.