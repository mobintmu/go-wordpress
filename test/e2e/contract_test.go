package e2e

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"go-wordpress/api"
	"go-wordpress/internal/diagnose"

	"go.yaml.in/yaml/v3"
)

// spec is api/openapi.yaml decoded into generic maps.
var (
	specOnce sync.Once
	spec     map[string]any
	specErr  error
)

func loadSpec(t *testing.T) map[string]any {
	t.Helper()
	specOnce.Do(func() { specErr = yaml.Unmarshal(api.OpenAPI, &spec) })
	if specErr != nil {
		t.Fatalf("api/openapi.yaml is not valid YAML: %v", specErr)
	}
	return spec
}

func m(v any) map[string]any { x, _ := v.(map[string]any); return x }

// responseSchema finds the JSON schema for method path status.
func responseSchema(t *testing.T, method, path string, status int) (map[string]any, bool) {
	paths := m(loadSpec(t)["paths"])
	op := m(m(paths[path])[strings.ToLower(method)])
	if op == nil {
		return nil, false
	}
	resp := m(m(op["responses"])[strconv.Itoa(status)])
	if ref, ok := resp["$ref"].(string); ok {
		resp = resolveRef(t, ref)
	}
	if resp == nil {
		return nil, true
	}
	return m(m(m(resp["content"])["application/json"])["schema"]), true
}

func resolveRef(t *testing.T, ref string) map[string]any {
	var cur any = loadSpec(t)
	for _, seg := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		cur = m(cur)[seg]
	}
	if cur == nil {
		t.Fatalf("unresolved $ref %s", ref)
	}
	return m(cur)
}

// checkContract validates a live response against the spec. Paths the
// spec does not describe (unknown routes, Swagger UI assets) are skipped,
// but a status code the spec does not list for a described path fails.
func checkContract(t *testing.T, method, path string, r *Result) {
	t.Helper()
	schema, described := responseSchema(t, method, path, r.Status)
	if !described {
		return
	}
	paths := m(loadSpec(t)["paths"])
	responses := m(m(m(paths[path])[strings.ToLower(method)])["responses"])
	if _, listed := responses[strconv.Itoa(r.Status)]; !listed {
		t.Errorf("contract: %s %s returned HTTP %d, which openapi.yaml does not list", method, path, r.Status)
		return
	}
	if schema == nil || r.JSON == nil {
		return
	}
	var doc any
	if err := json.Unmarshal(r.Raw, &doc); err != nil {
		t.Errorf("contract: %s %s: body is not JSON: %v", method, path, err)
		return
	}
	for _, e := range validate(t, schema, doc, "$") {
		t.Errorf("contract: %s %s HTTP %d: %s\nbody: %s", method, path, r.Status, e, r.Raw)
	}
}

