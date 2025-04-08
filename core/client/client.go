package client

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	maxBodySize    = 10 * 1024 * 1024 // 10MB
)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (h *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}

func (h *HttpClient) Get(url string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

func (h *HttpClient) Post(url string, body []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}
