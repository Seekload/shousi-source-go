package main

import (
	"fmt"
	"github.com/muesli/cache2go"
	"strconv"
	"time"
)

type myStruct struct {
	text     string
	moreData []byte
}

var (
	k = "testkey"
	v = "testvalue"
)

func main() {

	cache := cache2go.Cache("myCache") // 创建CacheTable

	val := myStruct{"This is a test!", []byte{}}
	cache.Add("someKey", 5*time.Second, &val) // 5s 是 item 的存活时间，超过 5s 不被访问就会被删除

	res, err := cache.Value("someKey") // 获取item
	if err == nil {
		fmt.Println("Found value in cache:", res.Data().(*myStruct).text)
	} else {
		fmt.Println("Error retrieving value from cache:", err)
	}

	time.Sleep(6 * time.Second)       // 休眠6s  5s过后，上面添加的 item 会过期，自动被删除
	res, err = cache.Value("someKey") // 再次获取 item，不存在会报错
	if err != nil {
		fmt.Println("Item is not cached (anymore).")
	}

	cache.Add("someKey", 0, &val)                                    // 无过期时间，表示不会过期
	cache.SetAboutToDeleteItemCallback(func(e *cache2go.CacheItem) { // 设置删除回调函数，删除 item 时会自动触发
		fmt.Println("Deleting:", e.Key(), e.Data().(*myStruct).text, e.CreatedOn())
	})

	cache.Delete("someKey") // 手动删除item，会触发删除回调函数

	cache.Flush() // 清空table

	// 完整例子请关注公众号【Golang来啦】，后台发送关键字 cache2go 获取
	//TestCache()
	//TestCacheExpire()
	//TestExists()
	//TestNotFoundAdd()
	//TestCacheKeepAlive()
	//TestDelete()
	//TestFlush()
	//TestCount()
	//TestAccessCount()
	//TestCallbacks()
	//TestDataLoader()
	// 完整例子请关注公众号【Golang来啦】，后台发送关键字 cache2go 获取
}

func TestCache() {
	table := cache2go.Cache("testCache")
	table.Add(k+"_1", 0*time.Second, v)
	table.Add(k+"_2", 1*time.Second, v) // 这一步会触发过期检查操作

	//time.Sleep(2 * time.Second)   // 延时 2s，第二个添加的 item 会过期，被自动删除

	p, err := table.Value(k + "_1")
	if err != nil || p == nil || p.Data().(string) != v {
		fmt.Println("Error retrieving non expiring data from cache", err)
	}
	p, err = table.Value(k + "_2")
	if err != nil || p == nil || p.Data().(string) != v {
		fmt.Println("Error retrieving data from cache", err)
		return // 返回，如果不存在的话后面的调用会报错
	}

	// 合理性检查
	if p.AccessCount() != 1 {
		fmt.Println("Error getting correct access count")
	}
	if p.LifeSpan() != 1*time.Second {
		fmt.Println("Error getting correct life-span")
	}
	if p.AccessedOn().Unix() == 0 {
		fmt.Println("Error getting access time")
	}
	if p.CreatedOn().Unix() == 0 {
		fmt.Println("Error getting creation time")
	}
}

func TestCacheExpire() {
	table := cache2go.Cache("testCache")

	table.Add(k+"_1", 250*time.Millisecond, v+"_1")
	table.Add(k+"_2", 200*time.Millisecond, v+"_2")

	time.Sleep(100 * time.Millisecond)

	// 检查是否存在
	_, err := table.Value(k + "_1")
	if err != nil {
		fmt.Println("Error retrieving value from cache:", err)
	}

	time.Sleep(150 * time.Millisecond)

	// 再次检查是否存在
	_, err = table.Value(k + "_1") // 已经累计延时 250ms，知道此时这个 item 为什么还存在吗？
	if err != nil {
		fmt.Println("Error retrieving value from cache:", err)
	}

	// 检查 key `2` 是否存在
	_, err = table.Value(k + "_2")
	if err != nil {
		fmt.Println("Found key which should have been expired by now")
	}
}

func TestDataLoader() {
	cache := cache2go.Cache("myCache")

	// 当 item 不存在时，会自动调用 data-loader，实现从数据库、文件等处加载数据到缓存中
	cache.SetDataLoader(func(key interface{}, args ...interface{}) *cache2go.CacheItem {
		// 在这里实现业务逻辑，比如读取数据库等
		val := "This is a test with key " + key.(string)

		// 创建一个新的 item 并返回
		item := cache2go.NewCacheItem(key, 0, val)
		return item
	})

	// 自动生成一些数据
	for i := 0; i < 10; i++ {
		res, err := cache.Value("someKey_" + strconv.Itoa(i))
		if err == nil {
			fmt.Println("Found value in cache:", res.Data())
		} else {
			fmt.Println("Error retrieving value from cache:", err)
		}
	}
}

