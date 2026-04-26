# edevlet-imei-check

A Go library that checks IMEI registration status on Turkey's e-Devlet (turkiye.gov.tr) portal. Automatically solves CAPTCHAs using [gemini-captcha-solver](https://github.com/KilimcininKorOglu/gemini-captcha-solver).

## Installation

```bash
go get github.com/KilimcininKorOglu/edevlet-imei-check
```

Requires Go 1.22 or later (developed with Go 1.26).

## Quick Start

```go
package main

import (
	"fmt"
	"os"

	edevlet "github.com/KilimcininKorOglu/edevlet-imei-check"
)

func main() {
	client := edevlet.New(edevlet.Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	})

	result, err := client.Query("YOUR_IMEI_NUMBER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("IMEI:   %s\n", result.IMEI)
	fmt.Printf("Status: %s (%s)\n", result.RawStatus, result.Status)
	fmt.Printf("Source: %s\n", result.Source)
	fmt.Printf("Brand:  %s\n", result.Brand)
	fmt.Printf("Model:  %s\n", result.Model)
}
```

## API Key Pool

For bulk queries, use multiple Gemini API keys to avoid rate limits:

```go
client := edevlet.New(edevlet.Config{
	GeminiAPIKeys: []string{
		os.Getenv("GEMINI_KEY_1"),
		os.Getenv("GEMINI_KEY_2"),
		os.Getenv("GEMINI_KEY_3"),
	},
})
```

Keys are rotated automatically on rate limit (429). See [gemini-captcha-solver](https://github.com/KilimcininKorOglu/gemini-captcha-solver) for details.

## Configuration

| Field         | Type     | Default                 | Description                               |
|---------------|----------|-------------------------|-------------------------------------------|
| GeminiAPIKey  | string   |                         | Single Gemini API key                     |
| GeminiAPIKeys | []string |                         | Key pool (takes priority over single key) |
| GeminiModel   | string   | `gemini-2.5-flash-lite` | Gemini model name                         |
| MaxAttempts   | int      | `10`                    | Max query retry attempts                  |

## Status Values

The `QueryResult.Status` field is normalized to one of these values:

| Status         | e-Devlet Response                                                  | Meaning                    |
|----------------|--------------------------------------------------------------------|----------------------------|
| `registered`   | IMEI NUMARASI KAYITLI                                              | Legally registered         |
| `blocked`      | 1 yıl veya daha uzun süredir kullanılmadığı için kapatılmış cihaz  | Deactivated (inactive 1y+) |
| `unregistered` | KAYITDIŞI OLDUĞU TESPİT EDİLEN IMEI                                | Detected as unregistered   |
| `cloned`       | Bu IMEI numarasının başka cihazlara kopyalandığı tespit edilmiştir | Cloned IMEI detected       |
| `not_found`    | KAYIT BULUNAMADI                                                   | Not in database            |
| `unknown`      | (any other response)                                               | Unrecognized status        |

## QueryResult Fields

```go
type QueryResult struct {
	IMEI      string // Queried IMEI number
	Status    string // Normalized status (see table above)
	RawStatus string // Original Turkish text from e-Devlet
	Source    string // Registration source (e.g. "İthalat yoluyla kaydedilen IMEI")
	Brand     string // Device brand
	Model     string // Device model
}
```

## How It Works

1. Loads the e-Devlet IMEI query form page
2. Extracts CSRF token and downloads CAPTCHA image
3. Solves CAPTCHA via Google Gemini API
4. Submits the form with IMEI, CAPTCHA solution, and CSRF token
5. Follows the 302 redirect to the result page
6. Parses the HTML response and normalizes the status

Each query creates a fresh HTTP session (new cookie jar) to avoid token reuse issues. Failed attempts are retried up to 10 times (configurable via `MaxAttempts`) with a 5-second delay.

## Requirements

- Google Gemini API key -- get one free at [Google AI Studio](https://aistudio.google.com/app/apikey)
- Network access to turkiye.gov.tr

## License

MIT
