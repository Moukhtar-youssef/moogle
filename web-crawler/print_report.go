package main

import(
    "fmt"
    "sort"
)

type pageCount struct {
    url     string
    count   int
}

func printReport(pages map[string]int, baseURL string) {
    fmt.Println("\n================================")
    fmt.Printf(" REPORT for %v \n", baseURL)
    fmt.Println("================================")

    var sortedPages []pageCount
    for url, count := range pages {
        sortedPages = append(sortedPages, pageCount{url, count})
    }

    sort.Slice(sortedPages, func(i, j int) bool {
        if sortedPages[i].count == sortedPages[j].count {
            return sortedPages[i].url < sortedPages[j].url
        }

        return sortedPages[i].count > sortedPages[j].count
    })

    for _, p := range sortedPages {
        var counter = "links"
        if p.count == 1 {
            counter = "link"
        }

        fmt.Printf("Found %d %s to %s\n", p.count, counter, p.url)
    }

    fmt.Println("--------------------------------")
    fmt.Printf("Total unique links found: %v\n", len(sortedPages))
    fmt.Println("--------------------------------")
}
