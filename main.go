package main

import (
    "log"
    "net/http"
    "gogit-proxy/proxy"
    "gogit-proxy/whitelist"
)

func main() {
    err := whitelist.LoadWhitelist("whitelist.json")
    if err != nil {
        log.Fatalf("Failed to load whitelist: %v", err)
    }
    http.HandleFunc("/github.com/", proxy.HandleGitHubProxy)
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}