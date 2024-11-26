package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// words that automatically pass the filter
var whitelistWords = []string{
	"/api", "api/", "/rest", "rest/", "search", "render", "admin", "redirect",
	"proxy", "iframe", "ajax", "jsonp", "tmp", "src", "callback", "private",
	"login", "logout", "signin", "graphql", "user", "signout", "signup",
	"register", "eyj",
}

// patterns that automatically pass the filter
var specialPatterns = []string{"http:", "https:", "http%3A", "https%3A"}

// Extensions that are allowed
var allowedExtensions = []string{
	".htm", ".php", ".phtml", ".php2", ".php3", ".asp", ".aspx", ".asmx", ".ashx",
	".jsp", ".cgi", ".pl", ".json", ".cmk", ".yaml", ".yml", ".do", ".jspa", ".sh",
	".zip", ".cfm", ".txt", ".rar", ".tar", ".gz", ".bak", ".env", ".conf", ".xml",
	".js", ".bin",
}

var ignoredQueryParams = regexp.MustCompile(`^(v|utm_.*)$`)

// each URL and Parameter Combination is printed this many times
var thresholdCombinations = 2

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	seen := make(map[string]int)

	for scanner.Scan() {
		rawURL := scanner.Text()

		// replace &amp; with &
		rawURL = strings.Replace(rawURL, "&amp;", "&", -1)

		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		if urlPassesFilter(u, rawURL) {
			pp := make([]string, 0)
			for p := range u.Query() {
				pp = append(pp, p)
			}
			sort.Strings(pp)

			key := fmt.Sprintf("%s%s?%s", u.Hostname(), u.EscapedPath(), strings.Join(pp, "&"))
			if seen[key] < thresholdCombinations {
				fmt.Println(rawURL)
				seen[key]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error when reading input:", err)
		os.Exit(1)
	}
}

func urlPassesFilter(parsed *url.URL, rawURL string) bool {

	// Comprobar palabras o patrones
	if containsAny(rawURL, whitelistWords) {
		return true
	}

	// Comprobar patrones especiales fuera del esquema inicial
	if containsSpecialPatterns(parsed) {
		return true
	}

	// Comprobar extensiones
	if hasAllowedExtension(parsed.Path) {
		return true
	}

	// Comprobar parámetros de query
	if hasInterestingQueryParams(parsed) {
		return true
	}

	return false
}

func containsAny(input string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}

func containsSpecialPatterns(parsed *url.URL) bool {
	urlWithoutScheme := strings.TrimPrefix(parsed.String(), parsed.Scheme+"://")
	for _, pattern := range specialPatterns {
		if strings.Contains(urlWithoutScheme, pattern) {
			return true
		}
	}
	return false
}

func hasInterestingQueryParams(parsed *url.URL) bool {
	query := parsed.Query()
	for param := range query {
		if !ignoredQueryParams.MatchString(param) {
			return true
		}
	}
	return false
}

func hasAllowedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}
