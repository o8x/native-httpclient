package nativehttpclient

import (
	"fmt"
	"strings"
)

var defaultClient = &HttpClient{
	Headers: map[string]string{},
	Cookies: map[string]string{},
}

func Defaults() *HttpClient {
	return defaultClient
}

func NewUnixSock(file string) *HttpClient {
	httpClient := Defaults()
	httpClient.Network = "unix"
	httpClient.Address = file
	return defaultClient
}

func NewTcp(link string) *HttpClient {
	// TODO 443 TLS
	httpClient := Defaults()
	httpClient.Network = "tcp"
	httpClient.Address = link + ":80"
	return httpClient
}

func (h *HttpClient) Get(route string, params map[string]string) (Response, error) {
	if len(params) > 0 {
		var queryString []string
		for param := range params {
			queryString = append(queryString, fmt.Sprintf("%s=%s", param, params[param]))
		}
		route = fmt.Sprintf("%s?%s", route, strings.Join(queryString, "&"))
	}

	return h.Do("GET", route, nil)
}

func (h *HttpClient) Post(url string, body interface{}) (Response, error) {
	return h.Do("POST", url, body)
}

func (h *HttpClient) Delete(url string, body interface{}) (Response, error) {
	return h.Do("DELETE", url, body)
}

func (h *HttpClient) Put(url string, body interface{}) (Response, error) {
	return h.Do("PUT", url, body)
}

func (h *HttpClient) Patch(url string, body interface{}) (Response, error) {
	return h.Do("PATCH", url, body)
}

func (h *HttpClient) Head(url string, body interface{}) (Response, error) {
	return h.Do("HEAD", url, body)
}

func (h *HttpClient) Options(url string, body interface{}) (Response, error) {
	return h.Do("OPTIONS", url, body)
}

func (h *HttpClient) WithHeader(key string, value string) *HttpClient {
	h.Headers[key] = value
	return h
}

func (h *HttpClient) WithCookie(key string, value string) *HttpClient {
	h.Cookies[key] = value
	return h
}
