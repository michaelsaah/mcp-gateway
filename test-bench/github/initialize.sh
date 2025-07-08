curl http://localhost:6000/mcp/ -H "Authorization: bearer $GITHUB_PAT" -i -X POST -H "content-type: application/json" -d @- << EOF
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-06-18",
    "capabilities": {
      "roots": {
        "listChanged": true
      },
      "sampling": {},
      "elicitation": {}
    },
    "clientInfo": {
      "name": "my-client",
      "title": "My very excellent MCP client",
      "version": "0.0.1"
    }
  }
}
EOF

curl http://localhost:6000/mcp/ -H "Authorization: bearer $GITHUB_PAT" -i -X POST -H "content-type: application/json" -d @- << EOF
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized"
}
EOF

curl http://localhost:6000/mcp/ -H "Authorization: bearer $GITHUB_PAT" -i
