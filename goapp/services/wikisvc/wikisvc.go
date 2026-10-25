// Package wikisvc reads pages from the wiki's MediaWiki API (cfg.App.APIServer).
package wikisvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/zetaoss/zengine/goapp/app/config"
)

// Backlink limits: how many linking articles are read, how much text is kept on each side of the first link to
// the page, and how much of the beginning is kept when the link comes from a template and is not in the text.
const (
	maxBacklinks      = 10
	backlinkContext   = 300
	backlinkLeadRunes = 500
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

type revisionPage struct {
	PageID    int    `json:"pageid"`
	Title     string `json:"title"`
	Missing   bool   `json:"missing"`
	Revisions []struct {
		Slots map[string]struct {
			Content string `json:"content"`
		} `json:"slots"`
	} `json:"revisions"`
}

func (p revisionPage) content() string {
	if len(p.Revisions) == 0 {
		return ""
	}
	return p.Revisions[0].Slots["main"].Content
}

func query(ctx context.Context, cfg *config.Config, params url.Values, out any) error {
	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	if apiServer == "" {
		return fmt.Errorf("APP_URL is required")
	}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("formatversion", "2")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiServer+"/w/api.php?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wiki api: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func revisions(ctx context.Context, cfg *config.Config, params url.Values) ([]revisionPage, error) {
	params.Set("prop", "revisions")
	params.Set("rvprop", "content")
	params.Set("rvslots", "main")
	var payload struct {
		Query struct {
			Pages []revisionPage `json:"pages"`
		} `json:"query"`
	}
	if err := query(ctx, cfg, params, &payload); err != nil {
		return nil, err
	}
	return payload.Query.Pages, nil
}

// PageText returns the current wikitext of title, and false when the page does not exist.
func PageText(ctx context.Context, cfg *config.Config, title string) (string, bool, error) {
	pages, err := revisions(ctx, cfg, url.Values{"titles": {title}})
	if err != nil {
		return "", false, err
	}
	if len(pages) == 0 || pages[0].Missing {
		return "", false, nil
	}
	return pages[0].content(), true, nil
}

// BacklinkTexts returns, for up to maxBacklinks articles that link to title (main namespace, no redirects; the
// page may not exist yet), the article's title and the text around its first link to title, with references and
// comments removed. When the link comes from a template and is not in the text, the beginning of the article
// stands in.
func BacklinkTexts(ctx context.Context, cfg *config.Config, title string) ([]string, error) {
	var bl struct {
		Query struct {
			Backlinks []struct {
				PageID int `json:"pageid"`
			} `json:"backlinks"`
		} `json:"query"`
	}
	err := query(ctx, cfg, url.Values{
		"list":          {"backlinks"},
		"bltitle":       {title},
		"blnamespace":   {"0"},
		"blfilterredir": {"nonredirects"},
		"bllimit":       {strconv.Itoa(maxBacklinks)},
	}, &bl)
	if err != nil || len(bl.Query.Backlinks) == 0 {
		return nil, err
	}
	ids := make([]string, len(bl.Query.Backlinks))
	for i, b := range bl.Query.Backlinks {
		ids[i] = strconv.Itoa(b.PageID)
	}
	pages, err := revisions(ctx, cfg, url.Values{"pageids": {strings.Join(ids, "|")}})
	if err != nil {
		return nil, err
	}
	byID := map[int]revisionPage{}
	for _, p := range pages {
		byID[p.PageID] = p
	}
	link := linkPattern(title)
	texts := []string{}
	for _, b := range bl.Query.Backlinks {
		p, ok := byID[b.PageID]
		if !ok || p.Missing {
			continue
		}
		if t := strings.TrimSpace(Excerpt(p.content(), link)); t != "" {
			texts = append(texts, "문서: "+p.Title+"\n"+t)
		}
	}
	return texts, nil
}

var noiseRe = regexp.MustCompile(`(?s)<!--.*?-->|<ref[^>/]*/>|<ref[^>]*>.*?</ref>`)

// linkPattern matches a wiki link to title: [[Title]], [[title|label]], [[Title#section]], with spaces and
// underscores alike and the first letter in either case.
func linkPattern(title string) *regexp.Regexp {
	title = strings.TrimSpace(strings.ReplaceAll(title, "_", " "))
	first, size := utf8.DecodeRuneInString(title)
	head := regexp.QuoteMeta(string(first))
	if up, low := unicode.ToUpper(first), unicode.ToLower(first); up != low {
		head = "[" + regexp.QuoteMeta(string(up)+string(low)) + "]"
	}
	words := strings.Split(title[size:], " ")
	for i, w := range words {
		words[i] = regexp.QuoteMeta(w)
	}
	return regexp.MustCompile(`\[\[\s*` + head + strings.Join(words, `[ _]+`) + `\s*(#[^|\]]*)?(\|[^\]]*)?\]\]`)
}

// Excerpt returns the text around the first match of link in wikitext, backlinkContext runes on each side, with
// references and comments removed; without a match, the first backlinkLeadRunes runes.
func Excerpt(wikitext string, link *regexp.Regexp) string {
	clean := noiseRe.ReplaceAllString(wikitext, "")
	loc := link.FindStringIndex(clean)
	if loc == nil {
		return cut(clean, 0, backlinkLeadRunes)
	}
	start := utf8.RuneCountInString(clean[:loc[0]]) - backlinkContext
	end := utf8.RuneCountInString(clean[:loc[1]]) + backlinkContext
	return cut(clean, max(start, 0), end)
}

func cut(s string, from, to int) string {
	r := []rune(s)
	to = min(to, len(r))
	out := strings.TrimSpace(string(r[from:to]))
	if from > 0 {
		out = "…" + out
	}
	if to < len(r) {
		out += "…"
	}
	return out
}
