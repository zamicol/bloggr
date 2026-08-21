package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	Coze "github.com/cyphrme/coze"
	"github.com/gomarkdown/markdown"
	"github.com/zamicol/jsonflag"
)

const (
	defaultConfig = `{
		"path": ".", // Path for bloggr operations. (Default directories: testing dir, public, to_publish, and published)
		"to_publish": "0_to_publish", // Prepend `0_` so directory appears first when sorted.
		"published": "0_published", // Prepend `0_` so directory appears first when sorted.

		"hashing_alg": "SHA-256",
		"max_log_size": 10485760,
		"date_format": "YYYY/MM/DD",
		"default_author": "Anonymous",

		"move_processed": true,
		"insert_date": true,
		"insert_author": true,
		"title_delimiter": "_",
		"article_list_show_date": true,
		"article_list_show_tags": true,
		"article_list_show_author": true,
		"hash_images": true,
		"hash_article": true,
		"sign_article": true,
		"delete_and_rebuild": false
	}`
)

type Config struct {
    Path          string `json:"path" flag:"path" desc:"Path for bloggr operations. (Default directories: testing dir, log file, header, footer, public, to_publish, and published)"`
    ToPublish     string `json:"to_publish" flag:"to_publish" desc:"To publish directory for input markdown articles to be processed"`
    Published     string `json:"published" flag:"published" desc:"Published directory.  This directory is served by the http server."`

    HashingAlg        string `json:"hashing_alg" flag:"hashing-alg" desc:"Hashing algorithm (e.g., SHA-256)"`
    MaxLogSize        int64  `json:"max_log_size" flag:"max-log-size" desc:"Max log size in bytes"`
    DateFormat        string `json:"date_format" flag:"date-format" desc:"Date format for articles"`
    DefaultAuthor     string `json:"default_author" flag:"default-author" desc:"Default author name"`

    MoveProcessed         bool `json:"move_processed" flag:"move-processed" desc:"Move processed files"`
    InsertDate            bool `json:"insert_date" flag:"insert-date" desc:"Insert date if missing"`
    InsertAuthor          bool `json:"insert_author" flag:"insert-author" desc:"Insert author if missing"`
    TitleDelimiter        string `json:"title_delimiter" flag:"title-delimiter" desc:"Delimiter for titles in filenames"`
    ArticleListShowDate   bool `json:"article_list_show_date" flag:"article-list-show-date" desc:"Show date in article list"`
    ArticleListShowTags   bool `json:"article_list_show_tags" flag:"article-list-show-tags" desc:"Show tags in article list"`
    ArticleListShowAuthor bool `json:"article_list_show_author" flag:"article-list-show-author" desc:"Show author in article list"`
    HashImages            bool `json:"hash_images" flag:"hash-images" desc:"Hash image filenames"`
    HashArticle           bool `json:"hash_article" flag:"hash-article" desc:"Hash article content"`
    SignArticle           bool `json:"sign_article" flag:"sign-article" desc:"Sign article with Coze"`
    DeleteAndRebuild      bool `json:"delete_and_rebuild" flag:"delete-and-rebuild" desc:"Delete and rebuild output"`
}

type Article struct {
	Title    string
	Date     string
	Author   string
	Tags     []string
	Content  string
	Images   map[string]string // original -> hashed name
	FilePath string
}

var logger *log.Logger

func main() {
	// Initialize logger to stdout temporarily
	logger = log.New(os.Stdout, "bloggr: ", log.LstdFlags)
	logger.Println("Starting bloggr")

	// Parse config with jsonflag
	config := &Config{}
	jsonflag.SetDefault(defaultConfig) // Set default config
	err := jsonflag.Parse(config, "config.json")
	if err != nil {
		logger.Fatal("Failed to parse config:", err)
	}

	// Initialize logging with config values
	logFile, err := setupLogging(filepath.Join(config.Path, "bloggr.log"), config.MaxLogSize)
	if err != nil {
		log.Fatal("Failed to setup logging:", err)
	}
	defer logFile.Close()
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "bloggr: ", log.LstdFlags)

	// Ensure directories and default files exist
	createDefaultFiles(config)

	// Process articles
	articles, err := processArticles(config)
	if err != nil {
		logger.Fatal("Failed to process articles:", err)
	}

	// Generate main page
	if err := generateMainPage(config, articles); err != nil {
		logger.Fatal("Failed to generate main page:", err)
	}

	logger.Println("bloggr completed successfully")
}

