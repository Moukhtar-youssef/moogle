package main

import(
    "strings"
    "net/url"
    "fmt"

    "golang.org/x/net/html"
)

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
    node, err := html.Parse(strings.NewReader(htmlBody))
    if err != nil {
        return nil, err
    }

    links := traverse(node, baseURL)

    if links == nil {
        return []string{}, nil
    }

    return links, nil
}

func traverse(node *html.Node, baseURL *url.URL) []string {
    if node == nil {
        return nil
    }

    var links []string

    if node.Type == html.ElementNode && node.Data == "a" {
        for _, attr := range node.Attr {
            if attr.Key == "href" {
                rawHref := attr.Val

                if strings.ContainsAny(rawHref, " <>\"") {
                    fmt.Printf("Skipping malformed URL: %v\n", rawHref)
                    continue
                }

                // Parse url and add to the list
                u, err := url.Parse(attr.Val)
                if err != nil {
                    continue
                }

                var resolved string

                if u.IsAbs() {
                    resolved = u.String()
                } else {
                    resolved = baseURL.ResolveReference(u).String()

                }

                // Append to list
                links = append(links, resolved)
            }
        }
    }

    for c := node.FirstChild; c != nil; c = c.NextSibling {
        links = append(links, traverse(c, baseURL)...)
    }

    return links
}

