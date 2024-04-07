package spider

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

func directGet(domain, path string, proxy []string) ([]byte, error) {
	uri := domain
	if len(path) > 0 {
		uri = fmt.Sprintf("%s%s", domain, path)
	}

	client := newClient(proxy)
	head := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
		"Origin":     domain,
		"Referer":    fmt.Sprintf("%s/", domain),
	}
	req, err := client.GetRequest(http.MethodGet, uri, head, nil)
	if err != nil {
		return nil, err
	}

	resp, body, err := client.GetResponse(req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.Errorf("crawlDirect get nil")
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("TxTransaction status code: %d, error: %s", resp.StatusCode, resp.Status)
	}

	return body, err
}
