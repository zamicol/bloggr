package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cyphrme/coz"
)

func TestParseMetadata(t *testing.T) {
	gravity, err := os.ReadFile("test/source_files_for_testing/gravity_wells/gravity.md")
	if err != nil {
		t.Fatal(err)
	}
	art, err := parseArticle(string(gravity))
	if err != nil {
		t.Fatal(err)
	}
	if art.Title != "Gravity Wells are Sombreros" {
		t.Fatalf("title %q", art.Title)
	}
	if art.Date != "20180921" {
		t.Fatalf("date %q", art.Date)
	}
	if strings.Join(art.Authors, ",") != "Zamicol" {
		t.Fatalf("authors %v", art.Authors)
	}
	if strings.Join(art.Tags, "|") != "Gravity|Special Relativity|Time Dilation" {
		t.Fatalf("tags %v", art.Tags)
	}
	if art.Subtitle != "Original post published Friday, September 21, 2018.  Modified 20260925" {
		t.Fatalf("subtitle %q", art.Subtitle)
	}

	intro, err := os.ReadFile("test/source_files_for_testing/introduction_to_coz/introduction to coz.md")
	if err != nil {
		t.Fatal(err)
	}
	art, err = parseArticle(string(intro))
	if err != nil {
		t.Fatal(err)
	}
	if art.Date != "20260921" {
		t.Fatalf("unix date %q", art.Date)
	}
	if art.Title != "Introduction to Coz" {
		t.Fatalf("title %q", art.Title)
	}

	raw := "# Shared Line\n#### Date: 20250101, Author: Zamicol, Alice\n#### Tags: Example Tag 1, Example Tag 2, Cryptography\n\nBody\n"
	art, err = parseArticle(raw)
	if err != nil {
		t.Fatal(err)
	}
	if art.Date != "20250101" {
		t.Fatalf("date %q", art.Date)
	}
	if strings.Join(art.Authors, "|") != "Zamicol|Alice" {
		t.Fatalf("authors %v", art.Authors)
	}
	if strings.Join(art.Tags, "|") != "Example Tag 1|Example Tag 2|Cryptography" {
		t.Fatalf("tags %v", art.Tags)
	}

	footer := "# Title\n## Date: 2020/01/02\n\nBody\n\n---\n## Author: Late\n"
	art, err = parseArticle(footer)
	if err != nil {
		t.Fatal(err)
	}
	if art.Date != "20200102" || strings.Join(art.Authors, "") != "Late" {
		t.Fatalf("footer meta date %q authors %v", art.Date, art.Authors)
	}

	if _, err := parseArticle("no title here\n"); err == nil {
		t.Fatal("expected missing title to skip")
	}
}

func TestSlugDateAndDigestSection(t *testing.T) {
	if got := canonName("20261012", "Hello World", "_"); got != "20261012_Hello_World" {
		t.Fatalf("canon %q", got)
	}
	if got := canonName("20180921", "Gravity Wells are Sombreros", "_"); got != "20180921_Gravity_Wells_are_Sombreros" {
		t.Fatalf("canon %q", got)
	}
	if got := formatDate("20260921", "YYYY/MM/DD"); got != "2026/09/21" {
		t.Fatalf("format %q", got)
	}
	if got := stampedName("image1.png", "SHA-256:abc"); got != "image1_SHA-256:abc.png" {
		t.Fatalf("stamp %q", got)
	}
	if got := stampedName("image1_SHA-256:old.png", "SHA-256:new"); got != "image1_SHA-256:new.png" {
		t.Fatalf("restamp %q", got)
	}

	body := canonicalBody("# Hello\n## Date: 20200101\n\nHi\n")
	file := withDigestSection(body, "SHA-256:abc")
	if canonicalBody(file) != body {
		t.Fatalf("digest section did not round-trip\n got %q\nwant %q", canonicalBody(file), body)
	}
	content := "see image.png and image_SHA-256:abc.png"
	got := replaceFilename(content, "image.png", "image_SHA-256:zzz.png")
	want := "see image_SHA-256:zzz.png and image_SHA-256:abc.png"
	if got != want {
		t.Fatalf("replace %q", got)
	}
}

