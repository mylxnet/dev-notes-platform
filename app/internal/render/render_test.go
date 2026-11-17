package render

import (
	"strings"
	"testing"
)

func TestMarkdownBasic(t *testing.T) {
	html := Markdown("# 标题\n\n正文 **加粗** 与 `代码`。")
	for _, want := range []string{"<h1", "标题", "<strong>加粗</strong>", "<code>代码</code>"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in %q", want, html)
		}
	}
}

func TestMarkdownStripsScript(t *testing.T) {
	html := Markdown("正常内容\n\n<script>alert(1)</script>\n\n<img src=x onerror=alert(1)>")
	if strings.Contains(html, "<script") || strings.Contains(html, "onerror") {
		t.Fatalf("script or handler leaked: %q", html)
	}
}

func TestMarkdownDropsJavascriptLink(t *testing.T) {
	html := Markdown("[点我](javascript:alert(1))")
	if strings.Contains(html, "javascript:") {
		t.Fatalf("javascript protocol leaked: %q", html)
	}
}

func TestMarkdownAllowsOnlyInternalImages(t *testing.T) {
	internal := Markdown("![图](/assets/demo.png)")
	if !strings.Contains(internal, "/assets/demo.png") {
		t.Fatalf("expected internal image kept, got %q", internal)
	}

	external := Markdown("![图](https://evil.example/track.gif)")
	if strings.Contains(external, "evil.example") {
		t.Fatalf("external image should be removed, got %q", external)
	}
}

func TestMarkdownKeepsCodeLanguageClass(t *testing.T) {
	html := Markdown("```go\nfmt.Println(1)\n```")
	if !strings.Contains(html, `class="language-go"`) {
		t.Fatalf("expected language class, got %q", html)
	}
}

func TestMarkdownEmpty(t *testing.T) {
	if got := Markdown("   \n\n "); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestMarkdownTable(t *testing.T) {
	html := Markdown("| 列 | 值 |\n|---|---|\n| a | b |")
	if !strings.Contains(html, "<table>") || !strings.Contains(html, "<td>a</td>") {
		t.Fatalf("expected table output, got %q", html)
	}
}
