package crawler

// Tests taken from www.boot.dev - Web Crawler

import (
    "testing"
    "reflect"
)

func TestGetURLsFromHTML(t *testing.T) {
    tests := []struct {
        name        string
        inputURL    string
        inputBody   string
        expected    []string
    } {
        {
            name:     "absolute and relative URLs",
            inputURL: "https://randomsite.com",
            inputBody: `
            <html>
                <body>
                    <h1>Some text here</h1>
                    <a href="/path/one">
                        <span>randomsite</span>
                    </a>
                    <a href="https://othersite.com/path/one">
                        <span>othersite.com</span>
                    </a>
                </body>
            </html>
            `,
            expected: []string{"https://randomsite.com/path/one", "https://othersite.com/path/one"},
        },
        {
            name:     "no URLs",
            inputURL: "https://randomsite.com",
            inputBody: `
            <html>
                <body>
                    <h1>Empty Website</h1>
                </body>
            </html>
            `,
            expected: []string{},
        },
        {
            name:     "malformed HTML but valid links",
            inputURL: "https://example.com",
            inputBody: `
            <html>
                <body>
                    <a href="/valid-link"><span>Valid</span></a>
                    <a href="<invalid></a>"><span>Broken</span></a>
                    <a href="https://valid.com/path"></a>
                </body>
            </html>
            `,
            expected: []string{"https://example.com/valid-link", "https://valid.com/path"},
        },
        {
            name:     "remove duplicate links",
            inputURL: "https://example.com",
            inputBody: `
            <html>
                <body>
                    <a href="/valid-link"><span>Valid</span></a>
                    <a href="<invalid></a>"><span>Broken</span></a>
                    <a href="https://valid.com/path"></a>
                    <a href="/valid-link"><span>Valid</span></a>
                    <a href="<invalid></a>"><span>Broken</span></a>
                    <a href="https://valid.com/path"></a>
                </body>
            </html>
            `,
            expected: []string{"https://example.com/valid-link", "https://valid.com/path"},
        },
        {
            name:     "ignore non-ASCII links",
            inputURL: "https://example.com",
            inputBody: `
            <html>
                <body>
                    <a href="/valid-link"><span>Valid</span></a>
                    <a href="https://valid.com/path"></a>
                    <a href="https://пример.рф">Cyrillic</a>
                    <a href="https://例子.com">Chinese</a>
                    <a href="https://テスト.jp">Japanese</a>
                    <a href="/another-valid"></a>
                </body>
            </html>
            `,
            expected: []string{
                "https://example.com/valid-link",
                "https://valid.com/path",
                "https://example.com/another-valid",
            },
        },
    }

    for i, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            actual, _, err := getURLsFromHTML(tc.inputBody, tc.inputURL)

            if err != nil {
                t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
                return
            }

            // Convert slices to sets for commparison
            expectedSet := make(map[string]struct{})
            for _, e := range tc.expected {
                expectedSet[e] = struct{}{}
            }

            actualSet := make(map[string]struct{})
            for _, e := range actual {
                actualSet[e] = struct{}{}
            }

            result := reflect.DeepEqual(expectedSet, actualSet)

            if result == false {
                t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
            }

        })
    }
}

