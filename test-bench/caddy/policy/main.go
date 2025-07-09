package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/open-policy-agent/opa/rego"
)

func init() {
	caddy.RegisterModule(Policy{})
	httpcaddyfile.RegisterHandlerDirective("opa_policy", parseCaddyfile)
}

// OPAInput represents the input structure sent to OPA for policy evaluation
type OPAInput struct {
	Method  string            `json:"method"`
	Path    []string          `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Query   map[string]string `json:"query"`
}

// Policy implements an HTTP middleware that evaluates requests against OPA policies
type Policy struct {
	// The path to the policy bundle directory or file
	BundlePath string `json:"bundle_path,omitempty"`
	// The OPA decision path to evaluate (e.g., "authz/allow")
	DecisionPath string `json:"decision_path,omitempty"`

	query rego.PreparedEvalQuery
}

func (Policy) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.opa_policy",
		New: func() caddy.Module { return new(Policy) },
	}
}

// Provision implements caddy.Provisioner
func (p *Policy) Provision(ctx caddy.Context) error {
	if p.BundlePath == "" {
		return fmt.Errorf("bundle_path is required")
	}

	if p.DecisionPath == "" {
		p.DecisionPath = "authz/allow"
	}

	// Read the policy file
	policyBytes, err := os.ReadFile(p.BundlePath)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	// Prepare the Rego query
	query, err := rego.New(
		rego.Query("data.authz.allow"),
		rego.Module("authz.rego", string(policyBytes)),
	).PrepareForEval(context.Background())

	if err != nil {
		return fmt.Errorf("failed to prepare rego query: %w", err)
	}

	p.query = query
	return nil
}

// Cleanup implements caddy.CleanerUpper
func (p *Policy) Cleanup() error {
	// No cleanup needed for rego queries
	return nil
}

// Validate implements caddy.Validator
func (p *Policy) Validate() error {
	if p.BundlePath == "" {
		return fmt.Errorf("bundle_path is required")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (p Policy) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Parse request body as JSON
	var body interface{}
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return fmt.Errorf("failed to read request body: %w", err)
		}

		// Reset body for downstream handlers
		r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// Parse JSON body if present
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &body); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return fmt.Errorf("failed to parse JSON body: %w", err)
			}
		}
	}

	// Split path into segments, removing empty segments
	pathSegments := []string{}
	for _, segment := range strings.Split(strings.Trim(r.URL.Path, "/"), "/") {
		if segment != "" {
			pathSegments = append(pathSegments, segment)
		}
	}

	// Parse query parameters
	query := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}

	// Prepare headers map
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[strings.ToLower(key)] = values[0]
		}
	}

	// Prepare input for OPA evaluation
	input := OPAInput{
		Method:  r.Method,
		Path:    pathSegments,
		Headers: headers,
		Body:    body,
		Query:   query,
	}

	// Evaluate policy using rego
	results, err := p.query.Eval(r.Context(), rego.EvalInput(input))
	if err != nil {
		http.Error(w, "Policy evaluation failed", http.StatusInternalServerError)
		return fmt.Errorf("Rego evaluation failed: %w", err)
	}

	// Check if the decision allows the request
	allowed := false
	if len(results) > 0 && len(results[0].Expressions) > 0 {
		// If result is undefined (nil), treat as false (deny)
		// If result is defined and true, allow
		val := results[0].Expressions[0].Value
		if val != nil {
			if b, ok := val.(bool); ok && b {
				allowed = true
			}
		}
	}

	if !allowed {
		return caddyhttp.Error(http.StatusForbidden, fmt.Errorf("request denied"))
	}

	// Request is allowed, continue to next handler
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (p *Policy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	for d.NextBlock(0) {
		switch d.Val() {
		case "bundle_path":
			if !d.NextArg() {
				return d.ArgErr()
			}
			p.BundlePath = d.Val()
		case "decision_path":
			if !d.NextArg() {
				return d.ArgErr()
			}
			p.DecisionPath = d.Val()
		default:
			return d.Errf("unknown subdirective: %s", d.Val())
		}
	}

	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Policy middleware
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var p Policy
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return &p, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Policy)(nil)
	_ caddy.CleanerUpper          = (*Policy)(nil)
	_ caddy.Validator             = (*Policy)(nil)
	_ caddyhttp.MiddlewareHandler = (*Policy)(nil)
	_ caddyfile.Unmarshaler       = (*Policy)(nil)
)
