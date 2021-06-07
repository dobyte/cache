# cache
A Cache Library Similar To Laravel-Cache

Support Redis、Memcached

## Use

Download and install

```shell script
go get github.com/dobyte/cache
```

API

```text
// Determine if an item exists in the cache.
Has(key string) (bool, error)
// Determine if multiple item exists in the cache.
HasMany(keys ...string) (map[string]bool, error)
// Retrieve an item from the cache by key.
Get(key string, defaultValue ...interface{}) Result
// Retrieve multiple items from the cache by key.
GetMany(keys ...string) (map[string]string, error)
// Retrieve or set an item from the cache by key.
GetSet(key string, fn func() (interface{}, time.Duration, error)) Result
// Store an item in the cache.
Set(key string, value interface{}, expire time.Duration) error
// Store multiple items in the cache for a given number of expire.
SetMany(values map[string]interface{}, expire time.Duration) error
// Store an item in the cache indefinitely.
Forever(key string, value interface{}) error
// Store multiple items in the cache indefinitely.
ForeverMany(values map[string]interface{}) error
// Store an item in the cache if the key does not exist.
Add(key string, value interface{}, expire time.Duration) (bool, error)
// Increment the value of an item in the cache.
Increment(key string, value int64) (int64, error)
// Increment the value of multiple items in the cache.
IncrementMany(values map[string]int64) (map[string]int64, error)
// Decrement the value of an item in the cache.
Decrement(key string, value int64) (int64, error)
// Decrement the value of multiple items in the cache.
DecrementMany(values map[string]int64) (map[string]int64, error)
// Remove an item from the cache.
Forget(key string) error
// Remove multiple items from the cache.
ForgetMany(keys ...string) (int64, error)
// Remove all items from the cache.
Flush() error
// Get a lock instance.
Lock(name string, time time.Duration) Lock
// Get a client instance.
GetClient() interface{}
```

Dome

```go
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
		log.Fatalf("Failed to retrieve cache")
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
```

## Example

View demo [example/main.go](example/main.go)