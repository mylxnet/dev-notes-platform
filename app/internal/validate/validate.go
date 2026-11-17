// Package validate 实现内容字段校验与备案边界词汇拦截。
//
// 所有规则在服务端执行，前端校验只作为交互提示，不能替代这里的检查。
package validate

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"dev-notes-platform/internal/model"
)

const (
	maxTitleRunes       = 120
	maxSummaryRunes     = 300
	maxContentBytes     = 100 * 1024
	maxDescriptionBytes = 20 * 1024
	maxTechStackSize    = 10
	maxTechItemRunes    = 30
	minSlugRunes        = 2
	maxSlugRunes        = 80
	minPasswordRunes    = 12
)

// FieldError 描述单个字段的校验失败，API 层据此返回 422。
type FieldError struct {
	Field      string
	Message    string
	Hits       []string
	Suggestion string
}

func (e *FieldError) Error() string { return e.Message }

func fail(field, message string) *FieldError {
	return &FieldError{Field: field, Message: message}
}

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	repoPattern = regexp.MustCompile(`^https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(?:\.git)?/?$`)
	// 封面只允许站内路径，与渲染层放行的图片前缀保持一致。
	coverPattern = regexp.MustCompile(`^/(?:assets|media)/[A-Za-z0-9._~/-]+\.(?:png|jpe?g|webp|svg|gif)$`)
	// 密码要求同时含字母与数字。
	hasLetter = regexp.MustCompile(`[A-Za-z]`)
	hasDigit  = regexp.MustCompile(`[0-9]`)
	// 备案号只允许汉字、字母、数字与连字符，例如「甘ICP备12345678号-1」。
	icpPattern = regexp.MustCompile(`^[\p{Han}A-Za-z0-9-]{5,100}$`)
)

// 硬拦截：经营性词汇，个人非经营性备案站点不得出现。
var hardWords = []string{
	"售卖", "下单", "支付", "收费", "会员", "广告",
	"赞助", "捐赠", "破解", "激活码", "代购", "推广",
}

// 下载类拦截：允许「源码仓库」等合规表述。
var downloadPhrases = []string{
	"下载安装包", "获取安装包", "安装包下载", "exe下载", "exe 下载",
	"网盘地址", "提取码", "下载地址", "下载链接",
}

// Project 校验项目字段，返回首个错误。
func Project(p *model.Project) *FieldError {
	if err := text("name", p.Name, 1, maxTitleRunes, "项目名称需为 1–120 个字符"); err != nil {
		return err
	}
	if err := Slug(p.Slug); err != nil {
		return err
	}
	if err := text("summary", p.Summary, 1, maxSummaryRunes, "摘要需为 1–300 个字符"); err != nil {
		return err
	}
	if len(p.TechStack) > maxTechStackSize {
		return fail("tech_stack", "技术栈最多 10 项，单项不超过 30 个字符")
	}
	for _, item := range p.TechStack {
		item = strings.TrimSpace(item)
		if item == "" || utf8.RuneCountInString(item) > maxTechItemRunes {
			return fail("tech_stack", "技术栈最多 10 项，单项不超过 30 个字符")
		}
	}
	if p.RepoURL != "" && !repoPattern.MatchString(p.RepoURL) {
		return fail("repo_url", "源码地址仅允许 GitHub 仓库地址（https://github.com/owner/repo）")
	}
	if len(p.Description) > maxDescriptionBytes {
		return fail("description", "项目说明不超过 20 KB")
	}
	if p.CoverImage != "" && !validCover(p.CoverImage) {
		return fail("cover_image", "封面图需填站内路径，如 /media/projects/example.png")
	}
	if !p.Status.Valid() {
		return fail("status", "内容状态不合法")
	}
	if p.SortOrder < 0 || p.SortOrder > 9999 {
		return fail("sort_order", "排序值需为 0–9999")
	}
	return scanAll(map[string]string{
		"name":        p.Name,
		"summary":     p.Summary,
		"description": p.Description,
	})
}

