package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/MY-Final/portWatch/internal/command"
)

// TestVersionIsSingleSourcedInMain guards the two version contracts:
//
//  1. goreleaser stamps the release number with -X main.version=..., which
//     only applies while version stays a plain string literal in main.go —
//     any other initializer makes the flag silently do nothing.
//  2. command.Version stays empty so the binary carries exactly one copy of
//     the number, injected by main; a default here would drift on bumps.
func TestVersionIsSingleSourcedInMain(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	// \r? keeps the anchor working for both LF and CRLF checkouts.
	if regexp.MustCompile(`(?m)^var version = "\d+\.\d+\.\d+"\r?$`).Find(source) == nil {
		t.Fatal(`main.go must declare version as a plain semver literal (var version = "X.Y.Z"); otherwise -X main.version stops overriding it`)
	}
	if command.Version != "" {
		t.Fatalf("command.Version default = %q, want empty; main injects the release number", command.Version)
	}
}
