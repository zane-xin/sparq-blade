package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCMSExampleRoutes(t *testing.T) {
	server := newServer(renderer())

	cases := []struct {
		path string
		want []string
	}{
		{
			path: "/",
			want: []string{
				"CMS 模板示例",
				"模板查找层级",
				"/directives",
				"/templates",
				"/data",
				"/articles",
				"/navigation",
				"/pagination",
				"/custom",
			},
		},
		{
			path: "/directives",
			want: []string{
				"基础指令",
				"当前栏目是 7",
				"暂无内容",
				`<span class="badge">5</span>`,
			},
		},
		{
			path: "/templates",
			want: []string{
				"模板组合",
				"include 接收第二参数",
				"组件标题来自第二参数",
				"来自 footer slot",
				"命名空间模板",
			},
		},
		{
			path: "/data",
			want: []string{
				"数据处理",
				"Sparq CMS",
				"栏目 7 分页文章 A",
				"站点信息为空",
				`data-source="theme"`,
			},
		},
		{
			path: "/articles",
			want: []string{
				"基于上下文推导数据",
				"栏目 7 文章 A",
				"栏目 99 文章 A",
			},
		},
		{
			path: "/navigation",
			want: []string{
				"直接子级渲染",
				`<ul class="submenu">`,
				"公司动态",
				"无子级",
			},
		},
		{
			path: "/pagination",
			want: []string{
				"输出型业务标签",
				`data-source="theme"`,
				"LayeredFS 主题覆盖",
			},
		},
		{
			path: "/custom",
			want: []string{
				"自定义扩展",
				`<span class="badge">自定义扩展</span>`,
				`class="directive-panel"`,
				"输出型业务标签",
				"数据型业务指令",
			},
		},
	}

	for _, item := range cases {
		t.Run(item.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, item.path, nil)

			server.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("unexpected status: got %d body %s", response.Code, response.Body.String())
			}

			body := response.Body.String()
			for _, want := range item.want {
				if !strings.Contains(body, want) {
					t.Fatalf("response for %s missing %q:\n%s", item.path, want, body)
				}
			}
			if item.path == "/pagination" && strings.Contains(body, `data-source="module"`) {
				t.Fatalf("expected theme pager override, got module pager:\n%s", body)
			}
		})
	}
}

func TestCMSExampleHealthz(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	newServer(renderer()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Body.String() != "ok" {
		t.Fatalf("unexpected body: %q", response.Body.String())
	}
}
