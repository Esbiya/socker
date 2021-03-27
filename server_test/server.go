package main

import (
	"fmt"
	"time"

	"github.com/Esbiya/socker"
)

func main() {
	server := socker.DefaultTCPServer()
	server.Router().Register("session.login", func(msg socker.Data) (out interface{}, next socker.Handler) {
		qr := socker.GenUUIDStr()
		out = fmt.Sprintf(`{"code":200,"msg":"success","data":{"qr":"%s"}}`, qr)
		next = func(msg socker.Data) (out interface{}, next socker.Handler) {
			<-time.After(1 * time.Second)
			out = map[string]interface{}{
				"code": 200,
				"msg":  "success",
				"data": map[string]interface{}{
					"session": "哈哈哈",
				},
			}
			next = nil
			return
		}
		return
	})
	server.Run()
}
