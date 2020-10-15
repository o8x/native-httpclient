package nativehttpclient

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

type Response struct {
    ContentType   string
    ContentLength uint
    StatusCode    uint
    Headers       map[string]string
    Cookies       map[string]string
    Body          string
}

type Request struct {
    Origin  string
    Payload string
}

type HttpClient struct {
    Network  string
    Address  string
    Headers  map[string]string
    Cookies  map[string]string
    Request  Request
    Response Response
}

var baseProtocol = `{{ .Method }} {{ .Route }} HTTP/1.1
Host: {{ .Host }}
Connection: close
`

func (h *HttpClient) makeHttpProtocol(data interface{}, method, route string) string {
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

    h.Request.Origin = StringTemplate(h.restfulHandler(method), &Data{
        method,
        route,
        host,
        len(h.Request.Payload),
        headersBuffer.String(),
        h.Request.Payload,
    })

    return strings.ReplaceAll(h.Request.Origin, "\n", "\r\n")
}

func (h *HttpClient) restfulHandler(method string) string {
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

func (h *HttpClient) Do(method string, route string, body interface{}) (Response, error) {
    conn, err := net.Dial(h.Network, h.Address)
    defer conn.Close()

    if err != nil {
        return Response{}, err
    }

    if _, sendErr := conn.Write([]byte(h.makeHttpProtocol(body, method, route))); sendErr != nil {
        return Response{}, err
    }

    // TODO 302 跟随
    return h.parseResponse(conn), nil
}

func (h *HttpClient) parseResponse(conn net.Conn) Response {
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
    h.Response.Headers = parseHeaderString(response[0:dataStartIndex])
    h.Response.Cookies = parseCookieString(h.Response.Headers["Set-Cookie"])
    h.Response.StatusCode = parseStatusCode(response[0:dataStartIndex])
    h.Response.ContentType = h.Response.Headers["Content-Type"]
    length, _ := strconv.Atoi(h.Response.Headers["Content-Length"])
    h.Response.ContentLength = uint(length)

    return h.Response
}

func parseHeaderString(headers string) map[string]string {
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

func parseCookieString(cookies string) map[string]string {
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

func parseStatusCode(headers string) uint {
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
