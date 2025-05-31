// Package sshconfig provides utilities for parsing and validating SSH configuration files.
//
//nolint:lll // using long lines in this file
package sshconfig

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/grafviktor/goto/internal/logger"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// Regex to validate SSH usernames.
// tests:
// "admin",        // Valid.
// "user123",      // Valid.
// "root",         // Valid.
// "_sshuser",     // Valid (leading underscore allowed).
// "user-name",    // Valid (dash allowed).
// "user.name",    // Invalid (dot is risky).
// "123username",  // Invalid (cannot start with digit).
// "user@domain",  // Invalid (contains `@`).
// "invalid user", // Invalid (contains space).
// "user!",        // Invalid (special character).
var sshUsernameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]{0,31}$`)

/*
Valid hostname regex (RFC 1035 + RFC 1123).

"example.com",     // Valid.
"sub.example.com", // Valid.
"localhost",       // Valid (for local use).
"123.example",     // Valid.
"-invalid.com",    // Invalid (starts with `-`).
"example-.com",    // Invalid (ends with `-`).
"ex@mpl$.com",     // Invalid (special characters).
"superlonglabelnamethatiswaytoolongtobevalid.example.com", // Invalid (over 63 chars per label).
"toolong......................................................................................................com", // Invalid (over 253 chars).
*/
var hostnameRegex = regexp.MustCompile(`^(?i:[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*)$`)

func isNetworkPortNumberValid(port int) bool {
	return port >= 0 && port <= 65535
}

/*
Regex to match exactly two or more words.

"hello world",     // Valid.
"  foo   bar  ",   // Valid.
"oneword",         // Invalid.
"three word test", // Valid.
"",                // Invalid.
*/
var twoWordsRegex = regexp.MustCompile(`^(\S+)\s+(.+)$`)

func parseKeyValuesLine(line string) (string, string, error) {
	matches := twoWordsRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1], matches[2], nil
	}

	return "", "", errors.New("not a key value string")
}

func isTextFileMime(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger := logger.Get()
			logger.Error("[SSHCONFIG] Error closing file %s: %v", filename, closeErr)
		}
	}()

	buf := make([]byte, 512) // Read first 512 bytes
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	mimeType := http.DetectContentType(buf[:n])
	return strings.Contains(mimeType, "text/plain")
}
