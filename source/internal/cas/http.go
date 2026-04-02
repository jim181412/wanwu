package cas

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var defaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

func httpGet(target string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected GET status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func httpPost(target string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, target, nil)
	if err != nil {
		return "", err
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected POST status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
