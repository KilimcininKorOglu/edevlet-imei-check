package edevlet

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	tokenRegex = regexp.MustCompile(`name="token"\s+value="([^"]+)"`)
	dtRegex    = regexp.MustCompile(`<dt[^>]*>(.*?)</dt>`)
	ddRegex    = regexp.MustCompile(`<dd[^>]*>(.*?)</dd>`)
	tagRegex   = regexp.MustCompile(`<[^>]+>`)
)

func parseResultHTML(html string) (*QueryResult, error) {
	dlStart := strings.Index(html, "<dl")
	if dlStart == -1 {
		return nil, fmt.Errorf("no <dl> element found in result page")
	}
	dlEnd := strings.Index(html[dlStart:], "</dl>")
	if dlEnd == -1 {
		return nil, fmt.Errorf("no </dl> closing tag found")
	}
	dlHTML := html[dlStart : dlStart+dlEnd+5]

	dts := dtRegex.FindAllStringSubmatch(dlHTML, -1)
	dds := ddRegex.FindAllStringSubmatch(dlHTML, -1)

	fields := map[string]string{}
	for i := 0; i < len(dts) && i < len(dds); i++ {
		key := strings.TrimSpace(tagRegex.ReplaceAllString(dts[i][1], ""))
		val := strings.TrimSpace(tagRegex.ReplaceAllString(dds[i][1], ""))
		val = strings.ReplaceAll(val, "&quot;", "")
		fields[key] = val
	}

	result := &QueryResult{
		RawStatus: fields["Durum"],
		Source:    fields["Kaynak"],
	}

	if bm, ok := fields["Marka/Model"]; ok {
		result.Brand, result.Model = parseBrandModel(bm)
	}

	result.Status = NormalizeStatus(result.RawStatus)

	return result, nil
}

func NormalizeStatus(raw string) string {
	switch {
	case strings.Contains(raw, "KAYITLI"):
		return "registered"
	case strings.Contains(raw, "kapatılmış") || strings.Contains(raw, "kapatilmis"):
		return "blocked"
	case strings.Contains(raw, "KAYITDIŞI") || strings.Contains(raw, "KAYITDI"):
		return "unregistered"
	case strings.Contains(raw, "kopyalandığı") || strings.Contains(raw, "kopyaland"):
		return "cloned"
	case strings.Contains(raw, "BULUNAMADI") || strings.Contains(raw, "bulunamadı"):
		return "not_found"
	default:
		return "unknown"
	}
}

func parseBrandModel(raw string) (brand, model string) {
	if idx := strings.Index(raw, "Marka:"); idx != -1 {
		rest := raw[idx+6:]
		if end := strings.Index(rest, ","); end != -1 {
			brand = strings.TrimSpace(rest[:end])
		} else if end := strings.Index(rest, "Model"); end != -1 {
			brand = strings.TrimSpace(rest[:end])
		} else {
			brand = strings.TrimSpace(rest)
		}
	}
	if idx := strings.Index(raw, "Model Bilgileri:"); idx != -1 {
		model = strings.TrimSpace(raw[idx+16:])
	} else if idx := strings.Index(raw, "Pazar Adı:"); idx != -1 {
		rest := raw[idx+11:]
		if end := strings.IndexAny(rest, ",\n"); end != -1 {
			model = strings.TrimSpace(rest[:end])
		} else {
			model = strings.TrimSpace(rest)
		}
	}
	return
}
