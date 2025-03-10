package pages

import (
    "time"
    "fmt"
    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
)

type Page struct {
    NormalizedURL   string
    HTML            string
    ContentType     string
    StatusCode      int
    LastCrawled     time.Time
}

// Custom String() method
func (p Page) String() string {
    htmlPreview := p.HTML

    // Truncate the HTML output
    if len(htmlPreview) > 15 {
        htmlPreview = htmlPreview[:15] + "..."
    }

    return fmt.Sprintf(
        "-------------------------------------------------\n" +
        "Normalized URL:    %-10s\n" +
        "HTML:              %-40s\n" +
        "Last Crawled:      %-30s\n" +
        "Status Code:       %-10d\n" +
        "Content Type:      %-20s\n" +
        "-------------------------------------------------\n",
        p.NormalizedURL, htmlPreview, p.LastCrawled.Format(time.RFC1123),
        p.StatusCode, p.ContentType,
        )
}

func CreatePage(normalizedUrl, html, contentType string, statusCode int) *Page {
    return &Page {
        NormalizedURL:  normalizedUrl,
        HTML:           html,
        ContentType:    contentType,
        StatusCode:     statusCode,
        LastCrawled:    time.Now(),
    }
}

func HashPage(page *Page) (map[string]interface{}, error) {
    // Convert it to a redis hash
    return map[string]interface{}{
        "normalized_url":   page.NormalizedURL,
        "html":             page.HTML,
        "content_type":     page.ContentType,
        "status_code":      page.StatusCode,
        "last_crawled":     page.LastCrawled.Format(time.RFC1123),
    }, nil
}

func DehashPage(data map[string]string) (*Page, error) {

    lastCrawled, err := utils.ParseTime(data["last_crawled"])
    if err != nil {
        return nil, fmt.Errorf("Error parsing 'LastCrawled' in hash: %w", err)
    }

    statusCode, err := utils.ParseInt(data["status_code"])
    if err != nil {
        return nil, fmt.Errorf("Error parsing 'StatusCode' in hash: %w", err)
    }

    return &Page {
        NormalizedURL:  data["normalized_url"],
        HTML:           data["html"],
        ContentType:    data["content_type"],
        StatusCode:     statusCode,
        LastCrawled:    lastCrawled,
    }, nil
}
