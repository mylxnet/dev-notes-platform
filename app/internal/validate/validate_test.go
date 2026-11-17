package validate

import (
	"strings"
	"testing"

	"dev-notes-platform/internal/model"
)

func validProject() *model.Project {
	return &model.Project{
		Name:      "SiteGuard",
		Slug:      "siteguard",
		Summary:   "站点巡检工具",
		TechStack: []string{"Go", "SQLite"},
		RepoURL:   "https://github.com/example/siteguard",
		Status:    model.StatusPublished,
	}
}

func TestProjectValid(t *testing.T) {
	if err := Project(validProject()); err != nil {
		t.Fatalf("expected valid project, got %+v", err)
	}
}

func TestProjectRejectsBadSlug(t *testing.T) {
	project := validProject()
	project.Slug = "Site_Guard"
	err := Project(project)
	if err == nil || err.Field != "slug" {
		t.Fatalf("expected slug error, got %+v", err)
	}
}

func TestProjectRejectsNonGitHubRepo(t *testing.T) {
	for _, raw := range []string{
		"http://github.com/example/siteguard",
		"https://gitlab.com/example/siteguard",
		"javascript:alert(1)",
	} {
		project := validProject()
		project.RepoURL = raw
		err := Project(project)
		if err == nil || err.Field != "repo_url" {
			t.Fatalf("%s: expected repo_url error, got %+v", raw, err)
		}
	}
}

func TestProjectRejectsTooManyTechItems(t *testing.T) {
	project := validProject()
	project.TechStack = make([]string, 11)
	for i := range project.TechStack {
		project.TechStack[i] = "Go"
	}
	err := Project(project)
	if err == nil || err.Field != "tech_stack" {
		t.Fatalf("expected tech_stack error, got %+v", err)
	}
}

func TestProjectBlocksCommercialWords(t *testing.T) {
	project := validProject()
	project.Summary = "支持在线支付与会员订阅"
	err := Project(project)
	if err == nil {
		t.Fatal("expected content blocked")
	}
	if err.Field != "summary" {
		t.Fatalf("expected summary field, got %q", err.Field)
	}
	if len(err.Hits) == 0 || err.Suggestion == "" {
		t.Fatalf("expected hits and suggestion, got %+v", err)
	}
}

func TestNoteBlocksDownloadPhrases(t *testing.T) {
	note := &model.Note{
		Title:   "打包说明",
		Slug:    "package-guide",
		Summary: "摘要",
		Content: "完整版请从网盘地址获取安装包",
		Status:  model.StatusDraft,
	}
	err := Note(note)
	if err == nil || err.Field != "content" {
		t.Fatalf("expected content blocked, got %+v", err)
	}
	if !strings.Contains(err.Suggestion, "源码仓库") {
		t.Fatalf("expected download suggestion, got %q", err.Suggestion)
	}
}

func TestNoteAllowsSourceRepoWording(t *testing.T) {
	note := &model.Note{
		Title:   "源码说明",
		Slug:    "source-note",
		Summary: "摘要",
		Content: "源码仓库地址在项目详情页，可查看源码。",
		Status:  model.StatusDraft,
	}
	if err := Note(note); err != nil {
		t.Fatalf("expected allowed content, got %+v", err)
	}
}

func TestNoteRequiresPublishedAt(t *testing.T) {
	note := &model.Note{
		Title:   "标题",
		Slug:    "some-note",
		Summary: "摘要",
		Content: "正文",
		Status:  model.StatusPublished,
	}
	err := Note(note)
	if err == nil || err.Field != "published_at" {
		t.Fatalf("expected published_at error, got %+v", err)
	}
}

func TestNoteRejectsOversizeContent(t *testing.T) {
	note := &model.Note{
		Title:   "标题",
		Slug:    "big-note",
		Summary: "摘要",
		Content: strings.Repeat("a", maxContentBytes+1),
		Status:  model.StatusDraft,
	}
	err := Note(note)
	if err == nil || err.Field != "content" {
		t.Fatalf("expected content size error, got %+v", err)
	}
}

func TestPasswordRules(t *testing.T) {
	cases := []struct {
		value string
		ok    bool
	}{
		{"admin12345678", true},
		{"short1", false},
		{"alllettersonly", false},
		{"123456789012", false},
	}
	for _, tc := range cases {
		err := Password(tc.value)
		if tc.ok && err != nil {
			t.Fatalf("%s: expected ok, got %+v", tc.value, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s: expected error", tc.value)
		}
	}
}
