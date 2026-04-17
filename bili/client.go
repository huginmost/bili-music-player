package bili

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const baseURL = "https://www.bilibili.com/"

// Bili holds the HTTP client and cookie state used for Bilibili requests.
type Bili struct {
	client *http.Client
	cookie string
}

// New creates a Bili instance and stores the provided cookie.
func New(cookie string) (*Bili, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	b := &Bili{
		client: &http.Client{
			Jar:     jar,
			Timeout: 20 * time.Second,
		},
		cookie: cookie,
	}

	if err := b.bili_init(cookie); err != nil {
		return nil, err
	}

	return b, nil
}

// bili_init stores the cookie in the HTTP jar for later requests.
func (b *Bili) bili_init(cookie string) error {
	if b == nil {
		return errors.New("bili is nil")
	}

	b.cookie = cookie
	if cookie == "" {
		return nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	b.client.Jar.SetCookies(u, []*http.Cookie{
		{
			Name:  "SESSDATA",
			Value: cookie,
			Path:  "/",
		},
	})

	return nil
}

// bili_try checks whether the Bilibili homepage can be reached successfully.
func (b *Bili) bili_try() bool {
	if b == nil {
		return false
	}

	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := b.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
