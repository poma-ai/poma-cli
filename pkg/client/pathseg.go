package client

import "net/url"

// EncodePathSegment percent-encodes a single URL path segment at the HTTP layer.
func EncodePathSegment(s string) string {
	return url.PathEscape(s)
}
