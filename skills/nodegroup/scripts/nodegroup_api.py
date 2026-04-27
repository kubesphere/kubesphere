#!/usr/bin/env python3
"""
NodeGroup API CLI - Authenticated wrapper for nodegroup KAPI operations.

Usage examples:
    python nodegroup_api.py login --username admin --password <password>
    python nodegroup_api.py request GET /kapis/infra.kubesphere.io/v1alpha1/nodegroups

    python nodegroup_api.py nodegroup list
    python nodegroup_api.py nodegroup get edge-a
    python nodegroup_api.py nodegroup create --name edge-a --alias "Edge A"
    python nodegroup_api.py nodegroup patch edge-a --description "updated"
    python nodegroup_api.py nodegroup delete edge-a

    python nodegroup_api.py bind node --nodegroup edge-a --node worker-1
    python nodegroup_api.py unbind namespace --nodegroup edge-a --namespace demo
"""

import argparse
import getpass
import json
import os
import sys
import time
from typing import Any, Dict, Optional

import requests


TOKEN_FILE = os.path.expanduser("~/.kubesphere_token")
DEFAULT_HOST = "http://ks-apiserver.kubesphere-system"
API_PREFIX = "/kapis/infra.kubesphere.io/v1alpha1"


def save_token(token: str, expires_at: Optional[int] = None) -> None:
    data = {"token": token, "expires_at": expires_at, "saved_at": int(time.time())}
    with open(TOKEN_FILE, "w", encoding="utf-8") as file:
        json.dump(data, file)


def load_token() -> Optional[str]:
    if not os.path.exists(TOKEN_FILE):
        return None

    try:
        with open(TOKEN_FILE, "r", encoding="utf-8") as file:
            data = json.load(file)
        if data.get("expires_at") and data["expires_at"] < int(time.time()):
            return None
        return data.get("token")
    except (OSError, json.JSONDecodeError):
        return None


def get_token(host: str, username: str, password: str) -> str:
    token_url = f"{host.rstrip('/')}/oauth/token"
    data = {
        "grant_type": "password",
        "username": username,
        "password": password,
        "client_id": "kubesphere",
        "client_secret": "kubesphere",
    }

    try:
        response = requests.post(
            token_url,
            data=data,
            headers={"Content-Type": "application/x-www-form-urlencoded"},
            timeout=30,
        )
        if response.status_code != 200:
            print(f"Error: Failed to get token. HTTP {response.status_code}", file=sys.stderr)
            print(response.text, file=sys.stderr)
            sys.exit(1)

        result = response.json()
        access_token = result.get("access_token")
        if not access_token:
            print("Error: No access_token in response", file=sys.stderr)
            sys.exit(1)

        expires_in = result.get("expires_in", 3600 * 5)
        save_token(access_token, int(time.time()) + expires_in)
        return access_token
    except requests.RequestException as exc:
        print(f"Error: {exc}", file=sys.stderr)
        sys.exit(1)


def require_token(args: argparse.Namespace) -> str:
    token = args.token or os.environ.get("KUBESPHERE_TOKEN")
    if token:
        return token

    token = load_token()
    if token:
        return token

    print("Error: No token available. Run `login` first or set KUBESPHERE_TOKEN.", file=sys.stderr)
    sys.exit(1)


def build_url(host: str, path: str) -> str:
    return f"{host.rstrip('/')}/{path.lstrip('/')}"


def emit_result(result: Any, quiet: bool = False) -> None:
    if quiet:
        print(json.dumps(result))
    else:
        print(json.dumps(result, indent=2))