func TestExists() {
	// add an expiring item
	table := cache2go.Cache("testExists")
	table.Add(k, 0, v)
	// check if it exists
	if !table.Exists(k) {
		fmt.Println("Error verifying existing data in cache")
	}
}

func TestNotFoundAdd() {
	table := cache2go.Cache("testNotFoundAdd")

	if !table.NotFoundAdd(k, 0, v) {
		fmt.Println("Error verifying NotFoundAdd, data not in cache")
	}

	if table.NotFoundAdd(k, 0, v) {
		fmt.Println("Error verifying NotFoundAdd data in cache")
	}
}

func TestCacheKeepAlive() {
	table := cache2go.Cache("testKeepAlive")
	p := table.Add(k, 250*time.Millisecond, v)

	time.Sleep(100 * time.Millisecond)
	p.KeepAlive() // 更新访问时间

	time.Sleep(150 * time.Millisecond)
	if !table.Exists(k) {
		fmt.Println("Error keeping item alive")
	}

	time.Sleep(300 * time.Millisecond)
	if table.Exists(k) {
		fmt.Println("Error expiring item after keeping it alive")
	}
}

func TestDelete() {
	table := cache2go.Cache("testDelete")
	table.Add(k, 0, v)

	p, err := table.Value(k) // 获取item
	if err != nil || p == nil || p.Data().(string) != v {
		fmt.Println("Error retrieving data from cache", err)
	}

	table.Delete(k)         // 删除
	p, err = table.Value(k) // 获取item，key不存在，Value()会报错
	if err == nil || p != nil {
		fmt.Println("Error deleting data")
	}

	_, err = table.Delete(k) // key不存在，Delete()返回错误
	if err == nil {
		fmt.Println("Expected error deleting item")
	}
}

func TestFlush() {
	table := cache2go.Cache("testFlush")
	table.Add(k, 10*time.Second, v)

	table.Flush() // 清空table

	p, err := table.Value(k) // 获取item
	if err == nil || p != nil {
		fmt.Println("Error flushing table")
	}
	if table.Count() != 0 {
		fmt.Println("Error verifying count of flushed table")
	}
}

func TestCount() {
	table := cache2go.Cache("testCount")
	count := 100000
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		table.Add(key, 10*time.Second, v)
	}
	// 确认每个元素已经添加到缓存中
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		p, err := table.Value(key)
		if err != nil || p == nil || p.Data().(string) != v {
			fmt.Println("Error retrieving data")
		}
	}
	if table.Count() != count {
		fmt.Println("Data count mismatch")
	}
}

func TestAccessCount() {
	count := 100
	table := cache2go.Cache("testAccessCount")
	for i := 0; i < count; i++ {
		table.Add(i, 10*time.Second, v)
	}
	// 第1个 item 不会被访问，第2个item 被访问1次，第3个item被访问2次，以此类推...
	for i := 0; i < count; i++ {
		for j := 0; j < i; j++ {
			table.Value(i)
		}
	}

	ma := table.MostAccessed(int64(count))
	for i, item := range ma {
		if item.Key() != count-1-i {
			fmt.Println("Most accessed items seem to be sorted incorrectly")
		}
	}

	ma = table.MostAccessed(int64(count - 1)) // 返回前 99 个元素
	if len(ma) != count-1 {
		fmt.Println("MostAccessed returns incorrect amount of items")
	}
}

func TestCallbacks() {
	cache := cache2go.Cache("myCache")

	// 每次添加新的 item 时会触发回调函数(添加)
	cache.SetAddedItemCallback(func(entry *cache2go.CacheItem) {
		fmt.Println("Added Callback 1:", entry.Key(), entry.Data(), entry.CreatedOn())
	})
	cache.AddAddedItemCallback(func(entry *cache2go.CacheItem) {
		fmt.Println("Added Callback 2:", entry.Key(), entry.Data(), entry.CreatedOn())
	})
	// 每次删除 item 时会触发回调函数(删除)
	cache.SetAboutToDeleteItemCallback(func(entry *cache2go.CacheItem) {
		fmt.Println("Deleting:", entry.Key(), entry.Data(), entry.CreatedOn())
	})

	// 添加时会触发回调函数(添加)
	cache.Add("someKey", 0, "This is a test!")

	// 获取 item
	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("Found value in cache:", res.Data())
	} else {
		fmt.Println("Error retrieving value from cache:", err)
	}

	// 删除时会触发回调函数(删除)
	cache.Delete("someKey")

	cache.RemoveAddedItemCallbacks() // 删除回调函数(添加)
	// 添加一个存活时长为 3s 的 item
	res = cache.Add("anotherKey", 3*time.Second, "This is another test")

	// 设置 item 过期回调函数，item 被自动删除时会调用
	res.SetAboutToExpireCallback(func(key interface{}) {
		fmt.Println("About to expire:", key.(string))
	})

	time.Sleep(5 * time.Second)
}
