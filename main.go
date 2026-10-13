package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyphrme/coz"
)

type source struct {
	mdPath string
	dir    string // empty for a standalone markdown file
}

type app struct {
	cfg       Config
	blog      string
	input     string
	output    string
	alg       coz.HshAlg
	logger    *log.Logger
	logFile   *os.File
	history   []historyRecord
	names     *nameIndex
	key       *coz.Key
	published map[string]bool
}

func main() {
	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := Run(cfg); err != nil {
		log.Fatal(err)
	}
}

// Run builds the blog described by cfg.
func Run(cfg Config) error {
	a, err := newApp(cfg)
	if err != nil {
		return err
	}
	defer a.close()
	return a.run()
}

func newApp(cfg Config) (*app, error) {
	blog, err := cfg.blogDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(blog, 0o755); err != nil {
		return nil, err
	}
	algName := cfg.Digest
	if algName == "" {
		algName = "SHA-256"
	}
	alg := coz.ParseHashAlg(algName)
	if _, err := coz.Hash(alg, []byte("x")); err != nil {
		return nil, fmt.Errorf("digest algorithm %q: %w", algName, err)
	}
	logFile, err := openLog(filepath.Join(blog, "build.log"), cfg.MaxLogSize)
	if err != nil {
		return nil, err
	}
	a := &app{
		cfg:       cfg,
		blog:      blog,
		input:     cfg.resolveDir(blog, cfg.Input, "markdown"),
		output:    cfg.resolveDir(blog, cfg.Output, "public"),
		alg:       alg,
		logger:    log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags),
		logFile:   logFile,
		published: map[string]bool{},
	}
	return a, nil
}

func openLog(path string, max int64) (*os.File, error) {
	if max > 0 {
		if info, err := os.Stat(path); err == nil && info.Size() >= max {
			if err := os.Rename(path, path+".old"); err != nil {
				return nil, err
			}
		}
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

func (a *app) close() {
	if a.logFile != nil {
		a.logFile.Close()
	}
}

func (a *app) logf(format string, args ...any) {
	a.logger.Printf(format, args...)
}

func (a *app) rel(path string) string {
	r, err := filepath.Rel(a.blog, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(r)
}

func (a *app) historyPath() string {
	return filepath.Join(a.output, "content", "history.txt")
}

func (a *app) run() error {
	a.logf("Starting bloggr")
	if err := a.ensure(); err != nil {
		return err
	}
	var err error
	a.history, err = loadHistory(a.historyPath())
	if err != nil {
		return err
	}
	a.names, err = loadNameIndex(filepath.Join(a.output, "content", "index.txt"))
	if err != nil {
		return err
	}
	if err := a.loadKey(); err != nil {
		return err
	}
	if a.cfg.DeleteAndRebuild {
		articleDir := filepath.Join(a.output, "article")
		if err := os.RemoveAll(articleDir); err != nil {
			return err
		}
		if err := os.MkdirAll(articleDir, 0o755); err != nil {
			return err
		}
		a.logf("Deleted %s for rebuild", a.rel(articleDir))
	}
	sources, err := a.discover()
	if err != nil {
		return err
	}
	var arts []*Article
	for _, src := range sources {
		art, err := a.process(src)
		if err != nil {
			var skip skipError
			if errors.As(err, &skip) {
				a.logf("Skipping %s: %s", filepath.Base(src.mdPath), err)
				continue
			}
			return err
		}
		arts = append(arts, art)
	}
	if err := a.writeIndex(arts); err != nil {
		return err
	}
	a.logf("bloggr completed")
	return nil
}

func (a *app) ensure() error {
	dirs := []string{
		a.input,
		a.output,
		filepath.Join(a.output, "article"),
		filepath.Join(a.output, "content"),
		filepath.Join(a.output, "content", "digests"),
		filepath.Join(a.output, "content", "history"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	copies := map[string]string{
		filepath.Join(a.output, "article_header.html"): "default_files/header.html",
		filepath.Join(a.output, "article_footer.html"): "default_files/footer.html",
		filepath.Join(a.output, "index.html"):          "default_files/index.html",
	}
	if a.cfg.StyleCSS {
		copies[filepath.Join(a.output, "style.css")] = "default_files/style.css"
	}
	for dst, src := range copies {
		if err := writeDefault(dst, src); err != nil {
			return err
		}
	}
	for _, name := range []string{"history.txt", "index.txt"} {
		p := filepath.Join(a.output, "content", name)
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(p, nil, 0o644); err != nil {
				return err
			}
		}
	}
	return ensureGitignore(filepath.Join(a.blog, ".gitignore"))
}

func ensureGitignore(path string) error {
	def, err := defaultFiles.ReadFile("default_files/gitignore")
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if len(def) > 0 && def[len(def)-1] != '\n' {
			def = append(def, '\n')
		}
		return os.WriteFile(path, def, 0o644)
	}
	if err != nil {
		return err
	}
	orig := string(b)
	for _, line := range strings.Split(string(def), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if bytesContainsLine(b, line) {
			continue
		}
		if len(b) > 0 && b[len(b)-1] != '\n' {
			b = append(b, '\n')
		}
		b = append(b, []byte(line+"\n")...)
	}
	if string(b) == orig {
		return nil
	}
	return os.WriteFile(path, b, 0o644)
}

func bytesContainsLine(b []byte, line string) bool {
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

func (a *app) discover() ([]source, error) {
	entries, err := os.ReadDir(a.input)
	if err != nil {
		return nil, err
	}
	var out []source
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join(a.input, name)
		if e.IsDir() {
			mds, err := markdownFiles(path)
			if err != nil {
				return nil, err
			}
			switch len(mds) {
			case 0:
				a.logf("Skipping %s: no .md file", name)
			case 1:
				out = append(out, source{mdPath: mds[0], dir: path})
			default:
				a.logf("Skipping %s: %d .md files found (expected 1)", name, len(mds))
			}
			continue
		}
		if name == "index.md" {
			continue
		}
		if strings.HasSuffix(name, ".md") {
			out = append(out, source{mdPath: path})
			continue
		}
		a.logf("Ignoring %s", name)
	}
	return out, nil
}

func markdownFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var mds []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			mds = append(mds, filepath.Join(dir, e.Name()))
		}
	}
	return mds, nil
}

