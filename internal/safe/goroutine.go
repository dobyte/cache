/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/3/8 1:01 下午
 * @Desc: TODO
 */

package safe

import (
	"log"
	"runtime"
)

func Go(f func()) {
	go func() {
		defer func() {
			switch err := recover(); err.(type) {
			case runtime.Error:
				log.Printf("runtime error: %v\n", err)
			default:
				log.Printf("error: %v\n", err)
			}
		}()

		f()
	}()
}