def api_request(
    host: str,
    token: str,
    method: str,
    path: str,
    body: Optional[Any] = None,
    quiet: bool = False,
) -> None:
    url = build_url(host, path)
    content_type = "application/json"
    if method == "PATCH":
        if not isinstance(body, list):
            print("Error: PATCH requests must use a JSON Patch array body.", file=sys.stderr)
            sys.exit(1)
        content_type = "application/json-patch+json"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": content_type,
    }

    try:
        response = requests.request(
            method=method,
            url=url,
            headers=headers,
            json=body if body is not None else None,
            timeout=30,
        )
    except requests.RequestException as exc:
        print(f"Error: {exc}", file=sys.stderr)
        sys.exit(1)

    try:
        result = response.json()
    except json.JSONDecodeError:
        result = {"raw": response.text}

    if response.status_code >= 400:
        print(f"Error: HTTP {response.status_code}", file=sys.stderr)
        emit_result(result, quiet=False)
        sys.exit(1)

    emit_result(result, quiet=quiet)


def nodegroup_path(name: Optional[str] = None) -> str:
    return f"{API_PREFIX}/nodegroups" if name is None else f"{API_PREFIX}/nodegroups/{name}"


def bind_path(resource_type: str, nodegroup: str, target: str) -> str:
    if resource_type == "node":
        return f"{nodegroup_path(nodegroup)}/nodes/{target}"
    if resource_type == "namespace":
        return f"{nodegroup_path(nodegroup)}/namespaces/{target}"
    if resource_type == "workspace":
        return f"{nodegroup_path(nodegroup)}/workspaces/{target}"
    raise ValueError(f"unsupported bind type: {resource_type}")


def handle_request(args: argparse.Namespace) -> None:
    token = require_token(args)
    body = None
    if args.body and args.method in {"POST", "PUT", "PATCH"}:
        try:
            body = json.loads(args.body)
        except json.JSONDecodeError as exc:
            print(f"Error: Invalid JSON body: {exc}", file=sys.stderr)
            sys.exit(1)
    api_request(args.host, token, args.method, args.path, body=body, quiet=args.quiet)


def handle_nodegroup(args: argparse.Namespace) -> None:
    token = require_token(args)
    if args.nodegroup_action == "list":
        api_request(args.host, token, "GET", nodegroup_path(), quiet=args.quiet)
        return
    if args.nodegroup_action == "get":
        api_request(args.host, token, "GET", nodegroup_path(args.name), quiet=args.quiet)
        return
    if args.nodegroup_action == "create":
        body = {
            "apiVersion": "infra.kubesphere.io/v1alpha1",
            "kind": "NodeGroup",
            "metadata": {"name": args.name},
            "spec": {},
        }
        if args.alias:
            body["spec"]["alias"] = args.alias
        if args.description:
            body["spec"]["description"] = args.description
        if args.manager:
            body["spec"]["manager"] = args.manager
        api_request(args.host, token, "POST", nodegroup_path(), body=body, quiet=args.quiet)
        return
    if args.nodegroup_action == "patch":
        patch_ops = []
        if args.alias is not None:
            patch_ops.append({"op": "add", "path": "/spec/alias", "value": args.alias})
        if args.description is not None:
            patch_ops.append({"op": "add", "path": "/spec/description", "value": args.description})
        if args.manager is not None:
            patch_ops.append({"op": "add", "path": "/spec/manager", "value": args.manager})
        if not patch_ops:
            print("Error: patch requires at least one of --alias, --description, --manager", file=sys.stderr)
            sys.exit(1)
        api_request(args.host, token, "PATCH", nodegroup_path(args.name), body=patch_ops, quiet=args.quiet)
        return
    if args.nodegroup_action == "delete":
        api_request(args.host, token, "DELETE", nodegroup_path(args.name), quiet=args.quiet)
        return


