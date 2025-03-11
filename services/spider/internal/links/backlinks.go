package links

import (
    // "fmt"
    "strings"
)

func isValidURL(link string) bool {
    return !strings.Contains(link, "%")
}


func CreateBacklinks(outgoingLinks []string, normalizedCurrentURL string) error {
        /*
        * for link in outgoing links:
        *   redis.add(backling.outgoingurl, currenturl)
        */
    // fmt.Printf("Creating backlinks...\n")
    // for i, link := range outgoingLinks {
    //     if isValidURL(link) {
    //         // normalize url
    //         normalizedOutgoingURL, err := utils.NormalizeURL(link)
    //         if err != nil {
    //             continue
    //         }
    //         fmt.Printf("backlinks:%v = {%v}\n", normalizedOutgoingURL, normalizedCurrentURL)
    //     }
    // }

    return nil
}
