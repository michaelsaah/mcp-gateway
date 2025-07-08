## Implementation Ideas

Caddy is the core. Should support all of the reverse proxy requirements out of the box.

OPA is the policy engine. can use the [SDK](https://pkg.go.dev/github.com/open-policy-agent/opa@v1.6.0/v1/sdk) via a Caddy module to eveluate policies on requests/responses.

Simplest way to start with this is probably separat


