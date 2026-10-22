package engine

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateFS 按逻辑名称读取模板源码。实现需要支持并发读取，并在未找到时返回可由 os.IsNotExist 识别的错误。
type TemplateFS interface {
	Load(name string) (*TemplateSource, error)
}

// TemplateSource 是已经完成定位的模板源码及其缓存身份。
type TemplateSource struct {
	// Name 是同一模板来源内稳定且唯一的规范名称，用于错误调用链和文件模板缓存。
	Name string
	// Content 是待词法分析的 UTF-8 模板源码；返回后应视为不可变数据。
	Content []byte
	// Version 在源码内容变化时必须同步变化，用于使 AST 和编译模板缓存失效。
	Version string
}

// LayeredFS 按顺序查找多个模板文件系统，前面的层可覆盖后面层中的同名模板。
type LayeredFS struct {
	// Layers 按优先级从高到低排列；nil 层会跳过。
	Layers []TemplateFS
}

// Load 从多个模板层中查找模板；仅当当前层返回 os.ErrNotExist 时继续查找下一层。
func (fs LayeredFS) Load(name string) (*TemplateSource, error) {
	for index, layer := range fs.Layers {
		if layer == nil {
			continue
		}

		source, err := layer.Load(name)
		if err == nil {
			return layerTemplateSource(index, source), nil
		}
		if os.IsNotExist(err) {
			continue
		}
		return nil, err
	}

	return nil, os.ErrNotExist
}

func layerTemplateSource(index int, source *TemplateSource) *TemplateSource {
	if source == nil {
		return nil
	}

	next := *source
	next.Name = fmt.Sprintf("layer%d:%s", index, source.Name)
	return &next
}

// OSFS 从本地受限目录读取模板。Root 处理普通模板名，Namespaces 为 namespace:view 配置独立根目录。
type OSFS struct {
	// Root 是无命名空间模板的根目录；空值表示不提供普通模板来源。
	Root string
	// Namespaces 把模板命名空间映射到各自根目录，例如 cms:pages.home 只会在 cms 对应目录中查找。
	Namespaces map[string]string
}

var defaultTemplateExtensions = [...]string{".tpl", ".html"}

// Load 只读取 Root 或对应命名空间根目录内的 .tpl/.html 文件，并拒绝绝对路径、上级目录和符号链接逃逸。
func (fs OSFS) Load(name string) (*TemplateSource, error) {
	namespace, logicalName := splitTemplateNamespace(name)
	root, prefix, ok := fs.resolveRoot(namespace)
	if !ok {
		return nil, os.ErrNotExist
	}

	for _, candidate := range fs.candidates(logicalName) {
		source, err := fs.loadCandidate(root, prefix, candidate)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		return source, nil
	}

	return nil, os.ErrNotExist
}

func (fs OSFS) candidates(name string) []string {
	seen := make(map[string]struct{})
	results := make([]string, 0, 8)

	add := func(candidate string) {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			return
		}

		seen[candidate] = struct{}{}
		results = append(results, candidate)
	}

	if fs.hasExplicitExtension(name) {
		add(name)
		add(fs.dottedCandidate(name))
		return results
	}

	dotted := strings.ReplaceAll(name, ".", string(os.PathSeparator))
	for _, extension := range fs.extensions() {
		add(name + extension)
		add(dotted + extension)
	}

	return results
}

func (fs OSFS) extensions() []string {
	return defaultTemplateExtensions[:]
}

func (fs OSFS) hasExplicitExtension(name string) bool {
	for _, extension := range fs.extensions() {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}

	return false
}

func (fs OSFS) dottedCandidate(name string) string {
	for _, extension := range fs.extensions() {
		if strings.HasSuffix(name, extension) {
			base := strings.TrimSuffix(name, extension)
			return strings.ReplaceAll(base, ".", string(os.PathSeparator)) + extension
		}
	}

	return strings.ReplaceAll(name, ".", string(os.PathSeparator))
}

func (fs OSFS) resolveRoot(namespace string) (string, string, bool) {
	if namespace == "" {
		return fs.Root, "", fs.Root != ""
	}

	root, ok := fs.Namespaces[namespace]
	if !ok || root == "" {
		return "", "", false
	}

	return root, namespace + ":", true
}

func (fs OSFS) loadCandidate(root string, prefix string, candidate string) (*TemplateSource, error) {
	path, err := safeTemplatePath(root, candidate)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &TemplateSource{
		Name:    prefix + candidate,
		Content: content,
		Version: fmt.Sprintf("%x", sha256.Sum256(content)),
	}, nil
}

func safeTemplatePath(root string, candidate string) (string, error) {
	if root == "" || !filepath.IsLocal(candidate) {
		return "", os.ErrNotExist
	}

	// 先检查词法路径，拒绝显式目录穿越；再解析符号链接检查真实路径，避免
	// 根目录内的链接指向模板目录之外。
	cleanCandidate := filepath.Clean(candidate)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(filepath.Join(absRoot, cleanCandidate))
	if err != nil {
		return "", err
	}

	if !isPathWithinRoot(absRoot, absPath) {
		return "", os.ErrNotExist
	}

	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", err
	}
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", err
	}
	if !isPathWithinRoot(realRoot, realPath) {
		return "", os.ErrNotExist
	}

	return realPath, nil
}

func isPathWithinRoot(root string, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	return relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relative)
}

func splitTemplateNamespace(name string) (string, string) {
	namespace, logicalName, ok := strings.Cut(name, ":")
	if !ok {
		return "", name
	}

	if namespace == "" || logicalName == "" {
		return "", name
	}

	return namespace, logicalName
}
