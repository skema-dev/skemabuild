package http

import (
	"github.com/go-resty/resty/v2"
	"github.com/skema-dev/skemabuild/internal/pkg/console"
)

func GetTextContent(url string) string {
	client := resty.New()
	resp, err := client.R().
		Get(url)
	console.FatalIfError(err, "failed fetching from "+url)
	content := string(resp.Body())
	return content
}
