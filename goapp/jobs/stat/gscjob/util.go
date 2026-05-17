package gscjob

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type GAReader struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
	PropertyID  string `json:"property_id"`
	PropertyID2 string `json:"propertyId"`
	Timezone    string `json:"timezone"`
	GscSiteURL  string `json:"gsc_site_url"`
	GscSiteURL2 string `json:"gscSiteUrl"`
	SiteURL     string `json:"site_url"`
	SiteURL2    string `json:"siteUrl"`
}

func LoadGAReaderFile(jobCtx job.JobContext) (*GAReader, error) {
	path := jobCtx.Config().Analytics.GAReaderFile
	if path == "" {
		return nil, fmt.Errorf("GA_READER_FILE is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gr GAReader
	if err := json.Unmarshal(raw, &gr); err != nil {
		return nil, err
	}
	if gr.ClientEmail == "" || gr.PrivateKey == "" {
		return nil, fmt.Errorf("service account missing client_email/private_key")
	}
	if gr.TokenURI == "" {
		gr.TokenURI = "https://oauth2.googleapis.com/token"
	}
	return &gr, nil
}

func FetchGoogleAccessToken(ctx context.Context, sa *GAReader, scope string) (string, error) {
	now := time.Now().Unix()
	headerJSON, _ := json.Marshal(app.H{"alg": "RS256", "typ": "JWT"})
	claimJSON, _ := json.Marshal(app.H{
		"iss":   sa.ClientEmail,
		"scope": scope,
		"aud":   sa.TokenURI,
		"iat":   now,
		"exp":   now + 3600,
	})
	header := base64URLEncode(headerJSON)
	claim := base64URLEncode(claimJSON)
	input := header + "." + claim

	key, err := parseRSAPrivateKey(sa.PrivateKey)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	assertion := input + "." + base64URLEncode(sig)

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sa.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("token request failed: %d %s", resp.StatusCode, string(raw))
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("missing access_token")
	}
	return out.AccessToken, nil
}

func parseRSAPrivateKey(key string) (*rsa.PrivateKey, error) {
	key = strings.ReplaceAll(key, `\n`, "\n")
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("invalid private key pem")
	}
	if pkcs8, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := pkcs8.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return rsaKey, nil
	}
	if pkcs1, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return pkcs1, nil
	}
	return nil, fmt.Errorf("unsupported private key format")
}

func base64URLEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func RunGSCQuery(ctx context.Context, token, siteURL string, body app.H) (app.H, error) {
	encoded := url.PathEscape(siteURL)
	endpoint := "https://searchconsole.googleapis.com/webmasters/v3/sites/" + encoded + "/searchAnalytics/query"
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("gsc api failed: %d %s", resp.StatusCode, string(b))
	}
	var payload app.H
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func asInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		n, _ := strconv.Atoi(x)
		return n
	default:
		return 0
	}
}

func asFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	default:
		return 0
	}
}

func round4(v float64) float64 {
	return float64(int(v*10000+0.5)) / 10000
}
