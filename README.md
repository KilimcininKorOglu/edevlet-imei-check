# edevlet-imei-check

A Go library that checks IMEI registration status on Turkey's e-Devlet (turkiye.gov.tr) portal. Uses [gemini-captcha-solver](https://github.com/KilimcininKorOglu/gemini-captcha-solver) to automatically solve CAPTCHAs.

## Installation

```bash
go get github.com/KilimcininKorOglu/edevlet-imei-check
```

## Usage

```go
package main

import (
    "fmt"
    "os"

    edevlet "github.com/KilimcininKorOglu/edevlet-imei-check"
)

func main() {
    client := edevlet.New(edevlet.Config{
        GeminiAPIKey: "your-gemini-api-key",
        GeminiModel:  "gemini-2.5-flash",
    })

    result, err := client.Query("352817020012523")
    if err != nil {
        fmt.Fprintf(os.Stderr, "query failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("IMEI:   %s\n", result.IMEI)
    fmt.Printf("Status: %s (%s)\n", result.RawStatus, result.Status)
    fmt.Printf("Brand:  %s\n", result.Brand)
    fmt.Printf("Model:  %s\n", result.Model)
}
```

## Status Values

| Status         | Raw Text (Turkish)                                            | Meaning                    |
|----------------|---------------------------------------------------------------|----------------------------|
| `registered`   | IMEI NUMARASI KAYITLI                                         | Legally registered         |
| `blocked`      | 1 yil veya daha uzun suredir kullanilmadigi icin kapatilmis   | Deactivated (inactive 1y+) |
| `unregistered` | KAYITDISI OLDUGU TESPIT EDILEN IMEI                           | Detected as unregistered   |
| `cloned`       | Bu IMEI numarasinin baska cihazlara kopyalandigi tespit edilmistir | Cloned IMEI detected  |
| `not_found`    | KAYIT BULUNAMADI                                              | Not in database            |
| `unknown`      | (any other response)                                          | Unrecognized status        |

## How It Works

1. Loads the e-Devlet IMEI query form page
2. Extracts CSRF token and downloads CAPTCHA image
3. Solves CAPTCHA using Google Gemini API
4. Submits the form with IMEI + CAPTCHA solution
5. Parses the result HTML page

Automatic retry: 3 attempts with 2s delay on CAPTCHA failures.

## Requirements

- Google Gemini API key ([get one here](https://aistudio.google.com/app/apikey))

## License

MIT
