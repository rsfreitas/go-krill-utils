package logger

// Attribute is a helper object that implements the loggerApi.Attribute interface
// allowing services to add more information into their log messages.
type Attribute struct {
	key   string
	value interface{}
}

// String wraps a string into a formatted log string field.
func String(key, value string) Attribute {
	return Attribute{
		key:   key,
		value: value,
	}
}

// Int32 wraps an int32 value into a formatted log string field.
func Int32(key string, value int32) Attribute {
	return Attribute{
		key:   key,
		value: value,
	}
}

// Any wraps a value into a formatted log string field.
func Any(key string, value interface{}) Attribute {
	return Attribute{
		key:   key,
		value: value,
	}
}

// Error wraps an error into a formatted log string field.
func Error(err error) Attribute {
	return Attribute{
		key:   "error.message",
		value: err.Error(),
	}
}

func (f Attribute) Key() string {
	return f.key
}

func (f Attribute) Value() interface{} {
	return f.value
}
