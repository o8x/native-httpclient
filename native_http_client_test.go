package nativehttpclient

import (
	"fmt"
	"testing"
)

func TestUnixSock(t *testing.T) {
	sock := NewUnixSock("/var/run/docker.sock")
	post, err := sock.Post("/containers/create?name=nginx", map[string]interface{}{
		"AutoRemove": true,
		"Image":      "nginx",
	})

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println(post)
}

func TestTcp(t *testing.T) {
	// 暂无可用的公共 API，可按照 README 进行测试
}
