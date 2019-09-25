package multiconfig

type multiValidator []Validator

// MultiValidator accepts variadic validators and satisfies Validator interface.
func MultiValidator(validators ...Validator) Validator {
	return multiValidator(validators)
}

// Validate tries to validate given struct with all the validators. If it doesn't
// have any Validator it will simply skip the validation step. If any of the
// given validators return err, it will stop validating and return it.
func (d multiValidator) Validate(s interface{}) error {
	for _, validator := range d {
		if err := validator.Validate(s); err != nil {
			return err
		}
	}

	return nil
}

// MustValidate validates the struct, it panics if gets any error
func (d multiValidator) MustValidate(s interface{}) {
	if err := d.Validate(s); err != nil {
		panic(err)
	}
}
