package engine

import (
	"sync"

	"github.com/zane-xin/sparq-blade/internal/compiler"
)

const maxParsedTemplateCacheEntries = 512

// Cache 保存解析后的模板 AST，不保存最终 Go 模板；实现必须支持 Engine 的并发读写。
//
// Get 返回的 AST 由调用方独占，Set 写入后也不能继续引用调用方持有的 AST。
type Cache interface {
	Get(name string, version string) (*compiler.Document, bool)
	Set(name string, version string, document *compiler.Document)
}

type cacheEntry struct {
	version  string
	document *compiler.Document
}

// MemoryCache 最多保存 512 个模板名称，达到上限后按首次写入顺序淘汰，更新同名模板不会新增条目。
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	order []string
}

// NewMemoryCache 创建空的并发安全 AST 缓存。
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]cacheEntry),
	}
}

// Get 仅在名称存在且版本完全一致时命中；版本变化会交由调用方重新解析。
func (c *MemoryCache) Get(name string, version string) (*compiler.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[name]
	if !ok || entry.version != version {
		return nil, false
	}

	return compiler.CloneDocument(entry.document), true
}

// Set 写入或替换指定模板的 AST；新增名称可能触发最早条目的淘汰。
func (c *MemoryCache) Set(name string, version string, document *compiler.Document) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[name]; !exists {
		if len(c.order) >= maxParsedTemplateCacheEntries {
			oldest := c.order[0]
			delete(c.items, oldest)
			c.order = c.order[1:]
		}
		c.order = append(c.order, name)
	}
	c.items[name] = cacheEntry{
		version:  version,
		document: compiler.CloneDocument(document),
	}
}
