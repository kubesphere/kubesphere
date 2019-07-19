# changes to the go-restful-openapi package

## v1.0.0

- Fix for #19 MapModelTypeNameFunc has incomplete behavior
- prevent array param.Type be overwritten in the else case below (#47)
- Merge paths with existing paths from other webServices (#48)

## v0.11.0

    - Register pointer to array/slice of primitives as such rather than as reference to the primitive type definition. (#46)
    - Add support for map types using "additional properties" (#44) 

## <= v0.10.0

See `git log`.