# edevlet-imei-check

e-Devlet (turkiye.gov.tr) üzerinden IMEI kayıt durumu sorgulayan Go kütüphanesi. CAPTCHA çözümünü [ai-captcha-solver](https://github.com/KilimcininKorOglu/ai-captcha-solver) ile otomatik olarak yapar.

## Kurulum

```bash
go get github.com/KilimcininKorOglu/edevlet-imei-check
```

Go 1.22 veya üstü gereklidir (Go 1.26 ile geliştirilmiştir).

## Hızlı Başlangıç

```go
package main

import (
	"fmt"
	"os"

	edevlet "github.com/KilimcininKorOglu/edevlet-imei-check"
)

func main() {
	client := edevlet.New(edevlet.Config{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})

	result, err := client.Query("IMEI_NUMARASI")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sorgu başarısız: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("IMEI:   %s\n", result.IMEI)
	fmt.Printf("Durum:  %s (%s)\n", result.RawStatus, result.Status)
	fmt.Printf("Kaynak: %s\n", result.Source)
	fmt.Printf("Marka:  %s\n", result.Brand)
	fmt.Printf("Model:  %s\n", result.Model)
}
```

## Farklı Sağlayıcı Kullanımı

Gemini dışında OpenAI veya Anthropic uyumlu API'ler de kullanılabilir:

```go
client := edevlet.New(edevlet.Config{
	Provider: "anthropic",
	BaseURL:  "https://nvidia.srv.hermestech.uk/v1/messages",
	APIKey:   os.Getenv("HERMES_API_KEY"),
	Model:    "microsoft/phi-4-multimodal-instruct",
})
```

Desteklenen sağlayıcılar ve ücretsiz seçenekler için [ai-captcha-solver dokümantasyonuna](https://github.com/KilimcininKorOglu/ai-captcha-solver#free-api-providers) bakın.

## API Anahtar Havuzu

Toplu sorgularda rate limit'e takılmamak için birden fazla API anahtarı kullanabilirsiniz (yalnızca Gemini):

```go
client := edevlet.New(edevlet.Config{
	APIKeys: []string{
		os.Getenv("KEY_1"),
		os.Getenv("KEY_2"),
		os.Getenv("KEY_3"),
	},
})
```

Rate limit (429) alındığında anahtarlar otomatik olarak rotate edilir. Detaylar için [ai-captcha-solver](https://github.com/KilimcininKorOglu/ai-captcha-solver) dokümantasyonuna bakın.

## Yapılandırma

| Alan        | Tip      | Varsayılan              | Açıklama                                             |
|-------------|----------|-------------------------|------------------------------------------------------|
| Provider    | string   | `gemini`                | Sağlayıcı: `gemini`, `openai`, `anthropic`           |
| BaseURL     | string   | Sağlayıcı varsayılanı   | Özel API base URL'i                                   |
| APIKey      | string   |                         | Tek API anahtarı                                      |
| APIKeys     | []string |                         | Anahtar havuzu (yalnızca Gemini, tek anahtara göre öncelikli) |
| Model       | string   | `gemini-2.5-flash-lite` | Model adı (OpenAI/Anthropic için zorunlu)             |
| MaxAttempts | int      | `10`                    | Maksimum sorgu deneme sayısı                          |

## Durum Değerleri

`QueryResult.Status` alanı aşağıdaki değerlerden birine normalize edilir:

| Durum          | e-Devlet Yanıtı                                                        | Anlamı                         |
|----------------|------------------------------------------------------------------------|--------------------------------|
| `registered`   | IMEI NUMARASI KAYITLI                                                  | Yasal olarak kayıtlı           |
| `blocked`      | 1 yıl veya daha uzun süredir kullanılmadığı için kapatılmış cihaz      | Devre dışı (1 yıldan fazla)    |
| `unregistered` | KAYITDIŞI OLDUĞU TESPİT EDİLEN IMEI                                   | Kayıt dışı tespit edilmiş      |
| `cloned`       | Bu IMEI numarasının başka cihazlara kopyalandığı tespit edilmiştir     | Klonlanmış IMEI tespit edilmiş |
| `not_found`    | KAYIT BULUNAMADI                                                       | Veritabanında bulunamadı       |
| `unknown`      | (diğer yanıtlar)                                                       | Tanınmayan durum               |

## QueryResult Alanları

```go
type QueryResult struct {
	IMEI      string // Sorgulanan IMEI numarası
	Status    string // Normalize edilmiş durum (yukarıdaki tabloya bakın)
	RawStatus string // e-Devlet'ten gelen orijinal Türkçe metin
	Source    string // Kayıt kaynağı
	Brand     string // Cihaz markası
	Model     string // Cihaz modeli
}
```

## Nasıl Çalışır

1. e-Devlet IMEI sorgulama form sayfasını yükler
2. CSRF token'ı çıkarır ve CAPTCHA görselini indirir
3. [ai-captcha-solver](https://github.com/KilimcininKorOglu/ai-captcha-solver) ile CAPTCHA'yı çözer
4. IMEI, CAPTCHA çözümü ve CSRF token ile formu gönderir
5. 302 yönlendirmesini takip ederek sonuç sayfasına ulaşır
6. HTML yanıtını parse eder ve durumu normalize eder

Her sorgu yeni bir HTTP oturumu (temiz cookie jar) oluşturur. Başarısız denemeler 5 saniye arayla en fazla 10 kez (`MaxAttempts` ile ayarlanabilir) tekrarlanır.

## Gereksinimler

- CAPTCHA çözümü için AI API anahtarı -- ücretsiz seçenekler için [ai-captcha-solver dokümantasyonuna](https://github.com/KilimcininKorOglu/ai-captcha-solver#free-api-providers) bakın
- turkiye.gov.tr'ye erişim

## Lisans

MIT
