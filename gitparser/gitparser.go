package gitparser

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type PushInfo struct {
	OldCommit string
	NewCommit string
	Reference string
	Commit    *Commit
}

type Commit struct {
	Sha     string
	Parent  string
	Author  string
	Email   string
	Message string
	Date    string // For simplicity, string; parse as needed
	// Add more fields as needed
}

// ParsePush parses the packet line and pack data from a git-receive-pack request.
func ParsePush(packetLine string, packData []byte) (*PushInfo, error) {
	parts := strings.Split(packetLine, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid packet line: %s", packetLine)
	}
	oldCommit := parts[0]
	newCommit := parts[1]
	reference := strings.Trim(parts[2], "\x00\r\n ")

	var commit *Commit
	if newCommit != "0000000000000000000000000000000000000000" && len(packData) > 0 {
		c, err := parsePackData(packData)
		if err != nil {
			log.Printf("Failed to parse pack data: %v", err)
		} else {
			c.Sha = newCommit
			if c.Parent == "" {
				c.Parent = oldCommit
			}
			commit = c
		}
	}

	return &PushInfo{
		OldCommit: oldCommit,
		NewCommit: newCommit,
		Reference: reference,
		Commit:    commit,
	}, nil
}

// parsePackData is a minimal parser for the first commit object in the packfile.
func parsePackData(data []byte) (*Commit, error) {
	// Look for "PACK"
	idx := bytes.Index(data, []byte("PACK"))
	if idx == -1 {
		return nil, fmt.Errorf("no PACK signature found")
	}
	// Skip 12-byte header
	pos := idx + 12
	if pos >= len(data) {
		return nil, fmt.Errorf("pack header too short")
	}
	// Parse object header
	objType := (data[pos] >> 4) & 7
	if objType != 1 { // 1 = commit
		return nil, fmt.Errorf("not a commit object")
	}
	// Skip variable-length size
	for data[pos]&0x80 != 0 {
		pos++
	}
	pos++

	// Decompress commit object
	zr, err := zlib.NewReader(bytes.NewReader(data[pos:]))
	if err != nil {
		return nil, fmt.Errorf("zlib: %w", err)
	}
	defer zr.Close()
	commitData, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("zlib read: %w", err)
	}

	// Parse commit fields (very basic)
	lines := strings.Split(string(commitData), "\n")
	var commit Commit
	for _, line := range lines {
		if strings.HasPrefix(line, "parent ") {
			commit.Parent = strings.TrimSpace(strings.TrimPrefix(line, "parent "))
		} else if strings.HasPrefix(line, "author ") {
			// Format: author Name <email> timestamp timezone
			parts := strings.SplitN(line[7:], "<", 2)
			if len(parts) == 2 {
				commit.Author = strings.TrimSpace(parts[0])
				rest := strings.SplitN(parts[1], ">", 2)
				if len(rest) == 2 {
					commit.Email = strings.TrimSpace(rest[0])
				}
			}
		} else if line == "" {
			// Message starts after first blank line
			break
		}
	}
	// Message is after the first blank line
	msgIdx := strings.Index(string(commitData), "\n\n")
	if msgIdx != -1 {
		commit.Message = strings.TrimSpace(string(commitData)[msgIdx+2:])
	}
	return &commit, nil
}

// Example function to demonstrate packet line parsing
func ExampleParsePush(body []byte) {
	pktLenHex := string(body[:4])
	plen, err := strconv.ParseUint(pktLenHex, 16, 16)
	if err != nil || plen < 4 || int(plen) > len(body) {
		log.Printf("Invalid or non-hex packet line length: %q (err=%v, plen=%d, bodylen=%d)", pktLenHex, err, plen, len(body))
	} else {
		pktLine := string(body[4:plen])
		packData := body[plen:]
		// Further processing...
		log.Printf("Packet Line: %s, Pack Data: %x", pktLine, packData)
	}
}