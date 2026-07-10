#!/usr/bin/env python3
"""
Stitch Remote MCP Submitter for LaunchFleet
Interacts with the Google Stitch MCP server via mcp-remote proxy.

Usage:
  python3 stitch_submit.py --list
  python3 stitch_submit.py --create "Project Name"
  python3 stitch_submit.py --project-info <project_id>
  python3 stitch_submit.py --generate --project-id <project_id> --prompt "Prompt description..."
"""

import subprocess
import json
import time
import sys
import os
import urllib.request
import ssl

# ──────────────────────────────────────────────────────────────────────────────
# Config & API Key Resolution
# ──────────────────────────────────────────────────────────────────────────────

def _load_api_key():
    # 1. Environment variable STITCH_API_KEY
    key = os.environ.get("STITCH_API_KEY")
    if key:
        return key

    # 2. Check .env.local or .env for STITCH_API_KEY
    for envfile in [".env.local", ".env"]:
        path = os.path.join(os.path.dirname(os.path.abspath(__file__)), envfile)
        if os.path.exists(path):
            with open(path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("STITCH_API_KEY="):
                        return line.split("=", 1)[1].strip()

    # 3. Extract from mcp_config.json (Stitch configuration)
    mcp_config_path = os.path.expanduser("~/.gemini/antigravity/mcp_config.json")
    if os.path.exists(mcp_config_path):
        try:
            with open(mcp_config_path) as f:
                content = f.read()
            import re
            cleaned_content = re.sub(r'(?<!http:)(?<!https:)//.*', '', content)
            config = json.loads(cleaned_content)
            stitch_config = config.get("mcpServers", {}).get("stitch", {})
            headers = stitch_config.get("headers", {})
            key = headers.get("X-Goog-Api-Key") or headers.get("x-goog-api-key")
            if key:
                return key
        except Exception:
            pass

    # 4. Fallback to JULES_API_KEY
    key = os.environ.get("JULES_API_KEY")
    if key:
        return key

    for envfile in [".env.local", ".env"]:
        path = os.path.join(os.path.dirname(os.path.abspath(__file__)), envfile)
        if os.path.exists(path):
            with open(path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("JULES_API_KEY="):
                        return line.split("=", 1)[1].strip()

    print("❌ API key not found in environment, .env files, or mcp_config.json")
    sys.exit(1)


API_KEY = _load_api_key()
print(f"🔑 Using API Key: {API_KEY[:12]}...{API_KEY[-6:]}")

# ──────────────────────────────────────────────────────────────────────────────
# Stitch MCP Client Wrapper
# ──────────────────────────────────────────────────────────────────────────────

class StitchMCPClient:
    def __init__(self):
        print("🔗 Connecting to Stitch MCP Server via mcp-remote...")
        self.proc = subprocess.Popen(
            ["npx", "-y", "mcp-remote", "https://stitch.googleapis.com/mcp", "--header", f"X-Goog-Api-Key: {API_KEY}"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        # Wait for initialization
        time.sleep(6)
        if self.proc.poll() is not None:
            print("❌ Failed to start mcp-remote proxy!")
            print("Stderr:", self.proc.stderr.read())
            sys.exit(1)
        self.request_id = 1
        self._initialize_handshake()

    def _initialize_handshake(self):
        print("  → Sending MCP initialize request...")
        init_req = {
            "jsonrpc": "2.0",
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {
                    "name": "stitch-client",
                    "version": "1.0.0"
                }
            },
            "id": self.request_id
        }
        self._send_raw(init_req)
        init_res = self._read_raw()
        if init_res:
            print("  → Received initialize response.")
            initialized_notification = {
                "jsonrpc": "2.0",
                "method": "notifications/initialized"
            }
            self._send_raw(initialized_notification)
            print("  → Sent initialized notification. MCP session ready.")
        else:
            print("  ⚠️ Warning: No initialize response received from proxy.")

    def _send_raw(self, msg):
        self.proc.stdin.write(json.dumps(msg) + "\n")
        self.proc.stdin.flush()

    def _read_raw(self, timeout=10.0):
        import select
        r, _, _ = select.select([self.proc.stdout], [], [], timeout)
        if r:
            line = self.proc.stdout.readline()
            return json.loads(line)
        return None

    def call_tool(self, name, arguments=None):
        self.request_id += 1
        req = {
            "jsonrpc": "2.0",
            "method": "tools/call",
            "params": {
                "name": name,
                "arguments": arguments or {}
            },
            "id": self.request_id
        }
        self.proc.stdin.write(json.dumps(req) + "\n")
        self.proc.stdin.flush()

        while True:
            line = self.proc.stdout.readline()
            if not line:
                return None
            try:
                data = json.loads(line)
                if "id" in data and data["id"] == self.request_id:
                    return data
            except Exception:
                pass

    def close(self):
        self.proc.terminate()
        self.proc.wait()


# ──────────────────────────────────────────────────────────────────────────────
# CLI Logic
# ──────────────────────────────────────────────────────────────────────────────

def main():
    args = sys.argv[1:]

    if not args or "--help" in args or "-h" in args:
        print(__doc__)
        sys.exit(0)

    # 1. List Projects
    if "--list" in args:
        client = StitchMCPClient()
        try:
            print("Listing Stitch projects...")
            res = client.call_tool("list_projects")
            if not res or "result" not in res:
                print("❌ Failed to list projects:", res)
                sys.exit(1)
            
            # Extract content text
            content_text = res["result"]["content"][0]["text"]
            data = json.loads(content_text)
            projects = data.get("projects", [])
            print(f"\n📊 Projects found: {len(projects)}")
            for p in projects:
                name = p.get("name", "Unknown")
                title = p.get("title", "No Title")
                project_id = name.split("/")[-1]
                print(f"  • ID: {project_id:<25} | Title: {title}")
        finally:
            client.close()
        sys.exit(0)

    # 2. Create Project
    if "--create" in args:
        idx = args.index("--create")
        if idx + 1 >= len(args):
            print("❌ Please specify a project title.")
            sys.exit(1)
        title = args[idx + 1]

        client = StitchMCPClient()
        try:
            print(f"Creating project: '{title}'...")
            res = client.call_tool("create_project", {"title": title})
            if not res or "result" not in res:
                print("❌ Failed to create project:", res)
                sys.exit(1)

            content_text = res["result"]["content"][0]["text"]
            data = json.loads(content_text)
            project_name = data.get("name", "")
            project_id = project_name.split("/")[-1]
            print(f"✅ Success! Project Created. ID: {project_id}")
        finally:
            client.close()
        sys.exit(0)

    # 3. Project Info
    if "--project-info" in args:
        idx = args.index("--project-info")
        if idx + 1 >= len(args):
            print("❌ Please specify a project ID.")
            sys.exit(1)
        project_id = args[idx + 1]

        client = StitchMCPClient()
        try:
            print(f"Fetching info for project '{project_id}'...")
            res = client.call_tool("get_project", {"name": f"projects/{project_id}"})
            if not res or "result" not in res:
                print("❌ Failed to get project info:", res)
                sys.exit(1)

            content_text = res["result"]["content"][0]["text"]
            data = json.loads(content_text)
            
            print(f"\n📁 Project: {data.get('title', 'Untitled')}")
            print(f"  Status: {data.get('status', 'Unknown')}")
            print(f"  Created: {data.get('createTime', 'Unknown')}")

            screens = data.get("screens", [])
            print(f"\n🖥️  Screens ({len(screens)}):")
            for s in screens:
                name = s.get("name", "")
                screen_id = name.split("/")[-1]
                title = s.get("title", "Untitled")
                print(f"  • Screen ID: {screen_id:<20} | Title: {title}")
        finally:
            client.close()
        sys.exit(0)

    # 4. Generate Screen
    if "--generate" in args:
        if "--project-id" not in args or "--prompt" not in args:
            print("❌ Make sure to supply both --project-id and --prompt.")
            sys.exit(1)
        
        pid_idx = args.index("--project-id")
        project_id = args[pid_idx + 1]

        prompt_idx = args.index("--prompt")
        prompt = args[prompt_idx + 1]

        client = StitchMCPClient()
        try:
            print(f"🚀 Generating screen in project '{project_id}' using Gemini 3.1 Pro...")
            res = client.call_tool("generate_screen_from_text", {
                "projectId": project_id,
                "modelId": "GEMINI_3_1_PRO",
                "deviceType": "DESKTOP",
                "prompt": prompt
            })

            if not res or "result" not in res:
                print("❌ Generation failed:", res)
                sys.exit(1)

            content_text = res["result"]["content"][0]["text"]
            try:
                data = json.loads(content_text)
            except json.JSONDecodeError:
                print("❌ Received non-JSON response from server:")
                print(content_text)
                sys.exit(1)

            print("\n✅ Generation Complete!")
            download_url = None
            for component in data.get("outputComponents", []):
                design = component.get("design", {})
                for screen in design.get("screens", []):
                    html_code = screen.get("htmlCode", {})
                    if html_code and html_code.get("downloadUrl"):
                        download_url = html_code["downloadUrl"]
                        print(f"  • Screen: {screen.get('title', 'Untitled')}")
                        print(f"  • Download URL: {download_url}")
            
            if download_url:
                print("\n💡 Tip: You can download the html using curl or urllib.")
        finally:
            client.close()
        sys.exit(0)

    print("❌ Unknown arguments. Use --help to see usage.")


if __name__ == "__main__":
    main()
