package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// historyRecord is one JSON object in public/content/history.txt.
// The file is JSONL. The latest record for a path is the current revision.
// mtime is the newest source-file modification time in nanoseconds, and size
// is the markdown file size, so an unchanged article can be skipped.
type historyRecord struct {
	PathInput  string            `json:"path_input"`
	PathOutput string            `json:"path_output"`
	Typ        string            `json:"typ"`
	ID         string            `json:"id"`
	Rev        string            `json:"rev"`
	Prv        string            `json:"prv,omitempty"`
	Now        int64             `json:"now"`
	Assets     map[string]string `json:"assets,omitempty"`
	Title      string            `json:"title"`
	Author     []string          `json:"author,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Date       string            `json:"date"`
	Mtime      int64             `json:"mtime"`
	Size       int64             `json:"size"`
}

func loadHistory(path string) ([]historyRecord, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return parseHistory(b)
}

func parseHistory(b []byte) ([]historyRecord, error) {
	var out []historyRecord
	for i, line := range bytes.Split(b, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var r historyRecord
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, fmt.Errorf("history.txt line %d: %w", i+1, err)
		}
		out = append(out, r)
	}
	return out, nil
}

func (a *app) writeHistory() error {
	var b bytes.Buffer
	for _, r := range a.history {
		line, err := json.Marshal(r)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(a.historyPath(), b.Bytes(), 0o644)
}

// findPrev returns the latest history row for this article, matched by the
// directory it lives in now or the directory it lived in before a rename.
func (a *app) findPrev(oldRel string) int {
	idx := -1
	for i := range a.history {
		p := a.history[i].PathInput
		if p == "" {
			continue
		}
		if oldRel != "" && p == oldRel {
			idx = i
		}
	}
	return idx
}

func snapshotName(r historyRecord) string {
	return strconv.FormatInt(r.Now, 10) + "_" + r.Rev + ".html"
}

func historyList(cfg Config, chain []historyRecord, current historyRecord) string {
	if len(chain) == 0 {
		return ""
	}
	recs := append([]historyRecord(nil), chain...)
	sort.SliceStable(recs, func(i, j int) bool {
		return recs[i].Now > recs[j].Now
	})
	var b strings.Builder
	b.WriteString("<ul>\n")
	for _, r := range recs {
		href := rootHref(cfg.URIBasePath, "content/history/"+snapshotName(r))
		when := time.Unix(r.Now, 0).In(time.Local)
		labelDate := formatDate(when.Format("20060102"), cfg.DateFormat)
		text := fmt.Sprintf("%s (%d)", labelDate, r.Now)
		if r.Now == current.Now && r.Rev == current.Rev {
			text = "Current " + text
		}
		fmt.Fprintf(&b, "  <li><a href=\"%s\">%s</a></li>\n", html.EscapeString(href), html.EscapeString(text))
	}
	b.WriteString("</ul>")
	return b.String()
}

type nameIndex struct {
	path string
	have map[string]string
}

func loadNameIndex(path string) (*nameIndex, error) {
	ix := &nameIndex{path: path, have: map[string]string{}}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ix, nil
	}
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		dig, name, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		ix.have[dig] = strings.TrimSpace(name)
	}
	return ix, nil
}

func (ix *nameIndex) add(digest, name string) error {
	if _, ok := ix.have[digest]; ok {
		return nil
	}
	ix.have[digest] = name
	f, err := os.OpenFile(ix.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s %s\n", digest, name)
	return err
}
