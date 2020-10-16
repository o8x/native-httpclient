package nativehttpclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Response struct {
	ContentType   string
	ContentLength uint
	StatusCode    uint
	Headers       map[string]string
	Cookies       map[string]string
	Body          string
	Origin        string
}

func (r *Response) BodyFormat(declare interface{}) error {
	if r.Body == "" {
		return errors.New("response empty")
	}

	err := json.Unmarshal([]byte(r.Body), &declare)
	if err != nil {
		return err
	}

	return nil
}

func (r *Response) responseHandler(conn net.Conn) *Response {
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

	r.Origin = body.String()
	log.Info("response protocol: ", body.String())
	dataStartIndex := strings.Index(r.Origin, "\r\n\r\n")

	// 避免数据为空
	if dataStartIndex+4 <= len(r.Origin)-4 {
		r.Body = r.Origin[dataStartIndex+4 : len(r.Origin)-4]
	}

	r.Headers = parseHeaderString(r.Origin[0:dataStartIndex])
	r.Cookies = parseCookieString(r.Headers["Set-Cookie"])
	r.StatusCode = parseStatusCode(r.Origin[0:dataStartIndex])
	r.ContentType = r.Headers["Content-Type"]
	length, _ := strconv.Atoi(r.Headers["Content-Length"])
	r.ContentLength = uint(length)

	return r
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

		result[header[0:index]] = header[index+2:]
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

		result[cookie[0:index]] = cookie[index+1:]
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
