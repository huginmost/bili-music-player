package bili

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

func audioURLExpired(audioURL string, now time.Time) bool {
	if strings.TrimSpace(audioURL) == "" {
		return true
	}

	parsedURL, err := url.Parse(strings.ReplaceAll(audioURL, "\\u0026", "&"))
	if err != nil {
		return true
	}

	deadline := parsedURL.Query().Get("deadline")
	if deadline == "" {
		return true
	}

	deadlineUnix, err := strconv.ParseInt(deadline, 10, 64)
	if err != nil {
		return true
	}

	return now.Unix() >= deadlineUnix
}
