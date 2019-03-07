package main

import (
	"container/list"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Pair struct {
	key     string
	value   interface{}
	curtime int64
	elem    *list.Element
}

type LRUCache struct {
	capacity int
	m        map[string]*Pair
	list     *list.List
	isExpLru bool
	expire   int64
	lock     sync.RWMutex
}

func (p *Pair) Reset() {
	p.key = ""
	p.value = ""
	p.curtime = 0
	p.elem = nil
}

func (p *Pair) Set(k string, v interface{}, curtime int64) {
	p.key = k
	p.value = v
	p.curtime = curtime
	p.elem = nil
}

var pairPool = sync.Pool{
	New: func() interface{} { return new(Pair) },
}

//expire = 0则不会有超时过期说法
func Constructor(capacity int, expire int64) *LRUCache {
	lru := new(LRUCache)
	lru.capacity = capacity
	lru.m = make(map[string]*Pair)
	lru.expire = expire
	lru.list = list.New()
	return lru
}

func (this *LRUCache) Get(key string) interface{} {
	if this.expire != 0 {
		this.lock.Lock()
		pair, ok := this.m[key]
		if ok {
			if time.Now().Unix()-pair.curtime >= this.expire { //时间过期则删除
				this.list.Remove(pair.elem)
				delete(this.m, key)

				//回收资源
				pair.Reset()
				pairPool.Put(pair)
			} else {
				this.list.MoveToFront(pair.elem)
				this.lock.Unlock()
				return pair.value
			}
		}
		this.lock.Unlock()
		return nil
	} else {
		this.lock.Lock()
		pair, ok := this.m[key]
		if ok {
			this.list.MoveToFront(pair.elem)
			this.lock.Unlock()
			return pair.value
		}
		this.lock.Unlock()
		return nil
	}
}

func (this *LRUCache) Put(key string, value interface{}) *Pair {
	if this.expire != 0 {
		this.lock.Lock()
		pair, ok := this.m[key]
		if ok {
			pair.value = value
			pair.curtime = time.Now().Unix()
			this.list.MoveToFront(pair.elem)
			this.lock.Unlock()
			return nil
		} else {
			if len(this.m) >= this.capacity {
				elem := this.list.Back()
				this.list.Remove(elem)
				if pair, ok := elem.Value.(*Pair); ok {
					delete(this.m, pair.key)

					//回收资源
					pair.Reset()
					pairPool.Put(pair)
				}
			} else {
				// 逆序遍历
				count := 0
				for elem := this.list.Back(); elem != nil; elem = elem.Prev() {
					if pair, ok := elem.Value.(*Pair); ok {
						if time.Now().Unix()-pair.curtime >= this.expire {
							//时间过期则删除
							this.list.Remove(pair.elem)
							delete(this.m, key)

							//回收资源
							pair.Reset()
							pairPool.Put(pair)
							break
						} else {
							if count++; count >= 3 {
								break
							}
						}
					}
				}
			}

			//复用对象
			pair = pairPool.Get().(*Pair)
			pair.Set(key, value, time.Now().Unix())
			//pair := &Pair{key: key, value: value, curtime: time.Now().Unix()}
			pair.elem = this.list.PushFront(pair)
			this.m[key] = pair

		}
		this.lock.Unlock()
		return pair
	} else {
		this.lock.Lock()
		pair, ok := this.m[key]
		if ok {
			pair.value = value
			this.list.MoveToFront(pair.elem)
			this.lock.Unlock()
			return nil
		} else {
			if len(this.m) >= this.capacity {
				elem := this.list.Back()
				this.list.Remove(elem)
				if pair, ok := elem.Value.(*Pair); ok {
					delete(this.m, pair.key)

					//回收资源
					pair.Reset()
					pairPool.Put(pair)
				}
			}
			//pair := &Pair{key: key, value: value}
			pair = pairPool.Get().(*Pair)
			pair.Set(key, value, 0)
			pair.elem = this.list.PushFront(pair)
			this.m[key] = pair

		}
		this.lock.Unlock()
		return pair
	}
}

const (
	MaxWorker int   = 6
	step      int64 = 100000
)

func main() {
	/*
		lru := Constructor(int(step*int64(MaxWorker-1)), 6)

		var wg sync.WaitGroup
		//wg.Add(MaxWorker)
		wg.Add(2)

		go func(count int64) {
			var m int64
			for m = 0; m < count; m++ {
				lru.Put(strconv.FormatInt(m, 10), m)
				fmt.Println("insert ", m)
			}
			wg.Done()
		}(step * int64(MaxWorker-1))

		//time.Sleep(time.Millisecond * 1000)
		go func() {
			var m int64
			for m < step*int64(MaxWorker-1) {
				val := lru.Get(strconv.FormatInt(m, 10))
				if val != nil {
					fmt.Printf("%d %d\n", m, val.(int64))
					m++
				} else {
					time.Sleep(time.Millisecond * 100)
					fmt.Printf("%s", "read nil")
				}
			}
			wg.Done()
		}()
	*/
	lru := Constructor(int(step*5), 0)

	var wg sync.WaitGroup
	wg.Add(MaxWorker)

	go func(count int64) {
		var m int64
		for m = 0; m < count; m++ {
			lru.Put(strconv.FormatInt(m, 10), m)
			fmt.Println("insert ", m)
		}
		wg.Done()
	}(step * int64(MaxWorker-1))

	/* 0 ~ 500000
	0 ~ 99999
	100000 ~ 199999
	*/

	for i := 0; i < MaxWorker-1; i++ {
		go func(n int) {
			var m int64 = int64(n) * step
			var max int64 = m + step
			for m < max {
				val := lru.Get(strconv.FormatInt(m, 10))
				if val != nil {
					fmt.Printf("%d %d\n", m, val.(int64))
					m++
				} else {
					//fmt.Printf("%s", ".")
					//m--
					//time.Sleep(time.Millisecond * 1000)
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
