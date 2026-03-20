package client

import (
	"net/url"
	"regexp"
	"strings"
)

// EncodePathSegment percent-encodes a single URL path segment at the HTTP layer.
func EncodePathSegment(s string) string {
	return url.PathEscape(s)
}

// uuidFirstBlockSevenHex matches a UUID-shaped string whose first block has only 7 hex digits
// (35 chars total) — some APIs omit a leading 0 in the first octet.
var uuidFirstBlockSevenHex = regexp.MustCompile(`^(?i)[0-9a-f]{7}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// JobPathSegment prepares a job ID for /jobs/{id}/... paths.
// It trims space, normalizes Unicode hyphen-like runes to ASCII '-', and if the value matches a
// 35-character UUID pattern with a 7-hex first block, prepends '0' (restores omitted leading zero).
// Normal 36-character UUIDs are left unchanged. Prefer this over PathEscape for job IDs so the path
// matches strict UUID parsers on the server. Callers should still validate untrusted input (CLI does).
func JobPathSegment(id string) string {
	id = strings.TrimSpace(id)
	id = normalizeDashesToASCII(id)
	if uuidFirstBlockSevenHex.MatchString(id) {
		return "0" + id
	}
	return id
}

func normalizeDashesToASCII(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2212':
			b.WriteByte('-')
		case '\u00ad':
			// soft hyphen — omit
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