// validate checks v against a subset of OpenAPI 3.0 schema rules: $ref,
// allOf, type, nullable, enum, required, properties, additionalProperties,
// items, and the date-time and uri formats. Properties the schema does not
// list are errors unless additionalProperties allows them, so the code
// cannot return fields the contract does not describe.
func validate(t *testing.T, s map[string]any, v any, at string) []string {
	if ref, ok := s["$ref"].(string); ok {
		return validate(t, resolveRef(t, ref), v, at)
	}
	var errs []string
	if all, ok := s["allOf"].([]any); ok {
		if v == nil && s["nullable"] == true {
			return nil
		}
		for _, sub := range all {
			errs = append(errs, validate(t, m(sub), v, at)...)
		}
		return errs
	}
	if v == nil {
		if s["nullable"] == true || s["type"] == nil {
			return nil
		}
		return []string{at + ": null is not allowed"}
	}
	if enum, ok := s["enum"].([]any); ok {
		found := false
		for _, e := range enum {
			if fmt.Sprint(e) == fmt.Sprint(v) {
				found = true
			}
		}
		if !found {
			errs = append(errs, fmt.Sprintf("%s: %v is not one of %v", at, v, enum))
		}
	}
	switch s["type"] {
	case "object":
		obj, ok := v.(map[string]any)
		if !ok {
			return append(errs, at+": want object")
		}
		props := m(s["properties"])
		for _, r := range asList(s["required"]) {
			if _, ok := obj[r]; !ok {
				errs = append(errs, fmt.Sprintf("%s: missing required %q", at, r))
			}
		}
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if ps, ok := props[k]; ok {
				errs = append(errs, validate(t, m(ps), obj[k], at+"."+k)...)
				continue
			}
			switch ap := s["additionalProperties"].(type) {
			case map[string]any:
				errs = append(errs, validate(t, ap, obj[k], at+"."+k)...)
			case bool:
				if !ap {
					errs = append(errs, fmt.Sprintf("%s: unexpected property %q", at, k))
				}
			default:
				errs = append(errs, fmt.Sprintf("%s: property %q is not in the contract", at, k))
			}
		}
	case "array":
		arr, ok := v.([]any)
		if !ok {
			return append(errs, at+": want array")
		}
		if items := m(s["items"]); items != nil {
			for i, it := range arr {
				errs = append(errs, validate(t, items, it, fmt.Sprintf("%s[%d]", at, i))...)
			}
		}
	case "string":
		str, ok := v.(string)
		if !ok {
			// YAML examples decode timestamps as time.Time.
			if _, isTime := v.(time.Time); isTime && s["format"] == "date-time" {
				return errs
			}
			return append(errs, fmt.Sprintf("%s: want string, got %T", at, v))
		}
		switch s["format"] {
		case "date-time":
			if _, err := time.Parse(time.RFC3339, str); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %q is not date-time", at, str))
			}
		case "uri":
			if u, err := url.Parse(str); err != nil || !u.IsAbs() {
				errs = append(errs, fmt.Sprintf("%s: %q is not an absolute URI", at, str))
			}
		}
	case "integer":
		switch n := v.(type) {
		case float64:
			if n != float64(int64(n)) {
				errs = append(errs, fmt.Sprintf("%s: %v is not an integer", at, v))
			}
		case int, int64, uint64:
		default:
			errs = append(errs, fmt.Sprintf("%s: want integer, got %T", at, v))
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			errs = append(errs, fmt.Sprintf("%s: want boolean, got %T", at, v))
		}
	}
	return errs
}

func asList(v any) []string {
	xs, _ := v.([]any)
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, fmt.Sprint(x))
	}
	return out
}

// TestContractExamplesMatchSchemas checks that every named example in
// openapi.yaml matches its own schema.
func TestContractExamplesMatchSchemas(t *testing.T) {
	sp := loadSpec(t)
	if sp["openapi"] != "3.0.3" {
		t.Fatalf("openapi version = %v", sp["openapi"])
	}
	count := 0
	for path, item := range m(sp["paths"]) {
		for method, opAny := range m(item) {
			op := m(opAny)
			if rb := m(op["requestBody"]); rb != nil {
				c := m(m(rb["content"])["application/json"])
				for _, e := range validate(t, m(c["schema"]), c["example"], "request") {
					t.Errorf("%s %s request example: %s", method, path, e)
				}
				count++
			}
			for status, respAny := range m(op["responses"]) {
				resp := m(respAny)
				if ref, ok := resp["$ref"].(string); ok {
					resp = resolveRef(t, ref)
				}
				c := m(m(resp["content"])["application/json"])
				if c == nil {
					continue
				}
				schema := m(c["schema"])
				if ex, ok := c["example"]; ok {
					count++
					for _, e := range validate(t, schema, ex, "example") {
						t.Errorf("%s %s %s example: %s", method, path, status, e)
					}
				}
				for name, ex := range m(c["examples"]) {
					count++
					for _, e := range validate(t, schema, m(ex)["value"], name) {
						t.Errorf("%s %s %s example %s: %s", method, path, status, name, e)
					}
				}
			}
		}
	}
	if count < 10 {
		t.Fatalf("only %d examples checked; the spec lost its examples", count)
	}
}

// TestContractListsAllFailureCodes keeps the error.code description and
// the code table in sync.
func TestContractListsAllFailureCodes(t *testing.T) {
	desc := fmt.Sprint(m(m(resolveRef(t, "#/components/schemas/ExtractError")["properties"])["code"])["description"])
	for _, c := range diagnose.Codes() {
		if !strings.Contains(desc, c) {
			t.Errorf("failure code %s is not listed in ExtractError.code", c)
		}
	}
}

// TestContractServedFileMatchesRepository checks /openapi.yaml serves the
// embedded file byte for byte.
func TestContractServedFileMatchesRepository(t *testing.T) {
	e := Start(t, Options{Env: opsEnv(t)})
	r := e.Get("/openapi.yaml", map[string]string{"Authorization": opsAuthHeader()})
	if r.Status != 200 || string(r.Raw) != string(api.OpenAPI) {
		t.Fatalf("served spec differs from api/openapi.yaml: HTTP %d", r.Status)
	}
}
