package main

import (
	"fmt"
	"net/http"
)

func request(url string) (*http.Response, error) {
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	var i int

	for i < 50 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "hltvorgtgbot")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case 200:
			return resp, nil
		case 302:
			newURL, err := resp.Location()
			if err != nil {
				return nil, fmt.Errorf("status code %d (%s) but get error in location(): %s", resp.StatusCode, resp.Status, err)
			}
			url = newURL.Scheme + "://" + newURL.Host + newURL.Path
		default:
			return nil, fmt.Errorf("unknown StatusCode %d %s, %s", resp.StatusCode, resp.Status, url)
		}
	}

	return nil, fmt.Errorf("limit request (%d)", 50)
}
