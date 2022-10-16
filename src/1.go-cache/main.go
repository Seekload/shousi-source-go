package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"time"
)

func main() {
	c := cache.New(10*time.Second, 30*time.Second) // 默认过期时间10s；清理间隔30s，即每30s钟会自动清理过期的键值对

	// 设置一个键值对，过期时间是 3s
	c.Set("a", "testa", 3*time.Second)

	// 设置一个键值对，采用 New() 时的默认过期时间，即 10s
	c.Set("foo", "bar", cache.DefaultExpiration)

	// 设置一个键值对，没有过期时间，不会自动过期，需要手动调用 Delete() 才能删除
	c.Set("baz", 42, cache.NoExpiration)

	v, found := c.Get("a")
	fmt.Println(v, found) // testa,true

	<-time.After(5 * time.Second) // 延时5s

	v, found = c.Get("a") // nil,false
	fmt.Println(v, found)

	<-time.After(6 * time.Second)
	v, found = c.Get("foo") // nil,false
	fmt.Println(v, found)

	v, found = c.Get("baz") // 42,true
	fmt.Println(v, found)

	// 完整例子请关注公众号【Golang来啦】，后台发送关键字 gocache 获取
	//TestCache()
	//TestCacheTimes()
	//TestNewFrom()
	//TestOnEvicted()
	//TestFileSerialization()
}

func TestCache() {
	// 如果第一个参数小于等于0，则默认过期时间小于0，这意味着如果执行 Set() 等操作时，如果默认过期时间小于0，则该键值对不会过期。
	tc := cache.New(cache.DefaultExpiration, 0) // 没有过期时间；定期清理周期为 0 表示不开启定期清理功能

	a, found := tc.Get("a")
	if found || a != nil {
		fmt.Println("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		fmt.Println("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		fmt.Println("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, cache.DefaultExpiration)
	tc.Set("b", "b", cache.DefaultExpiration)
	tc.Set("c", 3.5, cache.DefaultExpiration)

	x, found := tc.Get("a")
	if !found {
		fmt.Println("a was not found while getting a2")
	}
	if x == nil {
		fmt.Println("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		fmt.Println("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		fmt.Println("b was not found while getting b2")
	}
	if x == nil {
		fmt.Println("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		fmt.Println("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		fmt.Println("c was not found while getting c2")
	}
	if x == nil {
		fmt.Println("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		fmt.Println("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes() {
	var found bool

	tc := cache.New(50*time.Millisecond, 1*time.Millisecond)
	tc.Set("a", 1, cache.DefaultExpiration)
	tc.Set("b", 2, cache.NoExpiration)
	tc.Set("c", 3, 20*time.Millisecond)
	tc.Set("d", 4, 70*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = tc.Get("c")
	if found {
		fmt.Println("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		fmt.Println("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		fmt.Println("Did not find b even though it was set to never expire")
	}

	_, found = tc.Get("d")
	if !found {
		fmt.Println("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(20 * time.Millisecond)
	_, found = tc.Get("d")
	if found {
		fmt.Println("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestNewFrom() {
	m := map[string]cache.Item{
		"a": cache.Item{
			Object:     1,
			Expiration: 0,
		},
		"b": cache.Item{
			Object:     2,
			Expiration: 0,
		},
	}
	tc := cache.NewFrom(cache.DefaultExpiration, 0, m)
	a, found := tc.Get("a")
	if !found {
		fmt.Println("Did not find a")
	}
	if a.(int) != 1 {
		fmt.Println("a is not 1")
	}
	b, found := tc.Get("b")
	if !found {
		fmt.Println("Did not find b")
	}
	if b.(int) != 2 {
		fmt.Println("b is not 2")
	}
}

func TestOnEvicted() {
	tc := cache.New(cache.DefaultExpiration, 0)
	tc.Set("foo", 3, cache.DefaultExpiration)

	works := false
	tc.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
			fmt.Println(works)
		}
		tc.Set("bar", 4, cache.DefaultExpiration)
	})
	tc.Delete("foo")
	x, _ := tc.Get("bar")
	if !works {
		fmt.Println("works bool not true")
	}
	if x.(int) != 4 {
		fmt.Println("bar was not 4")
	}
}

func TestFileSerialization() {
	tc := cache.New(cache.DefaultExpiration, 0)
	tc.Add("a", "a", cache.DefaultExpiration)
	tc.Add("b", "b", cache.DefaultExpiration)
	f, err := ioutil.TempFile("", "go-cache-cache.dat")
	if err != nil {
		fmt.Println("Couldn't create cache file:", err)
	}
	fname := f.Name()
	f.Close()
	tc.SaveFile(fname)

	oc := cache.New(cache.DefaultExpiration, 0)
	oc.Add("a", "aa", 0) // this should not be overwritten
	err = oc.LoadFile(fname)
	if err != nil {
		fmt.Println(err)
	}
	a, found := oc.Get("a")
	if !found {
		fmt.Println("a was not found")
	}
	astr := a.(string)
	if astr != "aa" {
		if astr == "a" {
			fmt.Println("a was overwritten")
		} else {
			fmt.Println("a is not aa")
		}
	}
	b, found := oc.Get("b")
	if !found {
		fmt.Println("b was not found")
	}
	if b.(string) != "b" {
		fmt.Println("b is not b")
	}
}
