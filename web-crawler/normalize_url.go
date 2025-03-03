package main

import (
    "net/url"
    "strings"
)

func normalizeURL(rawURL string) (string, error) {
    u, err := url.Parse(rawURL)

    if err != nil {
        return "", err
    }

    if u.Host == "" {
        return "", err
    }

    result := u.Host

    if u.Path != "" {
        trimmedPath := strings.TrimSuffix(u.Path, "/")
        result += trimmedPath
    }

    if u.RawQuery != "" {
        result += "?" + u.RawQuery
    }

    if u.Fragment != "" {
        result += "#" + u.Fragment
    }

    return result, err
}
