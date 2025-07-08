curl https://api.githubcopilot.com/mcp/ -H "Authorization: bearer $GITHUB_PAT" -i -X POST -H "content-type: application/json" -d '{"jsonrpc": "2.0", "id": 1, "method": "prompts/list", "params": {}}'
