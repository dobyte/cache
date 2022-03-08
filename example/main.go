/**
 * @Author: wanglin
 * @Email: wanglin@vspn.com
 * @Date: 2021/6/7 9:50 上午
 * @Desc: example of cache.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dobyte/cache"
)

type student struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Birthday string `json:"birthday"`
}

func main() {
	c := cache.NewCache(&cache.Options{
		Driver: cache.RedisDriver,
		Prefix: "cache",
		Stores: cache.Stores{
			Redis: &cache.RedisOptions{
				Addrs: []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002"},
			},
		},
	})

	// The GetSet method first reads data from the cache.
	// If the read fails, an error is returned directly.
	// If the read data is nil, the data is obtained from the fn function and stored in the cache.
	// If an error occurs when reading the fn function data, an error will be returned directly.
	// If the fn function returns an error of cache.Nil,
	// the default null value (cache@nil) will be stored in the cache for a certain period of time (10s).
	{
		rst1 := c.GetSet(context.TODO(), "name", func() (interface{}, time.Duration, error) {
			return "fuxiao", 10 * time.Second, nil
		})
		if err := rst1.Err(); err != nil && err != cache.Nil {
			log.Fatalf("Failed to retrieve cache: %v", err.Error())
		} else {
			fmt.Println(rst1.Val())
		}
	}

	// No data found from fn function
	{
		rst2 := c.GetSet(context.TODO(), "fullname", func() (interface{}, time.Duration, error) {
			return nil, 0, cache.Nil
		})
		if err := rst2.Err(); err != nil && err != cache.Nil {
			log.Fatalf("Failed to retrieve cache: %v", err.Error())
		} else {
			fmt.Println(rst2.Val())
		}
	}

	{
		rst3 := c.GetSet(context.TODO(), "fuxiao", func() (interface{}, time.Duration, error) {
			return &student{
				Name:     "fuxiao",
				Age:      30,
				Birthday: "1991-03-11",
			}, 1 * time.Hour, nil
		})

		var s student

		if err := rst3.Scan(&s); err != nil {
			log.Fatalf("Failed to retrieve cache: %v", err.Error())
		} else {
			fmt.Println(s.Name)
		}
	}
}
