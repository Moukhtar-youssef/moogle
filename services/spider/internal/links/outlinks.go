package links

import (
    "fmt"
    "encoding/json"
)

type Outlinks struct {
    NormalizedURL   string
    Links           []string
}

func CreateOutlinks(normalizedURL string, links []string) *Outlinks {
    return &Outlinks {
        NormalizedURL:  normalizedURL,
        Links:          links,
    }
}

func (outlinks *Outlinks) GetOutlinks() []string {
    // This is definitely not well encapsulated but who cares
    // I'll fix it later
    // I should check the struct fields
    return outlinks.Links
}

func HashOutlinks(outlinks *Outlinks) (map[string]interface{}, error) {
    outlinksJSON, err := json.Marshal(outlinks.Links)
    if err != nil {
        return nil, fmt.Errorf("Error hashing struct: %w", err)
    }

    // Convert to redis hash
    return map[string]interface{}{
        "links":    string(outlinksJSON),
    }, nil
}

func (o Outlinks) String() string {
    return fmt.Sprintf(
        "-------------------------------------------------\n" +
        "Page %s has %d outlinks\n" +
        "-------------------------------------------------\n",
        o.NormalizedURL, len(o.Links),
        )
}
