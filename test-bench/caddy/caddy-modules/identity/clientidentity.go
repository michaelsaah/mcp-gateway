package clientidentity

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(ClientIdentity{})
	httpcaddyfile.RegisterHandlerDirective("client_identity", parseCaddyfile)
    httpcaddyfile.RegisterDirectiveOrder("client_identity", httpcaddyfile.After, "request_header")
}

// ClientIdentity implements an HTTP middleware that evaluates requests against OPA policies
type ClientIdentity struct {
	// The source of the client's identity - TODO should be an enum
	Source string `json:"source,omitempty"`
    HeaderConfig *HeaderConfig `json:"header,omitempty"`
}

type HeaderConfig struct {
    // The name of the header to use as the identity source
    Name string
}

func (ClientIdentity) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.client_identity",
		New: func() caddy.Module { return new(ClientIdentity) },
	}
}

// Provision implements caddy.Provisioner
func (ci *ClientIdentity) Provision(ctx caddy.Context) error {
	if ci.Source == "" {
		return fmt.Errorf("source is required")
	}

    if ci.HeaderConfig == nil {
		return fmt.Errorf("header_config is required")
    } else {
        if ci.HeaderConfig.Name == "" {
		    return fmt.Errorf("header_config.name is required")
        }
    }

	return nil
}

// Cleanup implements caddy.CleanerUpper
func (ci *ClientIdentity) Cleanup() error {
	// No cleanup needed for rego queries
	return nil
}

// Validate implements caddy.Validator
func (ci *ClientIdentity) Validate() error {
    // TODO
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (ci *ClientIdentity) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    // Fetch replacer
    repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	// Fetch and clean headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[strings.ToLower(key)] = values[0]
		}
	}

    identity := headers[ci.HeaderConfig.Name]

	repl.Set("mcproxy.identity.header", identity) 

	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (ci *ClientIdentity) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	for d.NextBlock(0) {
		switch d.Val() {
		case "source":
			if !d.NextArg() {
				return d.ArgErr()
			}
			ci.Source = d.Val()
		case "header":
			if ci.HeaderConfig == nil {
				ci.HeaderConfig = &HeaderConfig{}
			}
			for d.NextBlock(1) {
				switch d.Val() {
				case "name":
					if !d.NextArg() {
						return d.ArgErr()
					}
					ci.HeaderConfig.Name = d.Val()
				default:
					return d.Errf("unknown header subdirective: %s", d.Val())
				}
			}
		default:
			return d.Errf("unknown subdirective: %s", d.Val())
		}
	}

	return nil
}

// parseCaddyfile unmarshals tokens from h into a new ClientIdentity middleware
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var ci ClientIdentity
	err := ci.UnmarshalCaddyfile(h.Dispenser)
	return &ci, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ClientIdentity)(nil)
	_ caddy.CleanerUpper          = (*ClientIdentity)(nil)
	_ caddy.Validator             = (*ClientIdentity)(nil)
	_ caddyhttp.MiddlewareHandler = (*ClientIdentity)(nil)
	_ caddyfile.Unmarshaler       = (*ClientIdentity)(nil)
)
