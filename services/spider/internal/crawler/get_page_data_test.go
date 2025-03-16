package crawler

import (
    "testing"
)

func TestGetPageData(t *testing.T) {
    tests := []struct {
        name        string
        inputURL    string
    } {
        {
            name:     "absolute https url",
            inputURL: "https://ionelpopjara.github.io/",
        },
        {
            name:     "absolute http url",
            inputURL: "http://ionelpopjara.github.io/",
        },
    }

    for i, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            _, _, _, err := getPageData(tc.inputURL)

            if err != nil {
                t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
                return
            }

        })
    }
}

