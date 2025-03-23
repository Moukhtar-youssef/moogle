package utils

// Tests taken from www.boot.dev - Web Crawler

import (
    "testing"
)

func TestStripURL(t *testing.T) {
    tests := []struct {
        name        string
        inputURL    string
        expected    string
        wantErr     bool
    }{
        {
            name: "remove trailing slash",
            inputURL: "http://en.wikipedia.org/wiki/Mega_Man_X/",
            expected: "http://en.wikipedia.org/wiki/Mega_Man_X",
            wantErr: false,
        },
        {
            name: "remove fragments",
            inputURL: "https://en.wikipedia.org/wiki/Mega_Man_X#Plot",
            expected: "https://en.wikipedia.org/wiki/Mega_Man_X",
            wantErr: false,
        },
        {
            name: "remove query parameters",
            inputURL: "https://en.wikipedia.org/wiki/Mega_Man_X?version=1.0&language=en",
            expected: "https://en.wikipedia.org/wiki/Mega_Man_X",
            wantErr: false,
        },
        {
            name: "don't remove www.",
            inputURL: "https://www.mults.com/",
            expected: "https://www.mults.com",
            wantErr: false,
        },
    }

    for i, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            actual, err := StripURL(tc.inputURL)
            if err != nil && !tc.wantErr {
                t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
                return
            }

            if actual != tc.expected {
                t.Errorf("Test %v - '%s' FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
            }
        })
    }
}
