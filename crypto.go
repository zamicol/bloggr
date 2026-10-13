package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyphrme/coz"
)

func (a *app) digestBytes(data []byte) (string, error) {
	sum, err := coz.Hash(a.alg, data)
	if err != nil {
		return "", err
	}
	return string(a.alg) + ":" + sum.String(), nil
}

func splitDigest(d string) (alg, value string) {
	alg, value, _ = strings.Cut(d, ":")
	return alg, value
}

// signArticle signs the bare article digest. Field order matches the README
// example. Coz calls this timestamp "now".
func (a *app) signArticle(digestValue string, now int64) (string, error) {
	if a.key == nil {
		return "", errors.New("missing coz key")
	}
	pay := fmt.Sprintf(`{"article":%q,"alg":%q,"now":%d,"tmb":%q,"typ":"bloggr/sign"}`,
		digestValue, string(a.key.Alg), now, a.key.Tmb.String())
	sum, err := coz.Hash(a.key.Alg.Hash(), []byte(pay))
	if err != nil {
		return "", err
	}
	sig, err := a.key.Sign(sum)
	if err != nil {
		return "", err
	}
	raw := []byte(fmt.Sprintf(`{"pay":%s,"sig":%q}`, pay, sig.String()))
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return "", err
	}
	buf.WriteByte('\n')
	return buf.String(), nil
}

func (a *app) loadKey() error {
	if !a.cfg.SignArticle {
		return nil
	}
	path := filepath.Join(a.blog, "coz_key.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		key, err := coz.NewKey(coz.SEAlg(coz.ES256))
		if err != nil {
			return err
		}
		key.Typ = "bloggr/key"
		kb, err := coz.MarshalPretty(key)
		if err != nil {
			return err
		}
		kb = append(kb, '\n')
		if err := os.WriteFile(path, kb, 0o600); err != nil {
			return err
		}
		a.key = key
		a.logf("Generated coz key %s", path)
		return nil
	}
	if err != nil {
		return err
	}
	key := new(coz.Key)
	if err := json.Unmarshal(b, key); err != nil {
		return fmt.Errorf("coz key: %w", err)
	}
	a.key = key
	return nil
}
