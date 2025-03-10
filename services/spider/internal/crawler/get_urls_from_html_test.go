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
    }

    for i, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL)

            if err != nil {
                t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
                return
            }

            result := reflect.DeepEqual(tc.expected, actual)

            if result == false {
                t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
            }

        })
    }
}

