#!/usr/bin/env python3
"""Bridge script for Go parity tests.

Parses a JSONL session file using the same logic as extract_stats.py
and outputs structured JSON for comparison with Go's parser.

Usage: python test_parity_bridge.py <path_to_jsonl>
"""
import json
import sys
from pathlib import Path

# Import the extract module
sys.path.insert(0, str(Path(__file__).parent))
import extract_stats


def parse_jsonl_session(jsonl_path):
    """Parse a single JSONL file and return session stats as JSON."""
    sessions = {}

    with open(jsonl_path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except json.JSONDecodeError:
                continue

            msg_type = obj.get("type")
            session_id = obj.get("sessionId", "unknown")

            if session_id not in sessions:
                sessions[session_id] = {
                    "session_id": session_id,
                    "messages": 0,
                    "user_messages": 0,
                    "assistant_messages": 0,
                    "first_prompt": "",
                    "model_calls": {},
                    "total_cost": 0.0,
                    "total_input": 0,
                    "total_output": 0,
                    "total_cache_read": 0,
                    "total_cache_write": 0,
                    "tools": {},
                    "project_path": "",
                }

            sess = sessions[session_id]

            # Capture project path from cwd
            if obj.get("cwd") and not sess["project_path"]:
                sess["project_path"] = obj["cwd"]

            if msg_type == "user":
                sess["messages"] += 1
                sess["user_messages"] += 1

                # Extract first prompt (same filtering as Go)
                if not sess["first_prompt"]:
                    text = _extract_prompt_text(obj)
                    if (text
                            and not text.startswith("<command")
                            and not text.startswith("<local-command")
                            and not text.startswith("[Request interrupted")):
                        if len(text) > 200:
                            text = text[:200]
                        sess["first_prompt"] = text

            elif msg_type == "assistant":
                sess["messages"] += 1
                sess["assistant_messages"] += 1

                message = obj.get("message", {})
                model = message.get("model", "unknown")
                usage = message.get("usage", {})

                output_tokens = usage.get("output_tokens", 0)
                if output_tokens > 0:
                    input_tokens = usage.get("input_tokens", 0)
                    cache_read = usage.get("cache_read_input_tokens", 0)
                    cache_creation = usage.get("cache_creation_input_tokens", 0)

                    sess["model_calls"][model] = sess["model_calls"].get(model, 0) + 1
                    sess["total_input"] += input_tokens
                    sess["total_output"] += output_tokens
                    sess["total_cache_read"] += cache_read
                    sess["total_cache_write"] += cache_creation
                    sess["total_cost"] += extract_stats.calc_cost(model, usage)

                # Tool usage from content blocks
                content = message.get("content", [])
                if isinstance(content, list):
                    for block in content:
                        if isinstance(block, dict) and block.get("type") == "tool_use":
                            tool_name = block.get("name", "unknown")
                            sess["tools"][tool_name] = sess["tools"].get(tool_name, 0) + 1

    # Return the first (and typically only) session
    if sessions:
        return next(iter(sessions.values()))
    return {}


def _extract_prompt_text(obj):
    """Extract text from a user message, matching Go's extractPromptText."""
    message = obj.get("message", {})
    content = message.get("content")
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        for block in content:
            if isinstance(block, dict) and block.get("type") == "text":
                return block.get("text", "")
    return ""


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <jsonl_path>", file=sys.stderr)
        sys.exit(1)

    result = parse_jsonl_session(sys.argv[1])
    print(json.dumps(result))
