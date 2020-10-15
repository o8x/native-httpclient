NATIVE-HTTP-CLIENT
======

通过 tcp 和 unix sock 文件发送原生HTTP请求的HTTP客户端

### 开发计划
- [x] HTTP
- [ ] 公开 test api
- [ ] 302
- [ ] HTTPS
- [ ] TLS
- [ ] HTTP2

### 获取实例

连接到 TCP

```golang
client := NewTcp("domain")
```

连接到 Socket 文件
```golang
client := NewUnixSock("/var/run/docker.sock")
```

### 请求 

GET、DELETE、HEAD、OPTIONS

第二个参数将会被处理成 QueryString，追加到url后面。例如：/containers/json?sort=id

```golang
response, err := client.Get("/containers/json", map[string]string{
     "sort": "id",
 })
response, err := client.Delete("/containers/json", nil)
response, err := client.Head("/containers/json", nil)
response, err := client.Options("/containers/json", nil)
```

POST、PUT、PATCH

同时接受字符串和其它参数类型(将会被格式化为json)

```golang
response, err := client.Post("/containers/create", map[string]string{
    "Image": "nginx",
})
response, err := client.Put("/containers/create", `{"Image":"nginx"}`)
response, err :=  client.Patch("/containers/create", `{"Image":"nginx"}`)
```

### 原始 HTTP 协议

调用请求API，将会生成类似如下两种的原始 HTTP 协议进行发送，其中的GET和POST将会被替换成实际协议。


GET、DELETE、HEAD、OPTIONS

```ini
GET /containers/json HTTP/1.1
Host: domain:80
Connection: close
```

POST、PUT、PATCH
```ini 
POST /containers/create?name=nginx HTTP/1.1
Host: localhost
Connection: close
Content-Length: 17
Accept: */*
Content-Type: application/json;charset=UTF-8
User-Agent: native-http-client
Cookie: User-Agent=native-http-client; 

{"Image":"nginx"}
```

### 响应

response 为字符串格式的原始响应值，例如 Json xml 等，可直接使用

```golang
response, err := client.Get("/containers/json", nil)
response, err := client.Put("/containers/create", `{"Image":"nginx"}`)
```

### API

携带 Header
```golang
client.WithHeader("User-Agent", "native-http-client")
      .WithHeader("Token", "***")
      .WithHeader("...", "***")
```

携带 Cookie    
仅当 Header 中未设置 Cookie 字段时生效
```golang
client.WithCookie("UserName", "alex")
      .WithCookie("UID", "***")
      .WithCookie("...", "***")
```

[更多](http_client.go)

响应：client.Response

```javascript
{
    "ContentType": "application/json",
    "ContentLength": 1054,
    "StatusCode": 200,
    "Headers": {
        "Connection": "close",
        "Content-Length": "1054",
        "Content-Type": "application/json",
        "Date": "Wed, 14 Oct 2020 08:37:20 GMT",
        "Server": "nginx/1.18.0"
    },
    "Cookies": {},
    "Body": "{\"Status\":\"OK\",\"...\":\"...\"}"
}
```


### 核心原理

对网络发送原始HTTP协议，以下为 POST API 对应的命令示例

unix sock

```shell
curl -s --unix-socket /var/run/docker.sock \
    -X POST http://localhost/containers/create?name=nginx \
    --data '{"Image":"nginx"}' 
```

tcp
```shell 
telnet domain 80
Trying xxxxxx...
Connected to domain.
Escape character is '^]'.
POST /containers/create?name=nginx HTTP/1.1
Host: localhost
Connection: close
Content-Length: 17
Accept: */*
Content-Type: application/json;charset=UTF-8
User-Agent: native-http-client
Cookie: User-Agent=native-http-client; 

{"Image":"nginx"}
```


### 鸣谢

- 本项目的实现一定程度上参考了 [ddliu/go-httpclient](https://github.com/ddliu/go-httpclient) 在此致谢
