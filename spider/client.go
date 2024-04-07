package spider

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	*http.Client
}

func newClient(proxy []string) *Client {
	client := &Client{}
	if len(proxy) > 0 {
		rand.Seed(time.Now().UnixNano())
		// 随机获取切片中的元素
		randIndex := rand.Intn(len(proxy))
		u, _ := url.Parse(proxy[randIndex])
		transport := &http.Transport{
			Proxy: http.ProxyURL(u),
		}

		client = &Client{
			Client: &http.Client{
				Transport: transport,
			}}
	}

	return client
}

// GetRequest 获取HTTP请求
func (c *Client) GetRequest(method, rawUrl string, header map[string]string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, rawUrl, data)
	if err != nil {
		return nil, errors.WithMessagef(err, "new http request err, method = %v, rawurl = %v, header = %v", method, rawUrl, header)
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}

	return req, nil
}

// GetResponse 获取HTTP响应及其响应体内容
func (c *Client) GetResponse(req *http.Request) (*http.Response, []byte, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("get response err", err)
		}
	}()
	response, err := c.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "http client do request err")
	}

	if response == nil || response.Body == nil {
		return nil, nil, errors.New("http client do request get nil")
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "read all response body err")
	}

	return response, body, nil
}
