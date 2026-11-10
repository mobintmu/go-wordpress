package diagnose

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

// MaxSampleBytes is the largest HTML sample returned in a failure.
const MaxSampleBytes = 16 * 1024

// Evidence describes what the service saw. Every field is optional.
type Evidence struct {
	FailedURL       string            `json:"failed_url,omitempty"`
	FinalURL        string            `json:"final_url,omitempty"`
	Redirects       []string          `json:"redirects"`
	HTTPStatus      int               `json:"http_status,omitempty"`
	Attempts        int               `json:"attempts,omitempty"`
	ContentType     string            `json:"content_type,omitempty"`
	BodyBytes       int64             `json:"body_bytes,omitempty"`
	BodySHA256      string            `json:"body_sha256,omitempty"`
	PageTitle       string            `json:"page_title,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	SelectorTrace   []TraceStep       `json:"selector_trace,omitempty"`
	HTMLSample      *HTMLSample       `json:"html_sample,omitempty"`
	FieldSamples    map[string]string `json:"field_samples,omitempty"`
}

// TraceStep is the match count for one prefix of a selector path.
type TraceStep struct {
	Selector string `json:"selector"`
	Matches  int    `json:"matches"`
}

// HTMLSample is a bounded piece of untrusted third-party HTML.
type HTMLSample struct {
	Reason         string `json:"reason"`
	Content        string `json:"content"`
	Truncated      bool   `json:"truncated"`
	OriginalLength int    `json:"original_length,omitempty"`
}

// Sample reasons.
const (
	ReasonBodyStart    = "Start of the response body"
	ReasonDeepestMatch = "Deepest element matched by the failing selector path"
	ReasonProductRoot  = "Product root element"
)

// NewSample truncates content to MaxSampleBytes on a UTF-8 boundary.
func NewSample(reason, content string) *HTMLSample {
	s := &HTMLSample{Reason: reason, OriginalLength: len(content)}
	if len(content) <= MaxSampleBytes {
		s.Content = content
		return s
	}
	cut := MaxSampleBytes
	for cut > 0 && !utf8.RuneStart(content[cut]) {
		cut--
	}
	s.Content = content[:cut]
	s.Truncated = true
	return s
}

// allowedHeaders are the only response headers returned as evidence.
// Set-Cookie is never returned.
var allowedHeaders = []string{"server", "content-type", "retry-after", "cf-ray", "cf-mitigated", "location"}

// FilterHeaders returns the allow-listed response headers.
func FilterHeaders(h http.Header) map[string]string {
	if h == nil {
		return nil
	}
	out := map[string]string{}
	for _, name := range allowedHeaders {
		if v := h.Get(name); v != "" {
			out[name] = truncate(v, 256)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// TruncateText shortens a field sample to n bytes.
func TruncateText(s string, n int) string { return truncate(strings.TrimSpace(s), n) }