func (a *app) process(src source) (*Article, error) {
	raw, err := os.ReadFile(src.mdPath)
	if err != nil {
		return nil, err
	}
	mt, sz, err := fingerprintSource(src)
	if err != nil {
		return nil, err
	}
	oldRel := ""
	if src.dir != "" {
		oldRel = a.rel(src.dir)
	}
	prev := a.findPrev(oldRel)
	body := stripDigest(string(raw))
	art, err := a.finishParse(body)
	if err != nil {
		return nil, err
	}
	if !a.cfg.DeleteAndRebuild && prev >= 0 && a.history[prev].Mtime == mt && a.history[prev].Size == sz && art.Canon != "" {
		outHTML := filepath.Join(a.output, "article", art.Canon, "index.html")
		if _, err := os.Stat(outHTML); err == nil {
			if err := a.claim(art); err != nil {
				return nil, err
			}
			a.logf("Unchanged %s", art.Canon)
			return art, nil
		}
	}
	var inserts []string
	if art.Date == "" {
		if !a.cfg.InsertDate {
			return nil, skipf("no date")
		}
		art.Date = time.Now().Format("20060102")
		art.Canon = canonName(art.Date, art.Title, a.cfg.TitleDelimiter)
		inserts = append(inserts, "## Date: "+art.Date)
	}
	if len(art.Authors) == 0 && a.cfg.InsertAuthor {
		name := a.cfg.DefaultAuthor
		if name == "" {
			name = "Anonymous"
		}
		art.Authors = []string{name}
		inserts = append(inserts, "## Author: "+name)
	}
	if len(inserts) > 0 {
		body = insertAfterTitle(body, inserts)
	}
	if err := a.claim(art); err != nil {
		return nil, err
	}

	var assets []assetFile
	if src.dir != "" {
		assets, err = a.collectAssets(src.dir, src.mdPath)
		if err != nil {
			return nil, err
		}
		body = applyAssetNames(body, assets)
	}
	body = canonicalBody(body)

	wantHash := a.cfg.HashArticle || a.cfg.History || a.cfg.SignArticle
	if wantHash {
		sum, err := a.digestBytes([]byte(body))
		if err != nil {
			return nil, err
		}
		art.Digest = sum
		art.DigestAlg, art.DigestValue = splitDigest(sum)
		base := filepath.Base(src.mdPath)
		if err := a.saveDigest(sum, []byte(body), base); err != nil {
			return nil, err
		}
	}
	fileText := body
	if a.cfg.HashArticle && art.Digest != "" {
		fileText = withDigestSection(body, art.Digest)
	}

	destDir, err := a.writeSource(src, art.Canon, fileText, assets)
	if err != nil {
		return nil, err
	}
	for _, as := range assets {
		if as.digest == "" {
			continue
		}
		if err := a.saveDigest(as.digest, as.data, as.orig); err != nil {
			return nil, err
		}
	}
	mt, sz, err = fingerprint(destDir, filepath.Join(destDir, "index.md"))
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	id := art.Digest
	if id == "" {
		id = "path:" + art.Canon
	}
	rev := id
	rec := historyRecord{
		PathInput:  a.rel(destDir),
		PathOutput: a.rel(filepath.Join(a.output, "article", art.Canon)),
		Typ:        "article",
		ID:         id,
		Rev:        rev,
		Now:        now,
		Assets:     assetMap(assets),
		Title:      art.Title,
		Author:     art.Authors,
		Tags:       art.Tags,
		Date:       art.Date,
		Mtime:      mt,
		Size:       sz,
	}
	sameRev := false
	if prev >= 0 {
		old := a.history[prev]
		if old.ID != "" {
			rec.ID = old.ID
		}
		if old.Rev == rev {
			sameRev = true
			rec.Now = old.Now
			rec.Prv = old.Prv
		} else if old.Rev != "" {
			rec.Prv = old.Rev
		}
		if old.PathOutput != "" && old.PathOutput != rec.PathOutput {
			oldOut := filepath.Join(a.blog, filepath.FromSlash(old.PathOutput))
			if under(oldOut, filepath.Join(a.output, "article")) {
				if err := os.RemoveAll(oldOut); err != nil {
					return nil, err
				}
			}
		}
	}
	if a.cfg.SignArticle {
		if art.DigestValue == "" {
			return nil, fmt.Errorf("sign %s: article has no digest", art.Canon)
		}
		// Sign with the record timestamp so a rewritten article keeps one signature.
		cozJSON, err := a.signArticle(art.DigestValue, rec.Now)
		if err != nil {
			return nil, err
		}
		art.Coz = cozJSON
	}

	var chain []historyRecord
	if a.cfg.History {
		for _, r := range a.history {
			if r.ID == rec.ID {
				chain = append(chain, r)
			}
		}
		if sameRev {
			a.history[prev] = rec
			for i := range chain {
				if chain[i].Rev == rec.Rev {
					chain[i] = rec
					break
				}
			}
		} else {
			chain = append(chain, rec)
			a.history = append(a.history, rec)
		}
		art.HistoryHTML = historyList(a.cfg, chain, rec)
	}

	html, err := a.render(art, body)
	if err != nil {
		return nil, err
	}
	outDir := filepath.Join(a.output, "article", art.Canon)
	if err := os.RemoveAll(outDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(outDir, "index.html"), html, 0o644); err != nil {
		return nil, err
	}
	for _, as := range assets {
		if err := os.WriteFile(filepath.Join(outDir, as.fileName), as.data, 0o644); err != nil {
			return nil, err
		}
	}
	if a.cfg.History {
		snap := filepath.Join(a.output, "content", "history", snapshotName(rec))
		if err := writeIfChanged(snap, html); err != nil {
			return nil, err
		}
		if err := a.writeHistory(); err != nil {
			return nil, err
		}
	}
	a.logf("Published %s", art.Canon)
	return art, nil
}

