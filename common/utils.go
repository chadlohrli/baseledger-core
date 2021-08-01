package common

// PanicIfEmpty panics if the given string is empty
func PanicIfEmpty(val string, msg string) {
	if val == "" {
		panic(msg)
	}
}

// StringOrNil returns the given string or nil when empty
func StringOrNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}
