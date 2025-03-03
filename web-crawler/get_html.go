package main

import (
    "fmt"
    "io"
    "net/http"
    "strings"
)
func getHTML(rawURL string) (string, error) {
    res, err := http.Get(rawURL)

    if err != nil {
        return "", fmt.Errorf("failed to fetch URL: %w", err)
    }

    defer res.Body.Close() // Close the body to prevent memory leaks or something I don't remember

    if res.StatusCode > 399 {
        return "", fmt.Errorf("HTTP error: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
    }

    contentType := res.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "text/html") {
        return "", fmt.Errorf("invalid content type: %s", contentType)
    }

    body, err := io.ReadAll(res.Body)

    if err != nil {
        return "", fmt.Errorf("failed to read response body: %w", err)
    }

    return string(body), nil
}