func (a *app) finishParse(raw string) (*Article, error) {
	art, err := parseArticle(raw)
	if err != nil {
		return nil, err
	}
	if art.Date == "" {
		return art, nil
	}
	art.Canon = canonName(art.Date, art.Title, a.cfg.TitleDelimiter)
	return art, nil
}

func (a *app) claim(art *Article) error {
	if art.Canon == "" {
		return skipf("no date")
	}
	if a.published[art.Canon] {
		return skipf("duplicate article %s", art.Canon)
	}
	a.published[art.Canon] = true
	return nil
}

func (a *app) saveDigest(digest string, data []byte, name string) error {
	if digest == "" {
		return nil
	}
	path := filepath.Join(a.output, "content", "digests", digest)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return a.names.add(digest, name)
}

func (a *app) writeSource(src source, canon, content string, assets []assetFile) (string, error) {
	destDir := filepath.Join(a.input, canon)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	destMD := filepath.Join(destDir, "index.md")
	if err := writeIfChanged(destMD, []byte(content)); err != nil {
		return "", err
	}
	fromDir := src.dir
	if fromDir == "" {
		fromDir = a.input
	}
	for _, as := range assets {
		from := filepath.Join(fromDir, as.fileName)
		to := filepath.Join(destDir, as.fileName)
		if from == to {
			continue
		}
		if err := os.Remove(to); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if err := os.Rename(from, to); err != nil {
			return "", err
		}
	}
	if src.mdPath != destMD {
		if err := os.Remove(src.mdPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	if src.dir != "" && src.dir != destDir {
		removeIfEmpty(src.dir)
	}
	return destDir, nil
}

func writeIfChanged(path string, data []byte) error {
	old, err := os.ReadFile(path)
	if err == nil && string(old) == string(data) {
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}

func fingerprintSource(src source) (int64, int64, error) {
	if src.dir == "" {
		info, err := os.Stat(src.mdPath)
		if err != nil {
			return 0, 0, err
		}
		return info.ModTime().UnixNano(), info.Size(), nil
	}
	return fingerprint(src.dir, src.mdPath)
}

func fingerprint(dir, mdPath string) (int64, int64, error) {
	info, err := os.Stat(mdPath)
	if err != nil {
		return 0, 0, err
	}
	mtime := info.ModTime().UnixNano()
	size := info.Size()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		inf, err := e.Info()
		if err != nil {
			return 0, 0, err
		}
		if inf.ModTime().UnixNano() > mtime {
			mtime = inf.ModTime().UnixNano()
		}
	}
	return mtime, size, nil
}

func removeIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	os.Remove(dir)
}

func under(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..")
}
