# edevlet-imei-check

e-Devlet (turkiye.gov.tr) uzerinden IMEI kayit durumu sorgulayan Go kutuphanesi. CAPTCHA cozumunu [ai-captcha-solver](https://github.com/KilimcininKorOglu/ai-captcha-solver) ile otomatik olarak yapar.

## Kurulum

```bash
go get github.com/KilimcininKorOglu/edevlet-imei-check
```

Go 1.22 veya ustu gereklidir (Go 1.26 ile gelistirilmistir).

## Hizli Baslangic

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

	result, err := client.Query("IMEI_NUMARASI")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sorgu basarisiz: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("IMEI:   %s\n", result.IMEI)
	fmt.Printf("Durum:  %s (%s)\n", result.RawStatus, result.Status)
	fmt.Printf("Kaynak: %s\n", result.Source)
	fmt.Printf("Marka:  %s\n", result.Brand)
	fmt.Printf("Model:  %s\n", result.Model)
}
```

## API Key Havuzu

Toplu sorgularda rate limit'e takilmamak icin birden fazla Gemini API anahtari kullanabilirsiniz:

```go
client := edevlet.New(edevlet.Config{
	GeminiAPIKeys: []string{
		os.Getenv("GEMINI_KEY_1"),
		os.Getenv("GEMINI_KEY_2"),
		os.Getenv("GEMINI_KEY_3"),
	},
})
```

Rate limit (429) alindiginda anahtarlar otomatik olarak rotate edilir. Detaylar icin [ai-captcha-solver](https://github.com/KilimcininKorOglu/ai-captcha-solver) dokumantasyonuna bakin.

## Yapilandirma

| Alan          | Tip      | Varsayilan              | Aciklama                                     |
|---------------|----------|-------------------------|----------------------------------------------|
| GeminiAPIKey  | string   |                         | Tek API anahtari                             |
| GeminiAPIKeys | []string |                         | Anahtar havuzu (tek anahtara gore oncelikli) |
| GeminiModel   | string   | `gemini-2.5-flash-lite` | Gemini model adi                             |
| MaxAttempts   | int      | `10`                    | Maksimum sorgu deneme sayisi                 |

## Durum Degerleri

`QueryResult.Status` alani asagidaki degerlerden birine normalize edilir:

| Durum          | e-Devlet Yaniti                                                        | Anlami                         |
|----------------|------------------------------------------------------------------------|--------------------------------|
| `registered`   | IMEI NUMARASI KAYITLI                                                  | Yasal olarak kayitli           |
| `blocked`      | 1 yil veya daha uzun suredir kullanilmadigi icin kapatilmis cihaz      | Devre disi (1 yildan fazla)    |
| `unregistered` | KAYITDISI OLDUGU TESPIT EDILEN IMEI                                    | Kayit disi tespit edilmis      |
| `cloned`       | Bu IMEI numarasinin baska cihazlara kopyalandigi tespit edilmistir     | Klonlanmis IMEI tespit edilmis |
| `not_found`    | KAYIT BULUNAMADI                                                       | Veritabaninda bulunamadi       |
| `unknown`      | (diger yanitlar)                                                       | Taninmayan durum               |

## QueryResult Alanlari

```go
type QueryResult struct {
	IMEI      string // Sorgulanan IMEI numarasi
	Status    string // Normalize edilmis durum (yukaridaki tabloya bakin)
	RawStatus string // e-Devlet'ten gelen orijinal Turkce metin
	Source    string // Kayit kaynagi
	Brand     string // Cihaz markasi
	Model     string // Cihaz modeli
}
```

## Nasil Calisir

1. e-Devlet IMEI sorgulama form sayfasini yukler
2. CSRF token'i cikarir ve CAPTCHA gorselini indirir
3. [ai-captcha-solver](https://github.com/KilimcininKorOglu/ai-captcha-solver) ile CAPTCHA'yi cozer (Gemini, OpenAI, Anthropic destekler)
4. IMEI, CAPTCHA cozumu ve CSRF token ile formu gonderir
5. 302 yonlendirmesini takip ederek sonuc sayfasina ulasir
6. HTML yanitini parse eder ve durumu normalize eder

Her sorgu yeni bir HTTP oturumu (temiz cookie jar) olusturur. Basarisiz denemeler 5 saniye arayla en fazla 10 kez (MaxAttempts ile ayarlanabilir) tekrarlanir.

## Gereksinimler

- CAPTCHA cozumu icin AI API anahtari -- ucretsiz Gemini anahtari almak icin [Google AI Studio](https://aistudio.google.com/app/apikey)
- turkiye.gov.tr'ye erisim

## Lisans

MIT
