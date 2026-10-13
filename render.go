package main

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown"
)

//go:embed default_files/header.html default_files/footer.html default_files/style.css default_files/index.html default_files/gitignore
var defaultFiles embed.FS

const (
	articleListStart = "<!-- Start Article List -->"
	articleListEnd   = "<!-- End Article List -->"
)

func writeDefault(dst, embedPath string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	b, err := defaultFiles.ReadFile(embedPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func rootHref(base, rel string) string {
	rel = strings.TrimPrefix(rel, "/")
	if base == "" {
		return "/" + rel
	}
	return strings.TrimRight(base, "/") + "/" + rel
}

func (a *app) articleHref(canon string) string {
	rel := "article/" + canon + "/index.html"
	if a.cfg.URIBasePath == "" {
		return rel
	}
	return strings.TrimRight(a.cfg.URIBasePath, "/") + "/" + rel
}

func listText(art *Article, cfg Config) (label, meta string) {
	var b strings.Builder
	if cfg.ArticleListShowDate {
		b.WriteString("[" + formatDate(art.Date, cfg.DateFormat) + "] ")
	}
	b.WriteString(art.Title)
	if art.Subtitle != "" {
		b.WriteString(" - ")
		b.WriteString(art.Subtitle)
	}
	label = b.String()
	var parts []string
	if cfg.ArticleListShowAuthor && len(art.Authors) > 0 {
		parts = append(parts, "Author: "+strings.Join(art.Authors, ", "))
	}
	if cfg.ArticleListShowTags && len(art.Tags) > 0 {
		parts = append(parts, "Tags: "+strings.Join(art.Tags, ", "))
	}
	if len(parts) > 0 {
		meta = "(" + strings.Join(parts, ", ") + ")"
	}
	return label, meta
}

func applyVars(tmpl string, art *Article, cfg Config) string {
	return strings.NewReplacer(
		"{URI_BASE_PATH}", cfg.URIBasePath,
		"{ARTICLE_TITLE}", html.EscapeString(art.Title),
		"{ARTICLE_AUTHOR}", html.EscapeString(strings.Join(art.Authors, ", ")),
		"{ARTICLE_DATE}", html.EscapeString(formatDate(art.Date, cfg.DateFormat)),
		"{ARTICLE_DATE_YYYYMMDD}", html.EscapeString(art.Date),
		"{ARTICLE_TAGS}", html.EscapeString(strings.Join(art.Tags, ", ")),
		"{ARTICLE_DIGEST}", html.EscapeString(art.Digest),
		"{ARTICLE_DIGEST_VALUE}", html.EscapeString(art.DigestValue),
		"{ARTICLE_DIGEST_ALG}", html.EscapeString(art.DigestAlg),
		"{ARTICLE_COZ}", art.Coz,
		"{ARTICLE_HISTORY}", art.HistoryHTML,
	).Replace(tmpl)
}

func (a *app) render(art *Article, body string) ([]byte, error) {
	header, err := os.ReadFile(filepath.Join(a.output, "article_header.html"))
	if err != nil {
		return nil, err
	}
	footer, err := os.ReadFile(filepath.Join(a.output, "article_footer.html"))
	if err != nil {
		return nil, err
	}
	htmlBody := markdown.ToHTML([]byte(dropFirstHeader(body)), nil, nil)
	h := applyVars(string(header), art, a.cfg)
	f := applyVars(string(footer), art, a.cfg)
	var buf bytes.Buffer
	buf.WriteString(h)
	if !strings.HasSuffix(h, "\n") {
		buf.WriteByte('\n')
	}
	buf.Write(htmlBody)
	if len(htmlBody) == 0 || htmlBody[len(htmlBody)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString(f)
	if !strings.HasSuffix(f, "\n") {
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func (a *app) writeIndex(arts []*Article) error {
	sortArticles(arts)
	var md, list bytes.Buffer
	for _, art := range arts {
		label, meta := listText(art, a.cfg)
		if meta != "" {
			fmt.Fprintf(&md, "%s %s\n", label, meta)
		} else {
			fmt.Fprintf(&md, "%s\n", label)
		}
		href := a.articleHref(art.Canon)
		if meta != "" {
			fmt.Fprintf(&list, "<p><a href=\"%s\">%s</a> <span class=\"article-meta\">%s</span></p>\n",
				html.EscapeString(href), html.EscapeString(label), html.EscapeString(meta))
		} else {
			fmt.Fprintf(&list, "<p><a href=\"%s\">%s</a></p>\n", html.EscapeString(href), html.EscapeString(label))
		}
	}
	indexMD := filepath.Join(a.input, "index.md")
	if err := os.WriteFile(indexMD, md.Bytes(), 0o644); err != nil {
		return err
	}
	indexHTML := filepath.Join(a.output, "index.html")
	page, err := os.ReadFile(indexHTML)
	if err != nil {
		return err
	}
	updated, err := spliceArticleList(string(page), list.String())
	if err != nil {
		return err
	}
	return os.WriteFile(indexHTML, []byte(updated), 0o644)
}

func spliceArticleList(page, list string) (string, error) {
	start := strings.Index(page, articleListStart)
	end := strings.Index(page, articleListEnd)
	if start < 0 || end < 0 || end < start {
		return "", fmt.Errorf("public/index.html is missing Article List markers")
	}
	var b strings.Builder
	b.WriteString(page[:start+len(articleListStart)])
	b.WriteByte('\n')
	b.WriteString(list)
	b.WriteString(page[end:])
	return b.String(), nil
}

func sortArticles(arts []*Article) {
	sort.Slice(arts, func(i, j int) bool {
		return articleLess(arts[i], arts[j])
	})
}

func articleLess(a, b *Article) bool {
	if a.Date != b.Date {
		return a.Date > b.Date
	}
	return a.Title < b.Title
}
