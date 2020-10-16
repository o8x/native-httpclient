package nativehttpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

var UserAgent = "Native_Http_Client/0.5.0"

type Data struct {
	Method  string
	Route   string
	Host    string
	Length  int
	Headers string
	Body    string
}

type Request struct {
	Origin  string
	Payload string
}

type Configs struct {
	MaxRedirects uint
}

type HttpClient struct {
	Network  string
	Address  string
	Headers  map[string]string
	Cookies  map[string]string
	Request  Request
	Response Response
	Configs  Configs
}

var baseProtocol = `{{ .Method }} {{ .Route }} HTTP/1.1
Host: {{ .Host }}
Connection: close
`

func init() {
	InitLog()
}

func (h *HttpClient) makeHttpProtocol(data interface{}, method, route string) string {
	switch value := data.(type) {
	case nil:
		h.Request.Payload = ""
	case string:
		h.Request.Payload = value
	default:
		payload, _ := json.Marshal(data)
		h.Request.Payload = string(payload)
	}

	log.Info("request payload: ", h.Request.Payload)

	// 自定义数据类型
	if _, found := h.Headers["Content-Type"]; !found {
		h.WithHeader("Content-Type", "application/json;charset=UTF-8")
	}

	// 自定义数据类型
	if _, found := h.Headers["User-Agent"]; !found {
		h.WithHeader("User-Agent", UserAgent)
	}

	// 如果没有在 Header 中定义 Cookie，并且自定义 Cookie 不为空
	if _, found := h.Headers["Cookie"]; !found && len(h.Cookies) > 0 {
		h.Headers["Cookie"] = ""
		for key := range h.Cookies {
			if key == "" {
				continue
			}
			h.Headers["Cookie"] += fmt.Sprintf("%s=%s; ", key, h.Cookies[key])
		}

		log.Info("request cookies: ", h.Headers["Cookie"])
	}

	// 如果有 Header 则追加一个 \r\n
	var headersBuffer bytes.Buffer
	if len(h.Headers) > 0 {
		for key := range h.Headers {
			if key == "" {
				continue
			}
			headersBuffer.Write([]byte(fmt.Sprintf("%s: %s\n", key, h.Headers[key])))
		}

		log.Info("request headers: ", headersBuffer.String())
	}

	var host = "localhost"
	if h.Network == "tcp" {
		host = h.Address
	}

	h.Request.Origin = StringTemplate(h.restfulHandler(method), &Data{
		method,
		route,
		host,
		len(h.Request.Payload),
		headersBuffer.String(),
		h.Request.Payload,
	})

	return strings.ReplaceAll(
		strings.ReplaceAll(h.Request.Origin, "\r", ""), "\n", "\r\n",
	)
}

func (h *HttpClient) restfulHandler(method string) string {
	// TODO HTTP/2
	// TODO gzip
	switch method {
	case "PATCH":
		fallthrough
	case "PUT":
		fallthrough
	case "POST":
		return baseProtocol + `Content-Length: {{ .Length }}
Accept: */*
{{ .Headers }}
{{ .Body | Unescape }}`
	}

	return baseProtocol + "\n"
}

func (h *HttpClient) Do(method string, route string, body interface{}) (*Response, error) {
	log.Info(fmt.Sprintf("request detail, method: %s , route: %s , body: %v", method, route, body))
	if h.Configs.MaxRedirects > 30 {
		log.Info("too many redirects")
		return &Response{}, errors.New("too many redirects")
	}

	conn, err := net.Dial(h.Network, h.Address)
	if err != nil {
		log.Info("ConnectError: ", err.Error())
		return &Response{}, err
	}

	defer conn.Close()

	data := h.makeHttpProtocol(body, method, route)
	jsonData, _ := json.Marshal(data)

	log.Info("request http protocol: ", string(jsonData))
	if _, sendErr := conn.Write([]byte(data)); sendErr != nil {
		log.Info("WriteError: ", sendErr.Error())
		return &Response{}, err
	}

	// 实现跳转跟随
	response := h.Response.responseHandler(conn)
	if response.StatusCode == 302 || response.StatusCode == 301 {
		h.Configs.MaxRedirects++
		h.Address = strings.TrimSpace(response.Headers["Location"])
		log.Info("redirect to: ", h.Address)
		return h.Do(method, route, body)
	}

	return response, nil
}
