## Baseline Requirements

The goal for this project is to deliver an MCP-compatible reverse proxy that supports features
needed by security and operations teams to support agentic workflows in the enterprise. Call the
system OpenMCP.

At its core, OpenMCP is just a production grace HTTP server. It must be compatible with the latest
MCP HTTP transport layer, "Streamable HTTP".  This mode works by establishing "session IDs", which
are then used by the client and server to maintain state. Thus the system must allow for sticky
routing modes, based on the session ID, which is passed via header.

The system must expose a system for loading policies. The policies will allow the user to enforce
Tool and Resource access based on agent identity, as well as to enforce specific data access
requirements, i.e. disallowing certain responses from being returned by an MCP (e.g. based on the
presense of certain customer data.)

The system must expose all MCP activity via structured logs and metrics. The logs should provide all
relevant context, allowing downstream systems to replay and step through Agent flows. The obvious
use-cases here are auditability, debugging, and security scanning.

The system should bake-in an approval flow for human-in-the-loop (HITL) requests. E.g., consider an
organization with a policy that agents can access arbitrary support ticket text, but only after
human approval and only for a set length of time following approval. The system should allow this to
be expressed via policy. When an agent attempts to access a ticket, a notification is sent to a user
requesting approval. The user will either approve or deny the request, and the decision will be kept
in state until the access expires.

The system should be configurable via Kubernetes CRDs, but the configuration should be kept
pluggable, so that it can be used in non-Kubernetes contexts.

The system should expose a pluggable architecture for state maintenance. E.g., in-memory for demo
deployments, redis, DynamoDB, etc.

## Proxy Requirements

As a reverse proxy, OpenMCP should:
- Support serving multiple MCP servers out of a single deployment
  - e.g. mcp-gateway.internal.company-foo.com/github-mcp and /jira-mcp
- Support the latest "streamable HTTP" transport layer
  - It will not support the stdio-based transport layer, which is only used for local system access
- Support for mcp-session-id based "sticky" request routing
- Support for terminating TLS or operating in plaintext mode
- Support for terminating HTTP2

## Policy Requirements

As a policy enforcement engine, OpenMCP should:
- Allow users to specify which MCP servers an agent has access to
- Allow users to specify which Prompts, Resources, and Tools within a single MCP server an agent has
  access to
- Allow users to restrict arbitrary request or response patterns from being made to or returned from
  an MCP server, respectively
- Policies should be specifiable both at the MCP server level, as well as at the (agent, MCP server)
  level.

## Identity Requirements

In order to support the Policy Requirements detailed above, OpenMCP should:
- Allow users to configure a trusted source of agent identity
- This could be based on mTLS, SPIFFE, static credentials, service mesh features, shared secrets,
  etc.

## Observability Requirements

In order to maintain observability and provide a rich and complete data source, OpenMCP should:
- Log all requests, responses, and intermediary messages in such a way that it is easy to
  reconstruct the full interaction log for a given agent or session id
- Expose metrics that allow agent developers to understand their usage patterns
- Expose metrics that allow agent developers to monitor signals such as error rates, request rates,
  etc
- Expose metrics that allow operations teams tasked with running OpenMCP easy insight into
  performance characteristics and behavior

## HITL Flow Requirements

In order to provide differentiated value and standardize the HITL experience, OpenMCP should:
- Implement a type of policy action that triggers a human review flow
- The human review flow should be standardized, i.e. it should look and feel the same regardless of
  origin
- Familiar tools, such as slack or email, should be easily configurable for notification and action
  taking
