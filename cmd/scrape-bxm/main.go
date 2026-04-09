package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pages = []struct {
	url    string
	output string
}{
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/concepts/overview.html", "overview.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_1.html", "dev-process.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_3.html", "dbio.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_4.html", "bean.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_5.html", "service.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_1.html", "select-multi.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_2.html", "select-paging.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_6.html", "transaction.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_7.html", "exception.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_8.html", "logging.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-batch/ch01/ch01_4.html", "batch.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-configuration-reference-server/ch04/ch04.html", "config.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-user-guide-studio/ch01/ch01.html", "studio.md"},
}

func main() {
	outDir := "knowledge/docs/bxm"
	os.MkdirAll(outDir, 0755)

	for _, page := range pages {
		fmt.Printf("Fetching %s → %s\n", page.url, page.output)

		resp, err := http.Get(page.url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", page.url, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		md := htmlToMarkdown(string(body))
		outPath := filepath.Join(outDir, page.output)
		if err := os.WriteFile(outPath, []byte(md), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR writing %s: %v\n", outPath, err)
		}
		fmt.Printf("  → %s (%d bytes)\n", outPath, len(md))
	}

	fmt.Println("\nDone. Run 'go run ./cmd/build-index/' to update index.json")
}

// htmlToMarkdown is a simple HTML→Markdown converter for SWLab docs.
func htmlToMarkdown(html string) string {
	// Remove script/style/nav/header/footer tags (no backreference — Go regexp doesn't support \1)
	for _, tag := range []string{"script", "style", "nav", "header", "footer"} {
		re := regexp.MustCompile(`(?s)<` + tag + `[^>]*>.*?</` + tag + `>`)
		html = re.ReplaceAllString(html, "")
	}

	// Extract main content (article or main tag — no backreference in Go regexp)
	for _, tag := range []string{"article", "main"} {
		re := regexp.MustCompile(`(?s)<` + tag + `[^>]*>(.*?)</` + tag + `>`)
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			html = m[1]
			break
		}
	}

	// Convert headings
	for i := 6; i >= 1; i-- {
		re := regexp.MustCompile(fmt.Sprintf(`(?s)<h%d[^>]*>(.*?)</h%d>`, i, i))
		prefix := strings.Repeat("#", i) + " "
		html = re.ReplaceAllString(html, "\n"+prefix+"$1\n")
	}

	// Convert code blocks
	rePre := regexp.MustCompile(`(?s)<pre[^>]*><code[^>]*class="[^"]*language-(\w+)"[^>]*>(.*?)</code></pre>`)
	html = rePre.ReplaceAllString(html, "\n```$1\n$2\n```\n")
	rePre2 := regexp.MustCompile(`(?s)<pre[^>]*>(.*?)</pre>`)
	html = rePre2.ReplaceAllString(html, "\n```\n$1\n```\n")

	// Convert inline code
	reCode := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
	html = reCode.ReplaceAllString(html, "`$1`")

	// Convert lists
	reLi := regexp.MustCompile(`<li[^>]*>(.*?)</li>`)
	html = reLi.ReplaceAllString(html, "- $1")

	// Convert paragraphs
	reP := regexp.MustCompile(`<p[^>]*>(.*?)</p>`)
	html = reP.ReplaceAllString(html, "\n$1\n")

	// Convert tables (basic)
	reTd := regexp.MustCompile(`<t[dh][^>]*>(.*?)</t[dh]>`)
	html = reTd.ReplaceAllString(html, "| $1 ")
	reTr := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	html = reTr.ReplaceAllString(html, "$1|\n")

	// Strip remaining HTML tags
	reTag := regexp.MustCompile(`<[^>]+>`)
	html = reTag.ReplaceAllString(html, "")

	// Decode HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")
	html = strings.ReplaceAll(html, "&nbsp;", " ")

	// Clean up whitespace
	reBlank := regexp.MustCompile(`\n{3,}`)
	html = reBlank.ReplaceAllString(html, "\n\n")

	return strings.TrimSpace(html)
}