// Note 校验笔记字段，返回首个错误。
func Note(n *model.Note) *FieldError {
	if err := text("title", n.Title, 1, maxTitleRunes, "标题需为 1–120 个字符"); err != nil {
		return err
	}
	if err := Slug(n.Slug); err != nil {
		return err
	}
	if err := text("summary", n.Summary, 1, maxSummaryRunes, "摘要需为 1–300 个字符"); err != nil {
		return err
	}
	if strings.TrimSpace(n.Content) == "" || len(n.Content) > maxContentBytes {
		return fail("content", "正文不能为空且不超过 100 KB")
	}
	if !n.Status.Valid() {
		return fail("status", "内容状态不合法")
	}
	if n.Status == model.StatusPublished && n.PublishedAt == nil {
		return fail("published_at", "发布状态必须填写发布时间")
	}
	return scanAll(map[string]string{
		"title":   n.Title,
		"summary": n.Summary,
		"content": n.Content,
	})
}

// Slug 校验链接标识格式。
func Slug(value string) *FieldError {
	if !slugPattern.MatchString(value) {
		return fail("slug", "链接标识只能使用小写字母、数字和连字符，长度 2–80")
	}
	length := utf8.RuneCountInString(value)
	if length < minSlugRunes || length > maxSlugRunes {
		return fail("slug", "链接标识只能使用小写字母、数字和连字符，长度 2–80")
	}
	return nil
}

// Password 校验管理员新密码：至少 12 位且同时包含字母与数字。
func Password(value string) *FieldError {
	if utf8.RuneCountInString(value) < minPasswordRunes ||
		!hasLetter.MatchString(value) || !hasDigit.MatchString(value) {
		return fail("password", "新密码至少 12 位，且同时包含字母和数字")
	}
	return nil
}

// validCover 只放行站内图片路径，并挡掉 .. 之类的目录穿越写法。
func validCover(value string) bool {
	if strings.Contains(value, "..") || strings.Contains(value, "//") {
		return false
	}
	return coverPattern.MatchString(value)
}

func text(field, value string, min, max int, message string) *FieldError {
	length := utf8.RuneCountInString(strings.TrimSpace(value))
	if length < min || length > max {
		return fail(field, message)
	}
	return nil
}

// scanAll 按固定顺序扫描字段，保证同一份内容每次返回的错误一致。
func scanAll(fields map[string]string) *FieldError {
	order := []string{"title", "name", "summary", "content", "description"}
	for _, field := range order {
		value, ok := fields[field]
		if !ok {
			continue
		}
		if err := scanText(field, value); err != nil {
			return err
		}
	}
	return nil
}

// ICPNumber 校验站点备案号；空串表示清空备案号，恢复页脚占位文案。
func ICPNumber(value string) *FieldError {
	if value == "" {
		return nil
	}
	if !icpPattern.MatchString(value) {
		return fail("icp_number", "备案号需为 5–100 个汉字、字母、数字或连字符，例如 甘ICP备12345678号-1")
	}
	return nil
}

func scanText(field, value string) *FieldError {
	lower := strings.ToLower(value)

	var hardHits, downloadHits []string
	for _, word := range hardWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			hardHits = append(hardHits, word)
		}
	}
	for _, phrase := range downloadPhrases {
		if strings.Contains(lower, strings.ToLower(phrase)) {
			downloadHits = append(downloadHits, phrase)
		}
	}

	if len(hardHits) == 0 && len(downloadHits) == 0 {
		return nil
	}

	suggestion := "请删除或改写这些表述"
	if len(downloadHits) > 0 {
		suggestion = "请改为“源码仓库”或删除下载相关表述"
	} else if len(hardHits) > 0 {
		suggestion = "个人非经营性站点不能出现售卖、收费、赞助等表述，请删除"
	}

	hits := append(append([]string{}, hardHits...), downloadHits...)
	return &FieldError{
		Field:      field,
		Message:    "内容包含不符合备案边界的表述",
		Hits:       hits,
		Suggestion: suggestion,
	}
}
