package datasource

import (
	"net/url"
	"strings"
)

// grafanaResourceQueryKeys are query parameters Grafana adds to /api/datasources/.../resources/...
// requests. They must not be forwarded to the Gcore API.
var grafanaResourceQueryKeys = []string{
	"accesscontrol",
}

// resourceSubpath returns the path segment after /resources/ (e.g. "iam/users/me"),
// without a query string. If Path is empty, it parses the path from URL.
func resourceSubpath(pathField, urlField string) string {
	p := strings.TrimSpace(pathField)
	p = strings.TrimPrefix(p, "/")
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimPrefix(strings.TrimPrefix(p, "resources/"), "/")

	if p != "" {
		return p
	}

	u := strings.TrimSpace(urlField)
	if i := strings.IndexByte(u, '?'); i >= 0 {
		u = u[:i]
	}
	u = strings.TrimPrefix(u, "/")
	if j := strings.Index(u, "/resources/"); j >= 0 {
		u = u[j+len("/resources/"):]
	} else {
		u = strings.TrimPrefix(u, "resources/")
	}
	if i := strings.IndexByte(u, '?'); i >= 0 {
		u = u[:i]
	}
	return strings.TrimPrefix(u, "/")
}

// upstreamQueryFromResourceRequest builds the query string to append to the Gcore URL,
// excluding Grafana-internal parameters.
func upstreamQueryFromResourceRequest(pathField, urlField string) string {
	raw := ""
	if i := strings.IndexByte(urlField, '?'); i >= 0 {
		raw = urlField[i+1:]
	} else if i := strings.IndexByte(pathField, '?'); i >= 0 {
		raw = pathField[i+1:]
	}
	if raw == "" {
		return ""
	}
	q, err := url.ParseQuery(raw)
	if err != nil || len(q) == 0 {
		return ""
	}
	for _, k := range grafanaResourceQueryKeys {
		q.Del(k)
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}
