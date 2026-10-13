package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type assetFile struct {
	onDisk   string // name before this run
	fileName string // name after an optional digest stamp
	orig     string // name with any stamp removed
	digest   string // alg:b64, empty when not hashed
	data     []byte
}

func isImage(name string) bool {
	switch strings.ToLower(extOf(name)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".bmp", ".ico", ".avif", ".tif", ".tiff":
		return true
	default:
		return false
	}
}

// collectAssets reads the article directory one level deep. Images are
// digest-stamped in place when HashImages is set. Every non-markdown file is
// returned so it can be copied to the output.
func (a *app) collectAssets(dir, mdPath string) ([]assetFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var assets []assetFile
	hashThem := a.cfg.HashImages || a.cfg.History
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			a.logf("Skipping nested directory %s", a.rel(path))
			continue
		}
		if path == mdPath || strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileName := e.Name()
		digest := ""
		if hashThem {
			digest, err = a.digestBytes(data)
			if err != nil {
				return nil, err
			}
		}
		if a.cfg.HashImages && isImage(e.Name()) && digest != "" {
			fileName = stampedName(e.Name(), digest)
			if fileName != e.Name() {
				dest := filepath.Join(dir, fileName)
				if err := os.Rename(path, dest); err != nil {
					return nil, err
				}
			}
		}
		assets = append(assets, assetFile{
			onDisk:   e.Name(),
			fileName: fileName,
			orig:     unstampedName(e.Name()),
			digest:   digest,
			data:     data,
		})
	}
	return assets, nil
}

func applyAssetNames(body string, assets []assetFile) string {
	type rep struct{ old, new string }
	var reps []rep
	for _, as := range assets {
		if as.onDisk != as.fileName {
			reps = append(reps, rep{as.onDisk, as.fileName})
		}
		if as.orig != as.fileName && as.orig != as.onDisk {
			reps = append(reps, rep{as.orig, as.fileName})
		}
	}
	sort.Slice(reps, func(i, j int) bool { return len(reps[i].old) > len(reps[j].old) })
	for _, r := range reps {
		body = replaceFilename(body, r.old, r.new)
	}
	return body
}

func replaceFilename(content, old, new string) string {
	if old == "" || old == new || !strings.Contains(content, old) {
		return content
	}
	var b strings.Builder
	b.Grow(len(content) + len(new))
	i := 0
	for i < len(content) {
		j := strings.Index(content[i:], old)
		if j < 0 {
			b.WriteString(content[i:])
			break
		}
		j += i
		end := j + len(old)
		if nameBounded(content, j, end) {
			b.WriteString(content[i:j])
			b.WriteString(new)
		} else {
			b.WriteString(content[i:end])
		}
		i = end
	}
	return b.String()
}

func nameBounded(s string, start, end int) bool {
	if start > 0 && isNameByte(s[start-1]) {
		return false
	}
	if end < len(s) && isNameByte(s[end]) {
		return false
	}
	return true
}

func isNameByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
		c == '_' || c == '-' || c == '.' || c == ':' || c >= 128
}

func assetMap(assets []assetFile) map[string]string {
	m := make(map[string]string)
	for _, as := range assets {
		if as.digest == "" {
			continue
		}
		m[as.orig] = as.digest
	}
	if len(m) == 0 {
		return nil
	}
	return m
}
