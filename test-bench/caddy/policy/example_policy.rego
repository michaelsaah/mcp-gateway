package authz

import rego.v1

default allow := false

allow if {
    input.body.method != "tools/call"   # allow everything that isn't tools/call
}

allow if {
    input.body.method == "tools/call"
    not input.body.params.name == "create_branch"  # allow tools/call unless it's create_branch
}

allow if {
    input.body.method == "tools/call"
    input.body.params.name == "create_branch"
    startswith(input.body.params.arguments.branch, "agent-identity-")  # allow only if branch prefix matches
}
