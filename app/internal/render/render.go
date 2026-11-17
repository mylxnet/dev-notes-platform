// Package render 把 Markdown 渲染为安全 HTML。
//
// 管线：goldmark 解析（禁用原始 HTML）→ bluemonday 白名单净化。
// 前端只对净化后的 HTML 使用 v-html。
package render

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var (
	once   sync.Once
	engine goldmark.Markdown
	policy *bluemonday.Policy
	// 图片只允许站内路径：/assets/ 是前端构建产物，/media/ 是笔记配图。
	localImageSRC = regexp.MustCompile(`^/(?:assets|media)/[A-Za-z0-9._~/-]+$`)
	linkHref      = regexp.MustCompile(`^(?:https?://|/|#)[^\s]*$`)
	codeLang      = regexp.MustCompile(`^language-[a-z0-9+#-]+$`)
)

func setup() {
	// 不启用 html.WithUnsafe()：Markdown 里的原始 HTML 会被丢弃，不进入渲染结果。
	engine = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	p := bluemonday.NewPolicy()
	// 站内图片与相对链接（/assets/…、/notes/…）必须允许相对 URL，否则 src/href 会被清空。
	p.AllowRelativeURLs(true)
	p.AllowElements(
		"p", "br", "hr", "strong", "b", "em", "i", "del", "s", "sup", "sub",
		"blockquote", "ul", "ol", "li",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"pre", "code", "table", "thead", "tbody", "tr", "th", "td",
		"a", "img",
	)
	p.AllowAttrs("href").Matching(linkHref).OnElements("a")
	p.AllowAttrs("title").OnElements("a")
	p.AllowAttrs("align").OnElements("th", "td")
	// 图片只允许站内路径，避免外链图片与追踪像素。
	p.AllowAttrs("src").Matching(localImageSRC).OnElements("img")
	p.AllowAttrs("alt", "title").OnElements("img")
	p.AllowAttrs("class").Matching(codeLang).OnElements("code")
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)

	policy = p
}

// Markdown 把 Markdown 源码渲染为净化后的 HTML；空输入返回空串。
func Markdown(source string) string {
	if strings.TrimSpace(source) == "" {
		return ""
	}
	once.Do(setup)

	var buf bytes.Buffer
	if err := engine.Convert([]byte(source), &buf); err != nil {
		return ""
	}
	return strings.TrimSpace(policy.Sanitize(buf.String()))
}
