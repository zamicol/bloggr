package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the bloggr config.json file. Paths inside Input and Output are
// relative to Path unless they are absolute. The file itself is read from the
// working directory.
type Config struct {
	URIBasePath string `json:"URI_BASE_PATH"`
	Path        string `json:"Path"`
	Input       string `json:"Input"`
	Output      string `json:"Output"`
	Digest      string `json:"Digest"`

	MaxLogSize     int64  `json:"MaxLogSize"`
	DateFormat     string `json:"DateFormat"`
	DefaultAuthor  string `json:"DefaultAuthor"`
	TitleDelimiter string `json:"TitleDelimiter"`

	InsertDate   bool `json:"InsertDate"`
	InsertAuthor bool `json:"InsertAuthor"`
	StyleCSS     bool `json:"StyleCSS"`
	History      bool `json:"History"`
	HashImages   bool `json:"HashImages"`
	HashArticle  bool `json:"HashArticle"`
	SignArticle  bool `json:"SignArticle"`

	ArticleListShowDate   bool `json:"ArticleListShowDate"`
	ArticleListShowTags   bool `json:"ArticleListShowTags"`
	ArticleListShowAuthor bool `json:"ArticleListShowAuthor"`
	DeleteAndRebuild      bool `json:"DeleteAndRebuild"`
}

// DefaultConfig returns the README defaults. Bool fields that default to true
// are set here so a config file can omit them.
func DefaultConfig() Config {
	return Config{
		Path:           ".",
		Input:          "markdown",
		Output:         "public",
		Digest:         "SHA-256",
		MaxLogSize:     10 * 1024 * 1024,
		DateFormat:     "YYYYMMDD",
		DefaultAuthor:  "Anonymous",
		TitleDelimiter: "_",

		InsertDate:   true,
		InsertAuthor: true,
		StyleCSS:     true,
		History:      true,
		HashImages:   true,
		HashArticle:  true,
		SignArticle:  true,

		ArticleListShowDate:   true,
		ArticleListShowTags:   true,
		ArticleListShowAuthor: false,
		DeleteAndRebuild:      false,
	}
}

// LoadConfig reads path, creating it from DefaultConfig when it is missing.
// Keys absent from an existing file keep their defaults.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := saveConfig(path, cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func saveConfig(path string, cfg Config) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func (c Config) blogDir() (string, error) {
	p := c.Path
	if p == "" {
		p = "."
	}
	return filepath.Abs(p)
}

func (c Config) resolveDir(blog, value, fallback string) string {
	if value == "" {
		value = fallback
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(blog, value)
}
