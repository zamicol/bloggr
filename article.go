package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// digestTail is the article digest section appended after image names are
// rewritten. It is not part of the hashed body: the section names that hash.
var digestTail = regexp.MustCompile(`\n---\n#{1,6} Article Digest\n\S+\n?$`)

var headerLine = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// labelRE splits a metadata line on Label: fields. The value runs until the
// next label, so "Author: Zamicol, Alice" stays two authors when Date and
// Tags share the line.
var labelRE = regexp.MustCompile(`(?:^|,)\s*([A-Za-z][A-Za-z0-9]*):\s*`)

// stampRE matches a digest stamp in a file base: name_SHA-256:<b64>.
var stampRE = regexp.MustCompile(`^(.+)_([A-Za-z0-9]+(?:-[0-9]+)?):([A-Za-z0-9_-]+)$`)

type skipError struct{ msg string }

func (e skipError) Error() string { return e.msg }

func skipf(format string, args ...any) error {
	return skipError{msg: fmt.Sprintf(format, args...)}
}

// Article is one markdown article after metadata is parsed.
type Article struct {
	Title       string
	Subtitle    string
	Date        string // YYYYMMDD
	Authors     []string
	Tags        []string
	Canon       string
	Digest      string // alg:b64
	DigestAlg   string
	DigestValue string
	Coz         string
	HistoryHTML string
	dateRaw     string
}

func stripDigest(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if loc := digestTail.FindStringIndex(s); loc != nil {
		return s[:loc[0]]
	}
	return s
}

// canonicalBody is the exact byte string that is hashed. A following run
// strips the digest section and gets this string back.
func canonicalBody(s string) string {
	return strings.TrimRight(stripDigest(s), "\n") + "\n"
}

func withDigestSection(body, digest string) string {
	body = strings.TrimRight(stripDigest(body), "\n") + "\n"
	return body + "\n---\n###### Article Digest\n" + digest + "\n"
}

func headerText(trim string) (string, bool) {
	m := headerLine.FindStringSubmatch(trim)
	if m == nil {
		return "", false
	}
	return m[2], true
}

func parseLabelValues(line string) map[string]string {
	idx := labelRE.FindAllStringSubmatchIndex(line, -1)
	if len(idx) == 0 || idx[0][0] != 0 {
		return nil
	}
	out := make(map[string]string, len(idx))
	for i, m := range idx {
		name := line[m[2]:m[3]]
		valEnd := len(line)
		if i+1 < len(idx) {
			valEnd = idx[i+1][0]
		}
		val := strings.TrimSpace(line[m[1]:valEnd])
		val = strings.TrimSuffix(val, ",")
		val = strings.TrimSpace(val)
		out[name] = val
	}
	return out
}

func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func applyLabels(text string, art *Article) bool {
	labels := parseLabelValues(text)
	if len(labels) == 0 {
		return false
	}
	if v, ok := labels["Date"]; ok {
		art.dateRaw = v
	}
	if v, ok := labels["Author"]; ok {
		art.Authors = splitList(v)
	}
	if v, ok := labels["Tags"]; ok {
		art.Tags = splitList(v)
	}
	return true
}

const (
	modeSeek = iota
	modeHeader
	modeBody
	modeFooter
)

func nextIsHeader(lines []string, i int) bool {
	if i+1 >= len(lines) {
		return false
	}
	_, ok := headerText(strings.TrimSpace(lines[i+1]))
	return ok
}

// parseArticle reads title, date, authors, tags, and the first subtitle.
// Header metadata runs from the title until a blank line or a non-header line.
// A "---" line followed by a header enters footer metadata.
func parseArticle(raw string) (*Article, error) {
	raw = stripDigest(raw)
	lines := strings.Split(raw, "\n")
	art := &Article{}
	mode := modeSeek
	for i := 0; i < len(lines); i++ {
		trim := strings.TrimSpace(lines[i])
		if mode == modeSeek {
			if trim == "" {
				continue
			}
			if text, ok := headerText(trim); ok {
				art.Title = strings.TrimSpace(text)
				mode = modeHeader
			}
			continue
		}
		if mode == modeHeader || mode == modeFooter {
			if trim == "" {
				mode = modeBody
				continue
			}
			if text, ok := headerText(trim); ok {
				if !applyLabels(text, art) && mode == modeHeader && art.Subtitle == "" {
					art.Subtitle = strings.TrimSpace(text)
				}
				continue
			}
			mode = modeBody
		}
		if mode == modeBody && trim == "---" && nextIsHeader(lines, i) {
			mode = modeFooter
		}
	}
	if art.Title == "" {
		return nil, skipf("no title found")
	}
	if art.dateRaw != "" {
		d, err := normalizeDate(art.dateRaw)
		if err != nil {
			return nil, skipf("%s", err.Error())
		}
		art.Date = d
	}
	return art, nil
}

func normalizeDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	switch {
	case len(s) >= 10 && s[4] == '/':
		t, err := time.Parse("2006/01/02", s[:10])
		if err != nil {
			return "", fmt.Errorf("invalid date %q", s)
		}
		return t.Format("20060102"), nil
	case len(s) >= 10 && s[4] == '-':
		t, err := time.Parse("2006-01-02", s[:10])
		if err != nil {
			return "", fmt.Errorf("invalid date %q", s)
		}
		return t.Format("20060102"), nil
	case len(s) == 8 && digitsOnly(s):
		t, err := time.Parse("20060102", s)
		if err != nil {
			return "", fmt.Errorf("invalid date %q", s)
		}
		return t.Format("20060102"), nil
	case digitsOnly(s):
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid date %q", s)
		}
		return time.Unix(n, 0).UTC().Format("20060102"), nil
	default:
		return "", fmt.Errorf("invalid date %q", s)
	}
}

func digitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func formatDate(yyyymmdd, layout string) string {
	t, err := time.Parse("20060102", yyyymmdd)
	if err != nil {
		return yyyymmdd
	}
	switch layout {
	case "YYYY/MM/DD":
		return t.Format("2006/01/02")
	case "YYYY-MM-DD":
		return t.Format("2006-01-02")
	default:
		return t.Format("20060102")
	}
}

func insertAfterTitle(raw string, extra []string) string {
	parts := strings.Split(raw, "\n")
	for i, line := range parts {
		if _, ok := headerText(strings.TrimSpace(line)); ok {
			out := make([]string, 0, len(parts)+len(extra))
			out = append(out, parts[:i+1]...)
			out = append(out, extra...)
			out = append(out, parts[i+1:]...)
			return strings.Join(out, "\n")
		}
	}
	return strings.Join(extra, "\n") + "\n" + raw
}

func dropFirstHeader(raw string) string {
	parts := strings.Split(raw, "\n")
	for i, line := range parts {
		if _, ok := headerText(strings.TrimSpace(line)); ok {
			out := make([]string, 0, len(parts)-1)
			out = append(out, parts[:i]...)
			out = append(out, parts[i+1:]...)
			return strings.Join(out, "\n")
		}
	}
	return raw
}

// titleSlug keeps letters and digits and joins everything else with delim.
func titleSlug(title, delim string) string {
	if delim == "" || strings.ContainsAny(delim, `/\`) {
		delim = "_"
	}
	var b strings.Builder
	prevDelim := true
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDelim = false
			continue
		}
		if !prevDelim {
			b.WriteString(delim)
			prevDelim = true
		}
	}
	return strings.TrimSuffix(b.String(), delim)
}

func canonName(date, title, delim string) string {
	if delim == "" || strings.ContainsAny(delim, `/\`) {
		delim = "_"
	}
	slug := titleSlug(title, delim)
	if slug == "" {
		slug = "untitled"
	}
	return date + delim + slug
}

func unstampedName(filename string) string {
	ext := extOf(filename)
	base := strings.TrimSuffix(filename, ext)
	m := stampRE.FindStringSubmatch(base)
	if m == nil {
		return filename
	}
	return m[1] + ext
}

func stampedName(filename, digest string) string {
	ext := extOf(filename)
	base := strings.TrimSuffix(unstampedName(filename), ext)
	return base + "_" + digest + ext
}

func extOf(filename string) string {
	dot := strings.LastIndex(filename, ".")
	if dot <= 0 {
		return ""
	}
	return filename[dot:]
}
