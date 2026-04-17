package bili

// GetNestedString walks nested JSON objects represented as map[string]any and returns a string value.
func GetNestedString(data map[string]any, keys ...string) (string, bool) {
	var current any = data

	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return "", false
		}

		next, ok := obj[key]
		if !ok {
			return "", false
		}

		current = next
	}

	value, ok := current.(string)
	return value, ok
}
