package proxy

import (
	"bytes"
	"gogit-proxy/gitparser"
	"gogit-proxy/whitelist"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// NewProxy creates a reverse proxy that forwards requests to GitHub.
func NewProxy() *httputil.ReverseProxy {
	target, _ := url.Parse("https://github.com")
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ModifyResponse = func(response *http.Response) error {
		response.Header.Set("X-Forwarded-Host", response.Request.Host)
		return nil
	}

	return proxy
}

// HandleGitHubProxy proxies requests from /github.com/* to https://github.com/*
func HandleGitHubProxy(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request: %s %s", r.Method, r.URL.Path)
    owner, name, slug := whitelist.ParseRepoFromPath(r.URL.Path)
    if !whitelist.IsRepoAllowed(owner, name, slug) {
        http.Error(w, "Repository not allowed", http.StatusForbidden)
        log.Printf("Blocked repo: owner=%s, name=%s, slug=%s", owner, name, slug)
        return
    }

    // Only parse body for git-receive-pack POST (push)
    if strings.HasSuffix(r.URL.Path, "/git-receive-pack") && r.Method == "POST" {
        body, err := io.ReadAll(r.Body)
        r.Body = io.NopCloser(bytes.NewReader(body))
        r.GetBody = func() (io.ReadCloser, error) {
            return io.NopCloser(bytes.NewReader(body)), nil
        }
        if err == nil && len(body) >= 4 {
            pktLenHex := string(body[:4])
            plen, err := strconv.ParseUint(pktLenHex, 16, 16)
            if err == nil && plen >= 4 && int(plen) <= len(body) {
                pktLine := string(body[4:plen])
                packData := body[plen:]
                pushInfo, err := gitparser.ParsePush(pktLine, packData)
                if err == nil && pushInfo != nil {
                    log.Printf("Push: branch=%s old=%s new=%s commit=%+v", pushInfo.Reference, pushInfo.OldCommit, pushInfo.NewCommit, pushInfo.Commit)
                }
            }
        }
    }

    // Proxy all requests (including fetches) without reading the body
    targetPath := strings.TrimPrefix(r.URL.Path, "/github.com")
    proxy := httputil.NewSingleHostReverseProxy(&url.URL{
        Scheme: "https",
        Host:   "github.com",
    })
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.URL.Path = targetPath
        req.Host = "github.com"
    }
    proxy.ServeHTTP(w, r)
}