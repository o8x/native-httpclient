package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "net"
    "regexp"
    "strconv"
    "strings"
)

type Data struct {
    Method  string
    Route   string
    Host    string
    Length  int
    Headers string
    Body    string
}

type HttpClient struct {
    Network string
    Address string
    Headers map[string]string
    Cookies map[string]string
    Request struct {
        Origin  string
        Payload string
    }

    Response struct {
        ContentType   string
        ContentLength uint
        StatusCode    uint
        Headers       map[string]string
        Cookies       map[string]string
        Body          string
    }
}

var baseProtocol = `{{ .Method }} {{ .Route }} HTTP/1.1
Host: {{ .Host }}
Connection: close
`

func NewUnixSock(file string) *HttpClient {
    return &HttpClient{
        Network: "unix",
        Address: file,
        Headers: map[string]string{},
        Cookies: map[string]string{},
    }
}

func NewTcp(link string) *HttpClient {
    // TODO 443 TLS
    return &HttpClient{
        Network: "tcp",
        Address: link + ":80",
        Headers: map[string]string{},
        Cookies: map[string]string{},
    }
}

func (h *HttpClient) Get(route string, params map[string]string) (string, error) {
    if len(params) > 0 {
        var queryString []string
        for param := range params {
            queryString = append(queryString, fmt.Sprintf("%s=%s", param, params[param]))
        }
        route = fmt.Sprintf("%s?%s", route, strings.Join(queryString, "&"))
    }

    return h.Do("GET", route, nil)
}

func (h *HttpClient) Post(url string, body interface{}) (string, error) {
    return h.Do("POST", url, body)
}

func (h *HttpClient) Delete(url string, body interface{}) (string, error) {
    return h.Do("DELETE", url, body)
}

func (h *HttpClient) Put(url string, body interface{}) (string, error) {
    return h.Do("PUT", url, body)
}

func (h *HttpClient) Patch(url string, body interface{}) (string, error) {
    return h.Do("PATCH", url, body)
}

func (h *HttpClient) Head(url string, body interface{}) (string, error) {
    return h.Do("HEAD", url, body)
}

func (h *HttpClient) Options(url string, body interface{}) (string, error) {
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

func (h *HttpClient) MakeProtocol(data interface{}, method, route string) string {
    switch t := data.(type) {
    case string:
        h.Request.Payload = t
    default:
        payload, _ := json.Marshal(data)
        h.Request.Payload = string(payload)
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
    }

    // 如果有 Header 则追加一个 \r\n
    var headersBuffer bytes.Buffer
    if len(h.Headers) > 0 {
        for key := range h.Headers {
            if key == "" {
                continue
            }
            headersBuffer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, h.Headers[key])))
        }
    }

    var host = "localhost"
    if h.Network == "tcp" {
        host = h.Address
    }

    h.Request.Origin = StringTemplate(h.MakeHttpProtocol(method), &Data{
        method,
        route,
        host,
        len(h.Request.Payload),
        headersBuffer.String(),
        h.Request.Payload,
    })

    return strings.ReplaceAll(h.Request.Origin, "\n", "\r\n")
}

func (h *HttpClient) MakeHttpProtocol(method string) string {
    // TODO 自定义 Content-Type
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
Content-Type: application/json;charset=UTF-8
{{ .Headers }}
{{ .Body | Unescape }}`
    }

    return baseProtocol + "\n"
}

func (h *HttpClient) Do(method string, route string, body interface{}) (string,
    error) {

    conn, err := net.Dial(h.Network, h.Address)
    conn.Write([]byte(h.MakeProtocol(body, method, route)))
    defer conn.Close()

    if err != nil {
        return "", err
    }

    // TODO 302 跟随
    return h.ParseResponse(conn), nil
}

func (h *HttpClient) ParseResponse(conn net.Conn) string {
    reader := bufio.NewReader(conn)
    var body bytes.Buffer
    for {
        line, _, err := reader.ReadLine()
        body.Write(line)
        body.Write([]byte("\r\n"))

        if err != nil {
            break
        }
    }

    response := body.String()
    dataStartIndex := strings.Index(response, "\r\n\r\n")

    h.Response.Body = response[dataStartIndex+4 : len(response)-4]
    h.Response.Headers = ParseHeaderString(response[0:dataStartIndex])
    h.Response.Cookies = ParseCookieString(h.Response.Headers["Set-Cookie"])
    h.Response.StatusCode = ParseStatusCode(response[0:dataStartIndex])
    h.Response.ContentType = h.Response.Headers["Content-Type"]
    length, _ := strconv.Atoi(h.Response.Headers["Content-Length"])
    h.Response.ContentLength = uint(length)

    return h.Response.Body
}

func ParseHeaderString(headers string) map[string]string {
    result := map[string]string{}
    if headers == "" {
        return result
    }

    // 解析 Header
    headersList := strings.Split(headers, "\r\n")
    for ind := range headersList {
        header := headersList[ind]
        index := strings.Index(header, ": ")
        if index == -1 {
            continue
        }

        result[header[0:index]] = header[index+2 : len(header)]
    }

    return result
}

func ParseCookieString(cookies string) map[string]string {
    result := map[string]string{}
    if cookies == "" {
        return result
    }

    cookieList := strings.Split(cookies, "; ")
    for ind := range cookieList {
        cookie := cookieList[ind]
        index := strings.Index(cookie, "=")
        if index == -1 {
            continue
        }

        result[cookie[0:index]] = cookie[index+1 : len(cookie)]
    }

    return result
}

func ParseStatusCode(headers string) uint {
    if headers == "" {
        return 500
    }

    protoHeader := strings.Split(headers, "\r\n")[0]
    reg := regexp.MustCompile(`\d{3}`)
    code := reg.FindAllString(protoHeader, -1)[0]

    uintCode, e := strconv.Atoi(code)
    if e != nil {
        return 500
    }
    return uint(uintCode)
}
