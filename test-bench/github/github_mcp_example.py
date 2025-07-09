#!/usr/bin/env python3
"""
Example agent flow using GitHub MCP server with fastmcp.

This script demonstrates a typical workflow where an AI agent:
1. Connects to the GitHub MCP server
2. Gets user context information
3. Lists repositories for the authenticated user
4. Searches for issues in a repository
5. Creates a new issue
6. Lists pull requests

You'll need to set your GitHub Personal Access Token as an environment variable:
export GITHUB_PAT=your_token_here
"""

import asyncio
import json
import os
from fastmcp import Client


async def main():
    # Check for required environment variable
    if not os.getenv("GITHUB_PAT"):
        print("Error: GITHUB_PAT environment variable is required")
        print("Please set it with: export GITHUB_PAT=your_token_here")
        return

    print("🚀 Starting GitHub MCP Agent Flow Demo")
    print("=" * 50)

    # Configure GitHub MCP server
    config = {
        "mcpServers": {
            "github": {
                #"url": "https://api.githubcopilot.com/mcp/",
                "url": "http://localhost:6000/mcp/",
                "headers": {
                    "Authorization": f"Bearer {os.getenv('GITHUB_PAT')}"
                }
            }
        }
    }

    # The whole workflow here is, get a repo, create a new branch to work in
    # As policy, we want to enforce that agent-created branches start with 
    # `agent-identity-<id>/`.
    try:
        async with Client(config) as client:
            print("Getting user context...")
            try:
                get_me_res = await client.call_tool("get_me", {})
                my_info = json.loads(get_me_res.content[0].text)
                print(f"Authenticated as: {my_info['login']}")
            except Exception as e:
                print(f"Could not get user context: {e}")

            try:
                #import pdb; pdb.set_trace()
                create_branch_res = await client.call_tool("create_branch", {
                    #"branch": "agent-identity-123/fix-bug-456",
                    "branch": "my-bad-branch",
                    "owner": "michaelsaah",
                    "repo": "numpy"
                })
                print(create_branch_res)
            except Exception as e:
                print(f"Could not create branch: {e}")

    except Exception as e:
        print(f"❌ Error connecting to GitHub MCP server: {e}")
        print("Make sure you have:")
        print("1. Set GITHUB_PAT environment variable")
        print("2. The GitHub MCP server is accessible")
        print("3. Your token has the necessary permissions")

    print("\n✅ Agent flow completed!")
    print("=" * 50)


if __name__ == "__main__":
    asyncio.run(main())
