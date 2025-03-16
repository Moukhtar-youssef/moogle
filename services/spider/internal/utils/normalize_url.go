package utils

import (
    "net/url"
    "strings"
    "fmt"
)

func NormalizeURL(rawURL string) (string, error) {
    u, err := url.Parse(rawURL)

    if err != nil {
        return "", fmt.Errorf("Could not parse raw URL [%w]", err) 
    }

    if u.Host == "" {
        return "", fmt.Errorf("URL has no field field 'Host'")
    }

    host := u.Host
    if strings.HasPrefix(host, "www.") {
        host = host[4:]
    }

    normalizedURL := host

    if u.Path != "" {
        trimmedPath := strings.TrimSuffix(u.Path, "/")
        normalizedURL += trimmedPath
    }

    return normalizedURL, nil
}
