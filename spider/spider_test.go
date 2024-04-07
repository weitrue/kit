package spider

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/stretchr/testify/assert"
)

func Test_directGet(t *testing.T) {
	type args struct {
		domain string
		path   string
		proxy  []string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "",
			args: args{
				domain: "",
				path:   "",
				proxy:  []string{"127.0.0.1:7890"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := directGet(tt.args.domain, tt.args.path, tt.args.proxy)
			assert.Nil(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("directGet() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	startUrl := "https://etherscan.io/tx/0x55f80f859b27a5429d29bf1306cfff9cf7376524ce2867c566b461578df3d9fc"

	c := colly.NewCollector(
		colly.AllowedDomains("etherscan.io"),
		colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)

	c.SetProxy("http://127.0.0.1:7890")
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Origin", "https://etherscan.io")
		r.Headers.Add("Referer", "https://etherscan.io/")
	})

	c.OnError(func(r *colly.Response, err error) {
		ex, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		fmt.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
		file, err := os.OpenFile(fmt.Sprintf("%s/spider/error.html", filepath.Dir(ex)), os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Log(err)
			return
		}

		defer file.Close()

		_, err = file.WriteString(string(r.Body))
		if err != nil {
			return
		}

	})

	err := c.Visit(startUrl)
	if err != nil {
		panic(fmt.Sprintf("Unable to visit %s", startUrl))
	}

	c.Wait()
}
