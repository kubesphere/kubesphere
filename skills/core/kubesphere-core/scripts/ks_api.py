#!/usr/bin/env python3
"""
KubeSphere API CLI - Simple wrapper for KubeSphere IAM API calls

Usage:
    # Get token first, then call API
    python ks_api.py --login --username admin --password <password>
    python ks_api.py GET /users
    python ks_api.py POST /users '{"username": "test", "email": "test@example.com"}'

Environment:
    KUBESPHERE_HOST: KubeSphere host URL (default: http://ks-apiserver.kubesphere-system)
    KUBESPHERE_USERNAME: KubeSphere username (default: admin)
    KUBESPHERE_PASSWORD: KubeSphere password
    KUBESPHERE_TOKEN: Cached token (auto-refreshed if expired)
"""

import argparse
import getpass
import json
import os
import sys
import time

import requests


TOKEN_FILE = os.path.expanduser("~/.kubesphere_token")


def save_token(token: str, expires_at: int = None):
    """Save token to file"""
    data = {"token": token, "expires_at": expires_at, "saved_at": int(time.time())}
    with open(TOKEN_FILE, "w") as f:
        json.dump(data, f)


def load_token() -> str:
    """Load token from file if still valid"""
    if not os.path.exists(TOKEN_FILE):
        return None

    try:
        with open(TOKEN_FILE, "r") as f:
            data = json.load(f)

        # Check if token expired
        if data.get("expires_at") and data["expires_at"] < int(time.time()):
            return None

        return data.get("token")
    except (json.JSONDecodeError, IOError):
        return None


def get_token(host: str, username: str, password: str) -> str:
    """Get OAuth token from KubeSphere (always fetches a new token)"""
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

        # Save token
        expires_in = result.get("expires_in", 3600 * 5)  # default 5 hours
        expires_at = int(time.time()) + expires_in
        save_token(access_token, expires_at)

        return access_token

    except requests.RequestException as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="KubeSphere API CLI - Make API calls to KubeSphere"
    )

    # Login options
    parser.add_argument(
        "--login",
        action="store_true",
        help="Get new token (ignore cached token)",
    )
    parser.add_argument(
        "--username",
        default=os.environ.get("KUBESPHERE_USERNAME", "admin"),
        help="KubeSphere username",
    )
    parser.add_argument(
        "--password",
        default=os.environ.get("KUBESPHERE_PASSWORD"),
        help="KubeSphere password (will prompt securely if not provided with --login)",
    )

    # API call options
    parser.add_argument(
        "method",
        choices=["GET", "POST", "PUT", "PATCH", "DELETE"],
        nargs="?",
        help="HTTP method",
    )
    parser.add_argument("uri", nargs="?", help="API URI path (e.g., /users)")
    parser.add_argument(
        "body",
        nargs="?",
        default="{}",
        help="Request body as JSON string (optional)",
    )

    parser.add_argument(
        "--host",
        default=os.environ.get("KUBESPHERE_HOST", "http://ks-apiserver.kubesphere-system"),
        help="KubeSphere host URL",
    )
    parser.add_argument(
        "--token",
        default=os.environ.get("KUBESPHERE_TOKEN"),
        help="Admin token (takes precedence over cached token)",
    )
    parser.add_argument(
        "--quiet", "-q", action="store_true", help="Only output JSON result"
    )
    parser.add_argument(
        "--clear-cache", action="store_true", help="Clear cached token"
    )

    args = parser.parse_args()

    # Handle --clear-cache
    if args.clear_cache:
        if os.path.exists(TOKEN_FILE):
            os.remove(TOKEN_FILE)
            print("Token cache cleared.")
        else:
            print("No token cache found.")
        sys.exit(0)

    # If --login, get password securely (from env, args, or prompt)
    if args.login:
        if not args.password:
            # Prompt for password securely without echoing
            args.password = getpass.getpass("Password: ")

    # Get token
    token = args.token
    if args.login:
        token = get_token(args.host, args.username, args.password)
    elif not token:
        # Try cached token
        token = load_token()
        if not token:
            print("Error: No token available. Use --login to get a new token.", file=sys.stderr)
            print("Usage:", file=sys.stderr)
            print("  python ks_api.py --login --username admin --password <password>", file=sys.stderr)
            print("  python ks_api.py GET /users", file=sys.stderr)
            sys.exit(1)

    # If no method/uri, just print token info
    if not args.method or not args.uri:
        if token:
            print(f"Token loaded: {token[:20]}...")
        else:
            print("No token available.")
        sys.exit(0)

    # Build full URL
    base_url = args.host.rstrip("/")
    uri = args.uri.lstrip("/")
    url = f"{base_url}/{uri}"

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json",
    }

    # Parse body
    body = None
    if args.body and args.method in ["POST", "PUT", "PATCH"]:
        try:
            body = json.loads(args.body)
        except json.JSONDecodeError as e:
            print(f"Error: Invalid JSON body: {e}", file=sys.stderr)
            sys.exit(1)

    try:
        response = requests.request(
            method=args.method,
            url=url,
            headers=headers,
            json=body if body else None,
            timeout=30,
        )

        # Try to parse response as JSON
        try:
            result = response.json()
        except json.JSONDecodeError:
            result = {"raw": response.text}

        # Output
        if args.quiet:
            print(json.dumps(result, indent=2))
        else:
            if response.status_code >= 400:
                print(f"Error: HTTP {response.status_code}", file=sys.stderr)
                print(json.dumps(result, indent=2))
                sys.exit(1)
            else:
                print(json.dumps(result, indent=2))

    except requests.RequestException as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
