package fetcher

import (
	"fmt"
	"net/url"
	"strings"
)

type SourceType string

const (
	SourceGitHub          SourceType = "github"
	SourceHTTP            SourceType = "http"
	SourceGoogleWellKnown SourceType = "google-well-known"
	SourceLocal           SourceType = "local"
	SourceProtodex        SourceType = "protodex"
)

type SourceInfo struct {
	Type    SourceType
	Source  string // the main identifier (repo, service name, etc.)
	Version string // version, ref, tag, etc.
	Raw     string // original input
}

func ParseSource(input string) (*SourceInfo, error) {
	if input == "" {
		return nil, fmt.Errorf("empty source")
	}

	info := &SourceInfo{Raw: input}

	// Handle file:// scheme as local
	if strings.HasPrefix(input, "file://") {
		info.Type = SourceLocal
		info.Source = strings.TrimPrefix(input, "file://")
		return info, nil
	}

	// Check if it looks like a local path
	if isLocalPath(input) {
		info.Type = SourceLocal
		info.Source = input
		return info, nil
	}

	var baseInput, version string
	if atIndex := strings.LastIndex(input, "@"); atIndex != -1 {
		baseInput = input[:atIndex]
		version = input[atIndex+1:]
	} else {
		baseInput = input
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(baseInput)
	if err != nil {
		return nil, fmt.Errorf("invalid source format: %w", err)
	}

	// Extract scheme (type)
	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("missing scheme in source")
	}
	var sourceType SourceType
	switch strings.ToLower(parsedURL.Scheme) {
	case "http", "https":
		sourceType = SourceHTTP
	case "github":
		sourceType = SourceGitHub
	case "google-well-known":
		sourceType = SourceGoogleWellKnown
	case "protodex":
		sourceType = SourceProtodex
	case "file":
		sourceType = SourceLocal
	default:
		return nil, fmt.Errorf("unsupported source type: %s", parsedURL.Scheme)
	}

	info.Type = sourceType

	info.Source = parsedURL.Host + parsedURL.Path

	// Set version
	if version != "" {
		info.Version = version
	} else {
		info.Version = getDefaultVersion(info.Type)
	}

	return info, nil
}

func isLocalPath(input string) bool {
	return strings.HasPrefix(input, "./") ||
		strings.HasPrefix(input, "../") ||
		strings.HasPrefix(input, "/") ||
		(!strings.Contains(input, "://") && !strings.Contains(input, "@"))
}

func getDefaultVersion(sourceType SourceType) string {
	switch sourceType {
	case SourceGitHub:
		return "main"
	case SourceProtodex:
		return "latest"
	default:
		return ""
	}
}
