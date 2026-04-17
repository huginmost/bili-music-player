package bili

// BiliInit is the exported wrapper for the requested bili_init behavior.
func BiliInit(cookie string) (*Bili, error) {
	return New(cookie)
}

// Try is the exported wrapper for bili_try.
func (b *Bili) Try() bool {
	return b.bili_try()
}

// GetPlayInfo is the exported wrapper for bili_get_pi.
func (b *Bili) GetPlayInfo(bvid, outputPath string) (string, error) {
	return b.bili_get_pi(bvid, outputPath)
}

// GetInitialState is the exported wrapper for bili_get_is.
func (b *Bili) GetInitialState(bvid, outputPath string) (string, error) {
	return b.bili_get_is(bvid, outputPath)
}

// ParseJSON is the exported wrapper for bili_js.
func (b *Bili) ParseJSON(inputPath string) (map[string]any, string, error) {
	return b.bili_js(inputPath)
}

// GetNestedString is the Bili wrapper around the package helper.
func (b *Bili) GetNestedString(data map[string]any, keys ...string) (string, bool) {
	return GetNestedString(data, keys...)
}
