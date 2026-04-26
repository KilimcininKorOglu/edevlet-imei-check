package edevlet

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	captcha "github.com/KilimcininKorOglu/gemini-captcha-solver"
)

const (
	baseURL    = "https://www.turkiye.gov.tr"
	formPath   = "/imei-sorgulama"
	captchaURL = "/captcha?uniquePage=877"
	userAgent  = "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"

	defaultMaxAttempts = 10
	maxMaxAttempts     = 50
	retryDelay         = 5 * time.Second
)

type Config struct {
	GeminiAPIKey  string
	GeminiAPIKeys []string
	GeminiModel   string
	MaxAttempts   int
}

type QueryResult struct {
	IMEI      string `json:"imei"`
	Status    string `json:"status"`
	RawStatus string `json:"raw_status"`
	Source    string `json:"source"`
	Brand     string `json:"brand"`
	Model     string `json:"model"`
}

type Client struct {
	solver      *captcha.Solver
	maxAttempts int
}

func New(cfg Config) *Client {
	model := cfg.GeminiModel
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}
	solverRetries := len(cfg.GeminiAPIKeys) + 3
	if solverRetries < 5 {
		solverRetries = 5
	}
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	if maxAttempts > maxMaxAttempts {
		maxAttempts = maxMaxAttempts
	}
	return &Client{
		maxAttempts: maxAttempts,
		solver: captcha.New(captcha.Config{
			APIKey:     cfg.GeminiAPIKey,
			APIKeys:    cfg.GeminiAPIKeys,
			Model:      model,
			MaxRetries: solverRetries,
			Prompt:     "Read the CAPTCHA text. Reply with ONLY the characters (letters and numbers), nothing else. The CAPTCHA is usually 5 characters.",
		}),
	}
}

func validateIMEI(imei string) error {
	if len(imei) != 15 {
		return fmt.Errorf("invalid IMEI: must be exactly 15 characters, got %d", len(imei))
	}
	for i, c := range imei {
		if c < '0' || c > '9' {
			return fmt.Errorf("invalid IMEI: non-digit character at position %d", i)
		}
	}
	return nil
}

func (c *Client) Query(imei string) (*QueryResult, error) {
	if err := validateIMEI(imei); err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		result, err := c.queryOnce(imei)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if attempt < c.maxAttempts {
			time.Sleep(retryDelay)
		}
	}
	return nil, fmt.Errorf("query failed after %d attempts: %w", c.maxAttempts, lastErr)
}

func (c *Client) queryOnce(imei string) (*QueryResult, error) {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	token, err := loadFormPage(httpClient)
	if err != nil {
		return nil, fmt.Errorf("load form: %w", err)
	}

	captchaImage, err := downloadCaptcha(httpClient)
	if err != nil {
		return nil, fmt.Errorf("download captcha: %w", err)
	}

	captchaCode, err := c.solver.Solve(captchaImage)
	if err != nil {
		return nil, fmt.Errorf("solve captcha: %w", err)
	}
	if captchaCode == "" {
		return nil, fmt.Errorf("solve captcha: empty solution")
	}

	redirectPath, err := submitForm(httpClient, imei, captchaCode, token)
	if err != nil {
		return nil, fmt.Errorf("submit form: %w", err)
	}

	httpClient.CheckRedirect = nil

	result, err := fetchResult(httpClient, redirectPath)
	if err != nil {
		return nil, fmt.Errorf("fetch result: %w", err)
	}

	result.IMEI = imei
	return result, nil
}

func loadFormPage(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", baseURL+formPath, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	matches := tokenRegex.FindSubmatch(body)
	if matches == nil {
		return "", fmt.Errorf("CSRF token not found")
	}

	return string(matches[1]), nil
}

func downloadCaptcha(client *http.Client) ([]byte, error) {
	u := fmt.Sprintf("%s%s&rnd=%f", baseURL, captchaURL, rand.Float64())

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/*")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Referer", baseURL+formPath)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("captcha download HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) < 100 {
		return nil, fmt.Errorf("captcha image too small: %d bytes", len(data))
	}

	return data, nil
}

func submitForm(client *http.Client, imei, captchaCode, token string) (string, error) {
	form := url.Values{
		"txtImei":      {imei},
		"captcha_name": {captchaCode},
		"token":        {token},
	}

	req, err := http.NewRequest("POST", baseURL+formPath+"?submit", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", baseURL)
	req.Header.Set("Referer", baseURL+formPath)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != 302 {
		return "", fmt.Errorf("expected 302, got HTTP %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if !strings.Contains(location, "asama=1") {
		return "", fmt.Errorf("unexpected redirect: %s (captcha may be wrong)", location)
	}

	return location, nil
}

func fetchResult(client *http.Client, path string) (*QueryResult, error) {
	resultURL := baseURL + path
	if strings.HasPrefix(path, "http") {
		resultURL = path
	}

	req, err := http.NewRequest("GET", resultURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseResultHTML(string(body))
}
