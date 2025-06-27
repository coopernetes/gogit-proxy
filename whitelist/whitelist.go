package whitelist

import (
    "encoding/json"
    "os"
    "strings"
    "sync"
)

type RepoWhitelist struct {
    Owners []string `json:"owners"`
    Names  []string `json:"names"`
    Slugs  []string `json:"slugs"`
}

var (
    whitelist     RepoWhitelist
    whitelistOnce sync.Once
)

func LoadWhitelist(path string) error {
    whitelistOnce.Do(func() {
        f, err := os.Open(path)
        if err != nil {
            return
        }
        defer f.Close()
        json.NewDecoder(f).Decode(&whitelist)
    })
    return nil
}

func IsRepoAllowed(owner, name, slug string) bool {
    for _, o := range whitelist.Owners {
        if o == owner {
            return true
        }
    }
    for _, n := range whitelist.Names {
        if n == name {
            return true
        }
    }
    for _, s := range whitelist.Slugs {
        if s == slug {
            return true
        }
    }
    return false
}

// Helper to extract owner, name, slug from /github.com/{owner}/{repo}/...
func ParseRepoFromPath(path string) (owner, name, slug string) {
    parts := strings.Split(strings.TrimPrefix(path, "/github.com/"), "/")
    if len(parts) >= 2 {
        owner, name = parts[0], parts[1]
        slug = owner + "/" + name
    }
    return
}