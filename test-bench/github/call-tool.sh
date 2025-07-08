curl http://localhost:6000/mcp/ -H "Authorization: bearer $GITHUB_PAT" -i -X POST -H "content-type: application/json" -d @- << EOF
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_branch",
    "arguments": {
      "branch-name": "prod"
    }
  }
}
EOF
