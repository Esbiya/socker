/*
 * @Author: your name
 * @Date: 2021-03-26 16:11:06
 * @LastEditTime: 2021-03-26 16:13:06
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /socker/utils_test.go
 */

package socker

import (
	"sync"
	"testing"

	"github.com/Esbiya/loguru"
)

func TestUtils(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			loguru.Debug("%d: %s", index, GenUUIDStr())
		}(i)
	}
	wg.Wait()
}
