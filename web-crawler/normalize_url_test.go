package main

import (
    "testing"
)

func TestNormalizeURL(t *testing.T) {
    tests := []struct {
        name          string
        inputURL      string
        expected      string
        wantErr       bool
    }{
        {
            name:     "remove https scheme",
            inputURL: "https://blog.boot.dev/path",
            expected: "blog.boot.dev/path",
            wantErr:  false,
        },
        {
            name:     "remove http scheme",
            inputURL: "http://blog.boot.dev/path",
            expected: "blog.boot.dev/path",
            wantErr:  false,
        },
        {
            name:     "no path",
            inputURL: "https://blog.boot.dev",
            expected: "blog.boot.dev",
            wantErr:  false,
        },
        {
            name:     "trailing slash",
            inputURL: "https://blog.boot.dev/",
            expected: "blog.boot.dev",
            wantErr:  false,
        },
        {
            name:     "www subdomain",
            inputURL: "https://www.blog.boot.dev/path",
            expected: "www.blog.boot.dev/path",
            wantErr:  false,
        },
        {
            name:     "URL with port",
            inputURL: "https://blog.boot.dev:8080/path",
            expected: "blog.boot.dev:8080/path",
            wantErr:  false,
        },
        {
            name:     "query parameters",
            inputURL: "https://blog.boot.dev/path?query=123",
            expected: "blog.boot.dev/path?query=123",
            wantErr:  false,
        },
        {
            name:     "URL with fragment",
            inputURL: "https://blog.boot.dev/path#section",
            expected: "blog.boot.dev/path#section",
            wantErr:  false,
        },
        {
            name:     "multiple paths",
            inputURL: "https://blog.boot.dev/path/to/resource/",
            expected: "blog.boot.dev/path/to/resource",
            wantErr:  false,
        },
        {
            name:     "multiple paths plus query",
            inputURL: "https://blog.boot.dev/path/to/resource?query=123",
            expected: "blog.boot.dev/path/to/resource?query=123",
            wantErr:  false,
        },
        {
            name:     "multiple paths plus query with trailing slash",
            inputURL: "https://blog.boot.dev/path/to/resource/?query=123",
            expected: "blog.boot.dev/path/to/resource?query=123",
            wantErr:  false,
        },
        {
            name:     "empty URL",
            inputURL: "",
            expected: "",
            wantErr:  false,
        },
        {
            name:     "no colon after scheme",
            inputURL: "https//example.com",
            expected: "",
            wantErr:  false,
        },
        {
            name:     "only scheme",
            inputURL: "https://",
            expected: "",
            wantErr:  false,
        },
        {
            name:     "missing host",
            inputURL: "https://?query=123",
            expected: "",
            wantErr:  false,
        },
        {
            name:     "fragment without host",
            inputURL: "https://#fragment",
            expected: "",
            wantErr:  false,
        },
        {
            name:     "invalid characters",
            inputURL: "https://exa mple.com",
            expected: "",
            wantErr:  true,
        },
    }

    for i, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            actual, err := normalizeURL(tc.inputURL)
            if err != nil && !tc.wantErr {
                t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
                return
            }
            if actual != tc.expected {
                t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
            }
        })
    }
}
