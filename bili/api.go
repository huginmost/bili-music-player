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

// GetListPlayInfo is the exported wrapper for bili_mget_pi.
func (b *Bili) GetListPlayInfo(listID, outputPath string) (string, error) {
	return b.bili_mget_pi(listID, outputPath)
}

// GetListInitialState is the exported wrapper for bili_mget_is.
func (b *Bili) GetListInitialState(listID, outputPath string) (string, error) {
	return b.bili_mget_is(listID, outputPath)
}

// ParseJSON is the exported wrapper for bili_js.
func (b *Bili) ParseJSON(inputPath string) (map[string]any, string, error) {
	return b.bili_js(inputPath)
}

// GetNestedString is the Bili wrapper around the package helper.
func (b *Bili) GetNestedString(data map[string]any, keys ...string) (string, bool) {
	return GetNestedString(data, keys...)
}

// DeleteTitle is the exported wrapper for bili_del.
func (b *Bili) DeleteTitle(title string) error {
	return b.DeletePlaylist(title)
}
