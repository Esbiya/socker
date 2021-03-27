package main

import (
	"sync"
	"time"

	"github.com/Esbiya/loguru"
	"github.com/Esbiya/socker"
)

func main() {
	c, _ := socker.NewClient(&socker.Config{
		Mode:        socker.TCP,
		Addr:        socker.DefaultTCPAddr,
		InitCap:     500,
		MaxIdle:     500,
		MaxCap:      1 << 18,
		IdleTimeout: 5 * time.Second,
	})
	defer c.Close()

	c.Start()
	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Send("session.login", map[string]interface{}{
				"xxx": "111",
			}).Then(func(b interface{}) {
				loguru.Debug(b)
			}).Then(func(b interface{}) {
				loguru.Debug(b)
			})
		}()
	}
	wg.Wait()
}