def handle_bind(args: argparse.Namespace, method: str) -> None:
    token = require_token(args)
    target = (
        getattr(args, "node", None)
        or getattr(args, "namespace", None)
        or getattr(args, "workspace", None)
    )
    api_request(
        args.host,
        token,
        method,
        bind_path(args.bind_type, args.nodegroup, target),
        quiet=args.quiet,
    )


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="NodeGroup API CLI")
    parser.add_argument("--host", default=os.environ.get("KUBESPHERE_HOST", DEFAULT_HOST), help="KubeSphere host URL")
    parser.add_argument("--token", default=os.environ.get("KUBESPHERE_TOKEN"), help="Bearer token")
    parser.add_argument("--quiet", "-q", action="store_true", help="Output compact JSON")

    subparsers = parser.add_subparsers(dest="command", required=True)

    login = subparsers.add_parser("login", help="Get and cache a new token")
    login.add_argument("--username", default=os.environ.get("KUBESPHERE_USERNAME", "admin"))
    login.add_argument("--password", default=os.environ.get("KUBESPHERE_PASSWORD"))

    clear_cache = subparsers.add_parser("clear-cache", help="Clear cached token")

    request = subparsers.add_parser("request", help="Make a raw API request")
    request.add_argument("method", choices=["GET", "POST", "PUT", "PATCH", "DELETE"])
    request.add_argument("path")
    request.add_argument("body", nargs="?", default="")

    nodegroup = subparsers.add_parser("nodegroup", help="NodeGroup operations")
    ng_sub = nodegroup.add_subparsers(dest="nodegroup_action", required=True)
    ng_sub.add_parser("list")
    ng_get = ng_sub.add_parser("get")
    ng_get.add_argument("name")
    ng_create = ng_sub.add_parser("create")
    ng_create.add_argument("--name", required=True)
    ng_create.add_argument("--alias")
    ng_create.add_argument("--description")
    ng_create.add_argument("--manager")
    ng_patch = ng_sub.add_parser("patch")
    ng_patch.add_argument("name")
    ng_patch.add_argument("--alias")
    ng_patch.add_argument("--description")
    ng_patch.add_argument("--manager")
    ng_delete = ng_sub.add_parser("delete")
    ng_delete.add_argument("name")

    bind = subparsers.add_parser("bind", help="Bind node, namespace, or workspace")
    bind_sub = bind.add_subparsers(dest="bind_type", required=True)
    bind_node = bind_sub.add_parser("node")
    bind_node.add_argument("--nodegroup", required=True)
    bind_node.add_argument("--node", required=True)
    bind_ns = bind_sub.add_parser("namespace")
    bind_ns.add_argument("--nodegroup", required=True)
    bind_ns.add_argument("--namespace", required=True)
    bind_ws = bind_sub.add_parser("workspace")
    bind_ws.add_argument("--nodegroup", required=True)
    bind_ws.add_argument("--workspace", required=True)

    unbind = subparsers.add_parser("unbind", help="Unbind node, namespace, or workspace")
    unbind_sub = unbind.add_subparsers(dest="bind_type", required=True)
    unbind_node = unbind_sub.add_parser("node")
    unbind_node.add_argument("--nodegroup", required=True)
    unbind_node.add_argument("--node", required=True)
    unbind_ns = unbind_sub.add_parser("namespace")
    unbind_ns.add_argument("--nodegroup", required=True)
    unbind_ns.add_argument("--namespace", required=True)
    unbind_ws = unbind_sub.add_parser("workspace")
    unbind_ws.add_argument("--nodegroup", required=True)
    unbind_ws.add_argument("--workspace", required=True)

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()

    if args.command == "login":
        password = args.password or getpass.getpass("Password: ")
        token = get_token(args.host, args.username, password)
        print(f"Token loaded: {token[:20]}...")
        return

    if args.command == "clear-cache":
        if os.path.exists(TOKEN_FILE):
            os.remove(TOKEN_FILE)
            print("Token cache cleared.")
        else:
            print("No token cache found.")
        return

    if args.command == "request":
        handle_request(args)
        return

    if args.command == "nodegroup":
        handle_nodegroup(args)
        return

    if args.command == "bind":
        handle_bind(args, "POST")
        return

    if args.command == "unbind":
        handle_bind(args, "DELETE")
        return


if __name__ == "__main__":
    main()
