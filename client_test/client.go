/*
 * @Author: your name
 * @Date: 2021-03-25 13:43:56
 * @LastEditTime: 2021-03-26 19:11:17
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /socker/client_test/client.go
 */
package main

import (
	"sync"

	"github.com/Esbiya/loguru"
	"github.com/Esbiya/socker"
)

func main() {
	c := socker.DefaultTCPClient()
	c.Start()
	wg := sync.WaitGroup{}
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Send("session.login", map[string]interface{}{
				"xxx": "111",
			}).Then(func(b interface{}) {
				loguru.Debug(b)
			}).Then(func(b interface{}) {
				loguru.Debug(b)
			}).Close()
		}()
	}
	wg.Wait()
	c.Close()
}
