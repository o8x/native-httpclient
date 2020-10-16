package nativehttpclient

import (
	"fmt"
	"testing"
)

func TestUnixSock(t *testing.T) {
	sock := NewUnixSock("/var/run/docker.sock")

	created, createErr := sock.Post("/containers/create?name=84b5ffa51d51b", map[string]interface{}{
		"AutoRemove": true,
		"Image":      "centos",
	})
	if createErr != nil {
		fmt.Println("createErr:", createErr)
	}
	fmt.Println(created)

	killed, killErr := sock.Post("/containers/84b5ffa51d51b/kill", nil)
	if killErr != nil {
		fmt.Println("createErr:", killErr)
	}
	fmt.Println(killed)
}

func TestTcp(t *testing.T) {
	// 暂无可用的公共 API，可按照 README 进行测试
}