func TestPipelineSamples(t *testing.T) {
	blog := t.TempDir()
	md := filepath.Join(blog, "markdown")
	if err := os.MkdirAll(md, 0o755); err != nil {
		t.Fatal(err)
	}
	copyDir(t, "test/source_files_for_testing/gravity_wells", filepath.Join(md, "gravity_wells"))
	copyDir(t, "test/source_files_for_testing/introduction_to_coz", filepath.Join(md, "introduction_to_coz"))
	hello := "# Hello World\n## Date: 20200101\n\nThis is my first article.\n"
	if err := os.WriteFile(filepath.Join(md, "hello.md"), []byte(hello), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(md, "bad"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(md, "bad", "a.md"), []byte("# A\n## Date: 20200101\n\n"), 0o644)
	os.WriteFile(filepath.Join(md, "bad", "b.md"), []byte("# B\n## Date: 20200101\n\n"), 0o644)
	os.WriteFile(filepath.Join(md, "notitle.md"), []byte("just words\n"), 0o644)
	os.WriteFile(filepath.Join(md, "notes.txt"), []byte("ignore me\n"), 0o644)

	cfg := DefaultConfig()
	cfg.Path = blog
	cfg.DateFormat = "YYYY/MM/DD"
	if err := Run(cfg); err != nil {
		t.Fatal(err)
	}

	gravityDir := filepath.Join(md, "20180921_Gravity_Wells_are_Sombreros")
	introDir := filepath.Join(md, "20260921_Introduction_to_Coz")
	helloDir := filepath.Join(md, "20200101_Hello_World")
	for _, dir := range []string{gravityDir, introDir, helloDir} {
		if _, err := os.Stat(filepath.Join(dir, "index.md")); err != nil {
			t.Fatalf("missing %s: %v", dir, err)
		}
	}
	if _, err := os.Stat(filepath.Join(md, "hello.md")); !os.IsNotExist(err) {
		t.Fatal("standalone markdown file was not moved into its article directory")
	}
	if _, err := os.Stat(filepath.Join(md, "gravity_wells", "original_cyphrme", "gravity.md")); err != nil {
		t.Fatal("nested source directory should stay put")
	}
	gbytes, err := os.ReadFile(filepath.Join(gravityDir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	gtext := string(gbytes)
	if !strings.Contains(gtext, "###### Article Digest\nSHA-256:") {
		t.Fatalf("missing digest section:\n%s", gtext)
	}
	if strings.Contains(gtext, "](gravity_dip.png)") {
		t.Fatal("image reference was not digest-stamped")
	}
	matches, err := filepath.Glob(filepath.Join(gravityDir, "gravity_dip_SHA-256:*.png"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("stamped image matches %v err %v", matches, err)
	}
	img, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	sum, err := digestFile(img)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filepath.Base(matches[0]), sum) {
		t.Fatalf("stamp %q does not contain %s", filepath.Base(matches[0]), sum)
	}
	stored, err := os.ReadFile(filepath.Join(blog, "public", "content", "digests", sum))
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != string(img) {
		t.Fatal("digest store does not match image bytes")
	}

	intro, err := os.ReadFile(filepath.Join(introDir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(intro), "## Author: Anonymous") {
		t.Fatal("missing default author")
	}
	if _, err := os.Stat(filepath.Join(blog, "public", "article", "introduction_to_coz")); !os.IsNotExist(err) {
		t.Fatal("output used the input directory name instead of the dated title")
	}
	outArticle := filepath.Join(blog, "public", "article", "20180921_Gravity_Wells_are_Sombreros", "index.html")
	page, err := os.ReadFile(outArticle)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "<h1>Gravity Wells are Sombreros</h1>") {
		t.Fatal("header title was not substituted")
	}
	if !strings.Contains(string(page), "gravity_dip_SHA-256:") {
		t.Fatal("article HTML is missing the stamped image")
	}
	introImages, err := filepath.Glob(filepath.Join(blog, "public", "article", "20260921_Introduction_to_Coz", "*.png"))
	if err != nil || len(introImages) != 3 {
		t.Fatalf("unreferenced images were not copied: %v %v", introImages, err)
	}
	if !strings.Contains(string(page), "bloggr/sign") {
		t.Fatal("missing coz signature")
	}
	if !strings.Contains(string(page), "Current ") {
		t.Fatal("missing history list")
	}
	verifyArticleSignature(t, blog, string(page))

	index, err := os.ReadFile(filepath.Join(blog, "public", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	indexText := string(index)
	introAt := strings.Index(indexText, "20260921_Introduction_to_Coz")
	gravityAt := strings.Index(indexText, "20180921_Gravity_Wells_are_Sombreros")
	helloAt := strings.Index(indexText, "20200101_Hello_World")
	if introAt < 0 || gravityAt < 0 || helloAt < 0 {
		t.Fatalf("index missing articles:\n%s", indexText)
	}
	// 20260921, then 20200101, then 20180921.
	if !(introAt < helloAt && helloAt < gravityAt) {
		t.Fatal("article list is not newest first")
	}
	if !strings.Contains(indexText, "[2026/09/21] Introduction to Coz") {
		t.Fatal("index date was not formatted")
	}
	if !strings.Contains(indexText, "(Tags: Gravity, Special Relativity, Time Dilation)") {
		t.Fatal("index tags missing")
	}
	if strings.Contains(indexText, "Author:") {
		t.Fatal("author is hidden unless ArticleListShowAuthor is set")
	}
	names, err := os.ReadFile(filepath.Join(blog, "public", "content", "index.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(names), " gravity_dip.png") {
		t.Fatalf("index.txt:\n%s", names)
	}
	hist, err := os.ReadFile(filepath.Join(blog, "public", "content", "history.txt"))
	if err != nil {
		t.Fatal(err)
	}
	recs := historyRecords(t, hist)
	if len(recs) != 3 {
		t.Fatalf("history records %d", len(recs))
	}
	snaps, err := filepath.Glob(filepath.Join(blog, "public", "content", "history", "*.html"))
	if err != nil || len(snaps) != 3 {
		t.Fatalf("snapshots %v %v", snaps, err)
	}
	logText, err := os.ReadFile(filepath.Join(blog, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logText), "Skipping bad: 2 .md files found") {
		t.Fatalf("log missing multi-md skip:\n%s", logText)
	}
	if !strings.Contains(string(logText), "Skipping notitle.md: no title found") {
		t.Fatalf("log missing title skip:\n%s", logText)
	}
	if !strings.Contains(string(logText), "Skipping nested directory") {
		t.Fatalf("log missing nested skip:\n%s", logText)
	}

	before := append([]byte(nil), gbytes...)
	if err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(gravityDir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("second run rewrote an unchanged article")
	}
	hist, err = os.ReadFile(filepath.Join(blog, "public", "content", "history.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(historyRecords(t, hist)) != 3 {
		t.Fatal("second run appended a revision for unchanged articles")
	}
	logText, err = os.ReadFile(filepath.Join(blog, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logText), "Unchanged 20180921_Gravity_Wells_are_Sombreros") {
		t.Fatal("second run did not skip the unchanged article")
	}
	indexText = string(mustRead(t, filepath.Join(blog, "public", "index.html")))
	if strings.Count(indexText, "20180921_Gravity_Wells_are_Sombreros") != 1 {
		t.Fatal("second run duplicated the article list")
	}
	helloHTML := filepath.Join(blog, "public", "article", "20200101_Hello_World", "index.html")
	if err := os.Remove(helloHTML); err != nil {
		t.Fatal(err)
	}
	if err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(helloHTML); err != nil {
		t.Fatal("deleting an article page did not rebuild it")
	}
	if len(historyRecords(t, mustRead(t, filepath.Join(blog, "public", "content", "history.txt")))) != 3 {
		t.Fatal("rebuilding a missing page added a revision")
	}

	edited := strings.Replace(string(after), "\n---\n###### Article Digest\n", "\nEdited.\n\n---\n###### Article Digest\n", 1)
	if edited == string(after) {
		t.Fatal("could not find digest section to edit")
	}
	if err := os.WriteFile(filepath.Join(gravityDir, "index.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	// The editor's write bumps mtime. Make the change obvious if the clock is coarse.
	future := time.Now().Add(2 * time.Second)
	os.Chtimes(filepath.Join(gravityDir, "index.md"), future, future)
	if err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	recs = historyRecords(t, mustRead(t, filepath.Join(blog, "public", "content", "history.txt")))
	var gravity []historyRecord
	for _, r := range recs {
		if r.Title == "Gravity Wells are Sombreros" {
			gravity = append(gravity, r)
		}
	}
	if len(gravity) != 2 {
		t.Fatalf("gravity revisions %d", len(gravity))
	}
	if gravity[0].ID == "" || gravity[0].ID != gravity[1].ID {
		t.Fatalf("id changed across edit: %#v %#v", gravity[0], gravity[1])
	}
	if gravity[1].Prv != gravity[0].Rev || gravity[1].Rev == gravity[0].Rev {
		t.Fatalf("revision chain prv %s rev %s previous %s", gravity[1].Prv, gravity[1].Rev, gravity[0].Rev)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.History || !cfg.HashArticle || cfg.ArticleListShowAuthor {
		t.Fatalf("defaults %+v", cfg)
	}
	if cfg.Input != "markdown" || cfg.Output != "public" || cfg.Digest != "SHA-256" {
		t.Fatalf("paths %+v", cfg)
	}
	partial := []byte(`{"ArticleListShowAuthor": true, "SignArticle": false}`)
	if err := os.WriteFile(path, partial, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.ArticleListShowAuthor || cfg.SignArticle || !cfg.InsertDate {
		t.Fatalf("partial overlay %+v", cfg)
	}
}

func digestFile(b []byte) (string, error) {
	sum, err := coz.Hash(coz.SHA256, b)
	if err != nil {
		return "", err
	}
	return "SHA-256:" + sum.String(), nil
}

func historyRecords(t *testing.T, b []byte) []historyRecord {
	t.Helper()
	recs, err := parseHistory(b)
	if err != nil {
		t.Fatal(err)
	}
	return recs
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func verifyArticleSignature(t *testing.T, blog, page string) {
	t.Helper()
	i := strings.Index(page, "using Coz:")
	if i < 0 {
		t.Fatal("signature section missing")
	}
	j := strings.Index(page[i:], "{")
	if j < 0 {
		t.Fatal("signature json missing")
	}
	var cz struct {
		Pay json.RawMessage `json:"pay"`
		Sig string          `json:"sig"`
	}
	dec := json.NewDecoder(strings.NewReader(page[i+j:]))
	if err := dec.Decode(&cz); err != nil {
		t.Fatal(err)
	}
	var pay struct {
		Article string `json:"article"`
		Alg     string `json:"alg"`
		Typ     string `json:"typ"`
	}
	if err := json.Unmarshal(cz.Pay, &pay); err != nil {
		t.Fatal(err)
	}
	if pay.Typ != "bloggr/sign" || pay.Alg != "ES256" || pay.Article == "" {
		t.Fatalf("pay %+v", pay)
	}
	keyBytes, err := os.ReadFile(filepath.Join(blog, "coz_key.json"))
	if err != nil {
		t.Fatal(err)
	}
	var key coz.Key
	if err := json.Unmarshal(keyBytes, &key); err != nil {
		t.Fatal(err)
	}
	sig, err := coz.Decode(cz.Sig)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := key.VerifyCoz(&coz.Coz{Pay: cz.Pay, Sig: sig})
	if err != nil || !ok {
		t.Fatalf("signature did not verify ok=%v err=%v", ok, err)
	}
	info, err := os.Stat(filepath.Join(blog, "coz_key.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("key mode %v", info.Mode().Perm())
	}
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
