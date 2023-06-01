package filterql

import (
	"container/list"
	"sync"
)

type CacheProvider interface {
	Load(query string) (BoolAst, bool)
	Store(query string, cond BoolAst)
}

// MapCache 简单封装sync.Map实现的Cache，仅用于测试

type mapCache struct {
	sync.RWMutex
	m map[string]BoolAst
}

func NewMapCache() *mapCache { return &mapCache{m: make(map[string]BoolAst)} }

func (c *mapCache) Load(query string) (rv BoolAst, found bool) {
	c.RLock()
	defer c.RUnlock()
	rv, found = c.m[query]
	return
	/*
		if data, found := c.m.Load(query); !found {
			return nil, false
		} else if cond, is := data.(BoolAst); !is {
			return nil, false
		} else {
			return cond, true
		}
	*/
}

func (c *mapCache) Store(query string, cond BoolAst) {
	c.Lock()
	defer c.Unlock()
	c.m[query] = cond
}

// LruCache

type lruItem struct {
	query string
	cond  BoolAst
}

type lruCache struct {
	lock     sync.Mutex
	items    *list.List
	capacity int
	m        map[string]*list.Element
}

func NewLRUCache(capacity int) *lruCache {
	return &lruCache{
		items:    list.New(),
		capacity: capacity,
		m:        make(map[string]*list.Element),
	}
}

func (c *lruCache) Load(query string) (BoolAst, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if el, found := c.m[query]; !found {
		return nil, false
	} else if item, is := el.Value.(lruItem); !is {
		return nil, false
	} else {
		c.items.MoveToFront(el)
		return item.cond, true
	}
}

func (c *lruCache) Store(query string, cond BoolAst) {
	c.lock.Lock()
	defer c.lock.Unlock()
	item := lruItem{query: query, cond: cond}
	if el, found := c.m[query]; found {
		el.Value = item
		c.items.MoveToFront(el)
	} else {
		newEl := c.items.PushFront(item)
		c.m[query] = newEl
		if c.items.Len() > c.capacity {
			if back := c.items.Back(); back != nil {
				c.items.Remove(back)
				delete(c.m, back.Value.(lruItem).query)
			}
		}
	}
}
