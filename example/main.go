/**
 * @Author: wanglin
 * @Email: wanglin@vspn.com
 * @Date: 2021/6/7 9:50 上午
 * @Desc: TODO
 */

package main

import (
	"fmt"
	"log"
	"time"
	
	"github.com/dobyte/cache"
)

func main() {
	c := cache.NewCache(&cache.Options{
		Driver: cache.RedisDriver,
		Prefix: "cache",
		Stores: cache.Stores{
			Redis: &cache.RedisOptions{
				Addrs: []string{"127.0.0.1:6379"},
			},
		},
	})
	
	// The GetSet method first reads data from the cache.
	// If the read fails, an error is returned directly.
	// If the read data is nil, the data is obtained from the fn function and stored in the cache.
	// If an error occurs when reading the fn function data, an error will be returned directly.
	// If the fn function returns an error of cache.Nil,
	// the default null value (cache@nil) will be stored in the cache for a certain period of time (10s).
	rst1 := c.GetSet("name", func() (interface{}, time.Duration, error) {
		return "fuxiao", 10 * time.Second, nil
	})
	if err := rst1.Err(); err != nil && err != cache.Nil {
		log.Fatalf("Failed to retrieve cache: %v", err.Error())
	}
	
	fmt.Println(rst1.Val())
	
	// No data found from fn function
	rst2 := c.GetSet("fullname", func() (interface{}, time.Duration, error) {
		return nil, 0, cache.Nil
	})
	if err := rst2.Err(); err != nil && err != cache.Nil {
		log.Fatalf("Failed to retrieve cache: %v", err.Error())
	}
	
	fmt.Println(rst2.Val())
}
