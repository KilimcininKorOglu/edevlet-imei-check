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
	thRegex    = regexp.MustCompile(`<th[^>]*>(.*?)</th>`)
	tdRegex    = regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
)

func parseResultHTML(html string) (*QueryResult, error) {
	if dlStart := strings.Index(html, "<dl"); dlStart != -1 {
		return parseDLFormat(html, dlStart)
	}

	if tableStart := strings.Index(html, "<table"); tableStart != -1 {
		return parseTableFormat(html, tableStart)
	}

	status := extractStatusFromText(html)
	if status != "" {
		return &QueryResult{
			RawStatus: status,
			Status:    NormalizeStatus(status),
		}, nil
	}

	return nil, fmt.Errorf("no recognized result format in response page")
}

func parseDLFormat(html string, dlStart int) (*QueryResult, error) {
	dlEnd := strings.Index(html[dlStart:], "</dl>")
	if dlEnd == -1 {
		return nil, fmt.Errorf("no </dl> closing tag found")
	}
	dlHTML := html[dlStart : dlStart+dlEnd+5]

	dts := dtRegex.FindAllStringSubmatch(dlHTML, -1)
	dds := ddRegex.FindAllStringSubmatch(dlHTML, -1)

	fields := map[string]string{}
	for i := 0; i < len(dts) && i < len(dds); i++ {
		key := cleanHTML(dts[i][1])
		val := cleanHTML(dds[i][1])
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

func parseTableFormat(html string, tableStart int) (*QueryResult, error) {
	tableEnd := strings.Index(html[tableStart:], "</table>")
	if tableEnd == -1 {
		return nil, fmt.Errorf("no </table> closing tag found")
	}
	tableHTML := html[tableStart : tableStart+tableEnd+8]

	rows := strings.Split(tableHTML, "<tr")
	fields := map[string]string{}
	for _, row := range rows {
		ths := thRegex.FindAllStringSubmatch(row, -1)
		tds := tdRegex.FindAllStringSubmatch(row, -1)
		for i := 0; i < len(ths) && i < len(tds); i++ {
			key := cleanHTML(ths[i][1])
			val := cleanHTML(tds[i][1])
			fields[key] = val
		}
		if len(ths) == 0 && len(tds) >= 2 {
			key := cleanHTML(tds[0][1])
			val := cleanHTML(tds[1][1])
			fields[key] = val
		}
	}

	result := &QueryResult{
		RawStatus: fields["Durum"],
		Source:    fields["Kaynak"],
	}

	if bm, ok := fields["Marka/Model"]; ok {
		result.Brand, result.Model = parseBrandModel(bm)
	}

	result.Status = NormalizeStatus(result.RawStatus)
	if result.Status == "unknown" && result.RawStatus == "" {
		for _, v := range fields {
			if NormalizeStatus(v) != "unknown" {
				result.RawStatus = v
				result.Status = NormalizeStatus(v)
				break
			}
		}
	}

	return result, nil
}

func extractStatusFromText(html string) string {
	stripped := cleanHTML(html)
	statusPatterns := []string{
		"KAYITLI", "KAYITDIŞI", "BULUNAMADI",
		"kapatılmış", "kapatilmis",
		"kopyalandığı", "kopyaland",
	}
	for _, p := range statusPatterns {
		if strings.Contains(stripped, p) {
			idx := strings.Index(stripped, p)
			start := idx - 50
			if start < 0 {
				start = 0
			}
			end := idx + len(p) + 50
			if end > len(stripped) {
				end = len(stripped)
			}
			return strings.TrimSpace(stripped[start:end])
		}
	}
	return ""
}

func cleanHTML(s string) string {
	s = tagRegex.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&quot;", "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return strings.TrimSpace(s)
}

func NormalizeStatus(raw string) string {
	switch {
	case strings.Contains(raw, "KAYITLI") && !strings.Contains(raw, "KAYITDIŞI"):
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
