package bili

import "os"

// GetListTitle reads is.json and returns mediaListInfo.title.
// It returns an empty string when the title cannot be found.
func (b *Bili) GetListTitle() string {
	jsInfo, _, err := b.ParseJSON(InitialStatePath)
	if err == nil {
		if title, ok := b.GetNestedString(jsInfo, "mediaListInfo", "title"); ok {
			return title
		}
	}

	raw, readErr := os.ReadFile(InitialStatePath)
	if readErr != nil {
		return ""
	}

	listInfo, err := extractObjectForKey(string(raw), "mediaListInfo")
	if err != nil {
		return ""
	}

	title, err := extractStringForKey(listInfo, "title")
	if err != nil {
		return ""
	}

	return title
}
