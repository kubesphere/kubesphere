package multiconfig

type multiLoader []Loader

// MultiLoader creates a loader that executes the loaders one by one in order
// and returns on the first error.
func MultiLoader(loader ...Loader) Loader {
	return multiLoader(loader)
}

// Load loads the source into the config defined by struct s
func (m multiLoader) Load(s interface{}) error {
	for _, loader := range m {
		if err := loader.Load(s); err != nil {
			return err
		}
	}

	return nil
}

// MustLoad loads the source into the struct, it panics if gets any error
func (m multiLoader) MustLoad(s interface{}) {
	if err := m.Load(s); err != nil {
		panic(err)
	}
}
