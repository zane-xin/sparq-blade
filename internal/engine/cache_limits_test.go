package engine

import (
	"fmt"
	"testing"
)

func TestDefaultTemplateCachesAreBounded(t *testing.T) {
	engine := New(WithFS(uniqueTemplateFS{}))
	total := maxCompiledTemplateCacheEntries + 20

	for index := 0; index < total; index++ {
		if _, err := engine.RenderFile(fmt.Sprintf("view-%d", index), nil); err != nil {
			t.Fatalf("RenderFile(%d) error = %v", index, err)
		}
	}

	if got := len(engine.templateCache); got != maxCompiledTemplateCacheEntries {
		t.Fatalf("compiled template cache size = %d, want %d", got, maxCompiledTemplateCacheEntries)
	}
	if got := len(engine.fileCache); got != maxFileTemplateCacheEntries {
		t.Fatalf("file template cache size = %d, want %d", got, maxFileTemplateCacheEntries)
	}
	memoryCache, ok := engine.cache.(*MemoryCache)
	if !ok {
		t.Fatalf("default cache type = %T, want *MemoryCache", engine.cache)
	}
	if got := len(memoryCache.items); got != maxParsedTemplateCacheEntries {
		t.Fatalf("parsed template cache size = %d, want %d", got, maxParsedTemplateCacheEntries)
	}
}

type uniqueTemplateFS struct{}

func (uniqueTemplateFS) Load(name string) (*TemplateSource, error) {
	return &TemplateSource{
		Name:    name,
		Content: []byte("<p>" + name + "</p>"),
		Version: "1",
	}, nil
}