func setupLogging(logPath string, maxSize int64) (*os.File, error) {
	if _, err := os.Stat(logPath); err == nil {
		info, err := os.Stat(logPath)
		if err != nil {
			return nil, err
		}
		if info.Size() >= maxSize {
			os.Rename(logPath, logPath+".old")
		}
	}
	return os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func createDefaultFiles(config *Config) {
	files := map[string]string{
		"article_header.html": "default_header.html",
		"article_footer.html": "default_footer.html",
		".gitignore":          "default_gitignore",
	}
	for path, content := range files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logger.Printf("Creating default %s", path)
			os.WriteFile(path, []byte(content), 0644)
		}
	}
	dirs := []string{
		config.Input,
		config.Output,
		filepath.Join(config.Output, "a"),
		filepath.Join(config.Output, "0_processed"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func processArticles(config *Config) ([]Article, error) {
	var articles []Article

	// Clean output if DeleteAndRebuild is true
	if config.DeleteAndRebuild {
		os.RemoveAll(filepath.Join(config.Output, "a"))
		os.MkdirAll(filepath.Join(config.Output, "a"), 0755)
	}

	err := filepath.Walk(config.Input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != config.Input {
			return processDir(path, config, &articles)
		}
		if filepath.Ext(path) == ".md" && filepath.Dir(path) == config.Input {
			return processFile(path, config, &articles)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func processDir(dir string, config *Config, articles *[]Article) error {
	mdFiles := []string{}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".md" {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if len(mdFiles) != 1 {
		logger.Printf("Ignoring %s: %d .md files found (expected 1)", dir, len(mdFiles))
		return nil
	}
	return processFile(mdFiles[0], config, articles)
}

func processFile(path string, config *Config, articles *[]Article) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	article, err := parseArticle(data, path, config)
	if err != nil {
		logger.Printf("Skipping %s: %v", path, err)
		return nil
	}

	// Hash images
	if config.HashImages {
		article.Images = make(map[string]string)
		dir := filepath.Dir(path)
		filepath.Walk(dir, func(imgPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(imgPath) == ".md" {
				return nil
			}
			imgData, err := os.ReadFile(imgPath)
			if err != nil {
				return err
			}
			digest, err := Coze.Hash(Coze.HshAlg(config.Digest), imgData)
			if err != nil {
				return err
			}

			newName := fmt.Sprintf("%s_%s_%s%s", filepath.Base(imgPath[:len(imgPath)-len(filepath.Ext(imgPath))]), config.Digest, digest, filepath.Ext(imgPath))
			article.Images[filepath.Base(imgPath)] = newName
			article.Content = strings.ReplaceAll(article.Content, filepath.Base(imgPath), newName)
			return nil
		})
	}

	// Hash article
	var articleHash string
	if config.HashArticle {
		digest, err := Coze.Hash(Coze.HshAlg(config.Digest), []byte(article.Content))
		if err != nil {
			return err
		}
		articleHash = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(digest)
	}

	// Sign article
	var sig string
	if config.SignArticle && articleHash != "" {
		key, err := loadOrGenerateCozeKey("coze_key.json")
		if err != nil {
			return err
		}
		sigB64, err := key.Sign(Coze.B64(articleHash))
		if err != nil {
			return err
		}
		sig = string(sigB64)
	}

	// Append crypto section
	if config.HashArticle || config.SignArticle {
		cryptoSection := "\n<!-- Start Article Crypto -->\nThis article's markdown file is integrity protected:\n  " + config.Digest + ":" + articleHash
		if sig != "" {
			cryptoSection += fmt.Sprintf(`
    This article's markdown file is signed by the author using Coze:
{
  "pay": {
    "a": "%s",
    "alg": "ES256",
    "iat": %d,
    "tmb": "cLj8vsYtMBwYkzoFVZHBZo6SNL8wSdCIjCKAwXNuhOk",
    "typ": "bloggr/sign"
  },
  "sig": "%s"
}`, articleHash, time.Now().Unix(), sig)
		}
		cryptoSection += "\n<!-- End Article Crypto -->\n"
		article.Content += cryptoSection
	}

	// Generate HTML
	html := generateArticleHTML(article, config)

	// Write output
	outDir := filepath.Join(config.Output, "a", article.Date+"_"+strings.ReplaceAll(article.Title, " ", config.TitleDelimiter))
	os.MkdirAll(outDir, 0755)
	if err := os.WriteFile(filepath.Join(outDir, "index.html"), html, 0644); err != nil {
		return err
	}
	for orig, hashed := range article.Images {
		os.Rename(filepath.Join(filepath.Dir(path), orig), filepath.Join(outDir, hashed))
	}

	// Move to processed
	if config.MoveProcessed {
		procDir := filepath.Join(config.Output, "0_processed")
		if filepath.Dir(path) != config.Input {
			procDir = filepath.Join(procDir, article.Date+"_"+strings.ReplaceAll(article.Title, " ", config.TitleDelimiter))
			os.MkdirAll(procDir, 0755)
			os.Rename(path, filepath.Join(procDir, "index.md"))
			for _, hashed := range article.Images {
				os.Rename(filepath.Join(outDir, hashed), filepath.Join(procDir, hashed))
			}
		} else {
			os.Rename(path, filepath.Join(procDir, article.Date+"_"+strings.ReplaceAll(article.Title, " ", config.TitleDelimiter)+".md"))
		}
	}

	*articles = append(*articles, *article)
	return nil
}

func parseArticle(data []byte, path string, config *Config) (*Article, error) {
	lines := strings.Split(string(data), "\n")
	var title, date, author string
	var tags []string
	var content strings.Builder
	inHeader := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" && !inHeader {
			continue
		}
		if strings.HasPrefix(line, "# ") && !inHeader {
			title = strings.TrimPrefix(line, "# ")
			inHeader = true
		} else if inHeader && (strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "#### ") || strings.HasPrefix(line, "##### ") || strings.HasPrefix(line, "###### ")) {
			if strings.Contains(line, "Date: ") {
				date = strings.TrimSpace(strings.SplitN(line, "Date: ", 2)[1])
				if strings.Contains(date, "/") {
					date = strings.ReplaceAll(date, "/", "")
				}
			} else if strings.Contains(line, "Author: ") {
				author = strings.TrimSpace(strings.SplitN(line, "Author: ", 2)[1])
			} else if strings.Contains(line, "Tags: ") {
				tags = strings.Split(strings.TrimSpace(strings.SplitN(line, "Tags: ", 2)[1]), ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}
		} else if inHeader && line != "" {
			inHeader = false
			content.WriteString(strings.Join(lines[i:], "\n"))
			break
		}
	}

	if title == "" {
		return nil, fmt.Errorf("no # title found")
	}
	if date == "" && config.InsertDate {
		date = time.Now().Format("20060102")
	}
	if author == "" && config.InsertAuthor {
		author = config.DefaultAuthor
	}
	return &Article{Title: title, Date: date, Author: author, Tags: tags, Content: content.String(), FilePath: path}, nil
}

func generateArticleHTML(article *Article, config *Config) []byte {
	header, _ := os.ReadFile("article_header.html")
	footer, _ := os.ReadFile("article_footer.html")
	htmlContent := markdown.ToHTML([]byte(article.Content), nil, nil)

	var result bytes.Buffer
	result.WriteString(strings.ReplaceAll(string(header), "{ARTICLE_TITLE}", article.Title))
	result.Write(htmlContent)
	result.WriteString(string(footer))
	return result.Bytes()
}

func loadOrGenerateCozeKey(path string) (*Coze.Key, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Println("Generating new Coze key at", path)
		// Placeholder: Replace with actual Coze key generation
		key := &Coze.Key{} // Mock; use real Coze key gen
		data, _ := json.Marshal(key)
		os.WriteFile(path, data, 0600)
		return key, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var key Coze.Key
	json.Unmarshal(data, &key)
	return &key, nil
}

func generateMainPage(config *Config, articles []Article) error {
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date > articles[j].Date
	})

	var mdContent strings.Builder
	for _, a := range articles {
		line := "["
		if config.ArticleListShowDate {
			d := a.Date[:4] + "/" + a.Date[4:6] + "/" + a.Date[6:]
			line += d + "] "
		}
		line += a.Title
		if config.ArticleListShowAuthor {
			line += " (Author: " + a.Author
			if config.ArticleListShowTags && len(a.Tags) > 0 {
				line += ", Tags: " + strings.Join(a.Tags, ", ")
			}
			line += ")"
		} else if config.ArticleListShowTags && len(a.Tags) > 0 {
			line += " (Tags: " + strings.Join(a.Tags, ", ")
		}
		mdContent.WriteString(line + "\n")
	}

	procIndex := filepath.Join(config.Output, "0_processed", "index.md")
	os.WriteFile(procIndex, []byte(mdContent.String()), 0644)

	var htmlContent strings.Builder
	for _, a := range articles {
		href := fmt.Sprintf("[%s_%s]/index.html", a.Date, strings.ReplaceAll(a.Title, " ", config.TitleDelimiter))
		line := "<p><a href=\"" + href + "\">["
		if config.ArticleListShowDate {
			d := a.Date[:4] + "/" + a.Date[4:6] + "/" + a.Date[6:]
			line += d + "] "
		}
		line += a.Title + "</a>"
		if config.ArticleListShowAuthor {
			line += " (Author: " + a.Author
			if config.ArticleListShowTags && len(a.Tags) > 0 {
				line += ", Tags: " + strings.Join(a.Tags, ", ")
			}
			line += ")"
		} else if config.ArticleListShowTags && len(a.Tags) > 0 {
			line += " (Tags: " + strings.Join(a.Tags, ", ")
		}
		line += "</p>"
		htmlContent.WriteString(line + "\n")
	}

	indexPath := filepath.Join(config.Output, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		os.WriteFile(indexPath, []byte("<!DOCTYPE html>\n<html><body>\n"+htmlContent.String()+"</body></html>"), 0644)
	} else {
		data, _ := os.ReadFile(indexPath)
		start := strings.Index(string(data), "<!-- Start Article List -->")
		end := strings.Index(string(data), "<!-- End Article List -->")
		if start == -1 || end == -1 {
			return fmt.Errorf("invalid home page ([output directory]/index.html format")
		}
		newContent := string(data[:start+len("<!-- Start Article List -->\n")]) + htmlContent.String() + string(data[end:])
		os.WriteFile(indexPath, []byte(newContent), 0644)
	}

	return nil
}
