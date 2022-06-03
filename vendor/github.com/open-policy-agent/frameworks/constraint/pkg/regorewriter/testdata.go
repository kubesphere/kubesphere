package regorewriter

// TestData represents a json or yaml data file used in unit tests.
type TestData struct {
	FilePath
	// content is the file contents.
	content []byte
}

// Content implements sourceFile
func (t *TestData) Content() ([]byte, error) {
	return t.content, nil
}
