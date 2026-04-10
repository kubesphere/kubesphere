#!/usr/bin/env python3
"""Generate canonical FrontendIntegration YAML from a simplified JSON authoring spec."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any


API_VERSION = "frontend-forge.kubesphere.io/v1alpha1"
RESOURCE_KIND = "FrontendIntegration"
DEFAULT_MENU_ICON = "GridDuotone"
VALID_PLACEMENTS = {"cluster", "workspace", "global"}
VALID_PAGE_TYPES = {"crdTable", "iframe"}
VALID_RENDER_TYPES = {"text", "time", "link"}
YAML_RESERVED_WORDS = {
    "",
    "~",
    "null",
    "true",
    "false",
    "yes",
    "no",
    "on",
    "off",
}
PLAIN_VALUE_RE = re.compile(r"[A-Za-z0-9_./:-]+")
PLAIN_KEY_RE = re.compile(r"[A-Za-z0-9_-]+")
NUMERIC_LIKE_RE = re.compile(
    r"""
    [-+]?(?:
        0
        | [1-9][0-9_]*
        | 0[0-7_]+
        | 0x[0-9a-f]+
        | (?:[0-9][0-9_]*)?\.[0-9_]+(?:e[-+]?[0-9]+)?
        | [0-9][0-9_]*\.(?:e[-+]?[0-9]+)?
        | [0-9][0-9_]*e[-+]?[0-9]+
        | \.(?:inf|nan)
    )
    """,
    re.IGNORECASE | re.VERBOSE,
)


class ValidationError(Exception):
    """Raised when the authoring spec is invalid."""


def is_non_empty_string(value: Any) -> bool:
    return isinstance(value, str) and bool(value.strip())


def require_dict(value: Any, path: str) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise ValidationError(f"{path} must be an object.")
    return value


def require_list(value: Any, path: str) -> list[Any]:
    if not isinstance(value, list):
        raise ValidationError(f"{path} must be an array.")
    return value


def require_string(value: Any, path: str) -> str:
    if not is_non_empty_string(value):
        raise ValidationError(f"{path} must be a non-empty string.")
    return value.strip()


def normalize_bool(value: Any, path: str, default: bool | None = None) -> bool:
    if value is None:
        if default is None:
            raise ValidationError(f"{path} must be a boolean.")
        return default
    if not isinstance(value, bool):
        raise ValidationError(f"{path} must be a boolean.")
    return value


def normalize_scope(raw_scope: Any) -> str:
    scope = require_string(raw_scope, "pages[].crdTable.scope").lower()
    if scope == "cluster":
        return "Cluster"
    if scope in {"namespace", "namespaced"}:
        return "Namespaced"
    raise ValidationError(
        "pages[].crdTable.scope must be one of Cluster, Namespace, or Namespaced."
    )


def is_namespaced_scope(scope: str) -> bool:
    return scope == "Namespaced"


def create_name_column() -> dict[str, Any]:
    return {
        "enableSorting": True,
        "key": "name",
        "render": {
            "path": "metadata.name",
            "type": "text",
        },
        "title": "NAME",
    }


def create_namespace_column() -> dict[str, Any]:
    return {
        "enableHiding": True,
        "key": "namespace",
        "render": {
            "path": "metadata.namespace",
            "type": "text",
        },
        "title": "PROJECT",
    }


def create_update_time_column() -> dict[str, Any]:
    return {
        "enableHiding": True,
        "enableSorting": True,
        "key": "updateTime",
        "render": {
            "format": "local-datetime",
            "path": "metadata.creationTimestamp",
            "type": "time",
        },
        "title": "CREATION_TIME",
    }


def default_columns(scope: str) -> list[dict[str, Any]]:
    columns = [create_name_column()]
    if is_namespaced_scope(scope):
        columns.append(create_namespace_column())
    columns.append(create_update_time_column())
    return columns


def validate_column(column: Any, index: int) -> dict[str, Any]:
    value = require_dict(column, f"pages[].crdTable.columns[{index}]")
    key = require_string(value.get("key"), f"pages[].crdTable.columns[{index}].key")
    title = require_string(value.get("title"), f"pages[].crdTable.columns[{index}].title")
    render = require_dict(value.get("render"), f"pages[].crdTable.columns[{index}].render")
    render_type = require_string(
        render.get("type"), f"pages[].crdTable.columns[{index}].render.type"
    )
    if render_type not in VALID_RENDER_TYPES:
        raise ValidationError(
            f"pages[].crdTable.columns[{index}].render.type must be one of "
            f"{', '.join(sorted(VALID_RENDER_TYPES))}."
        )
    path = require_string(render.get("path"), f"pages[].crdTable.columns[{index}].render.path")

    normalized_render: dict[str, Any] = {
        "type": render_type,
        "path": path,
    }
    if "format" in render and render["format"] is not None:
        normalized_render["format"] = require_string(
            render["format"], f"pages[].crdTable.columns[{index}].render.format"
        )
    if "pattern" in render and render["pattern"] is not None:
        normalized_render["pattern"] = require_string(
            render["pattern"], f"pages[].crdTable.columns[{index}].render.pattern"
        )
    if "link" in render and render["link"] is not None:
        normalized_render["link"] = require_string(
            render["link"], f"pages[].crdTable.columns[{index}].render.link"
        )
    if "payload" in render and render["payload"] is not None:
        normalized_render["payload"] = require_dict(
            render["payload"], f"pages[].crdTable.columns[{index}].render.payload"
        )

    normalized_column: dict[str, Any] = {
        "key": key,
        "title": title,
        "render": normalized_render,
    }
    if "enableSorting" in value:
        normalized_column["enableSorting"] = normalize_bool(
            value["enableSorting"], f"pages[].crdTable.columns[{index}].enableSorting"
        )
    if "enableHiding" in value:
        normalized_column["enableHiding"] = normalize_bool(
            value["enableHiding"], f"pages[].crdTable.columns[{index}].enableHiding"
        )
    return normalized_column


def normalize_columns(columns: list[dict[str, Any]], scope: str) -> list[dict[str, Any]]:
    next_columns = [column for column in columns if column.get("key") != "namespace"]
    if not is_namespaced_scope(scope):
        return next_columns

    namespace_column = create_namespace_column()
    update_time_index = next(
        (index for index, column in enumerate(next_columns) if column.get("key") == "updateTime"),
        -1,
    )
    if update_time_index >= 0:
        return (
            next_columns[:update_time_index]
            + [namespace_column]
            + next_columns[update_time_index:]
        )

    name_index = next(
        (index for index, column in enumerate(next_columns) if column.get("key") == "name"),
        -1,
    )
    if name_index >= 0:
        return next_columns[: name_index + 1] + [namespace_column] + next_columns[name_index + 1 :]

    return next_columns + [namespace_column]


def sanitize_annotations(raw: Any) -> dict[str, str] | None:
    if raw is None:
        return None
    annotations = require_dict(raw, "metadata.annotations")
    normalized: dict[str, str] = {}
    for key, value in annotations.items():
        if not is_non_empty_string(key):
            raise ValidationError("metadata.annotations keys must be non-empty strings.")
        normalized[key.strip()] = require_string(value, f"metadata.annotations[{key!r}]")
    return normalized or None


def normalize_placements(raw: Any) -> list[str]:
    placements = require_list(raw, "menu.placements")
    if not placements:
        raise ValidationError("menu.placements must contain at least one placement.")
    normalized: list[str] = []
    seen: set[str] = set()
    for index, placement in enumerate(placements):
        value = require_string(placement, f"menu.placements[{index}]").lower()
        if value not in VALID_PLACEMENTS:
            raise ValidationError(
                f"menu.placements[{index}] must be one of {', '.join(sorted(VALID_PLACEMENTS))}."
            )
        if value in seen:
            raise ValidationError(f"menu.placements contains a duplicate placement: {value}.")
        seen.add(value)
        normalized.append(value)
    return normalized


def normalize_metadata(payload: dict[str, Any]) -> dict[str, Any]:
    metadata = require_dict(payload.get("metadata"), "metadata")
    normalized: dict[str, Any] = {
        "name": require_string(metadata.get("name"), "metadata.name"),
    }
    annotations = sanitize_annotations(metadata.get("annotations"))
    if annotations:
        normalized["annotations"] = annotations
    return normalized


def normalize_spec_options(payload: dict[str, Any]) -> dict[str, Any]:
    raw_spec = payload.get("spec")
    if raw_spec is None:
        return {"enabled": True}
    spec = require_dict(raw_spec, "spec")
    normalized: dict[str, Any] = {
        "enabled": normalize_bool(spec.get("enabled"), "spec.enabled", default=True),
    }
    if "displayName" in spec and spec["displayName"] is not None:
        normalized["displayName"] = require_string(spec["displayName"], "spec.displayName")
    if "builder" in spec and spec["builder"] is not None:
        builder = require_dict(spec["builder"], "spec.builder")
        normalized_builder: dict[str, Any] = {}
        if "engineVersion" in builder and builder["engineVersion"] is not None:
            normalized_builder["engineVersion"] = require_string(
                builder["engineVersion"], "spec.builder.engineVersion"
            )
        if normalized_builder:
            normalized["builder"] = normalized_builder
    if "locales" in spec and spec["locales"] is not None:
        locales = require_dict(spec["locales"], "spec.locales")
        normalized_locales: dict[str, dict[str, str]] = {}
        for locale_key, locale_map in locales.items():
            if not is_non_empty_string(locale_key):
                raise ValidationError("spec.locales keys must be non-empty strings.")
            locale_dict = require_dict(locale_map, f"spec.locales[{locale_key!r}]")
            normalized_locale_entries: dict[str, str] = {}
            for message_key, message_value in locale_dict.items():
                if not is_non_empty_string(message_key):
                    raise ValidationError(
                        f"spec.locales[{locale_key!r}] keys must be non-empty strings."
                    )
                normalized_locale_entries[message_key.strip()] = require_string(
                    message_value, f"spec.locales[{locale_key!r}][{message_key!r}]"
                )
            normalized_locales[locale_key.strip()] = normalized_locale_entries
        if normalized_locales:
            normalized["locales"] = normalized_locales
    return normalized


def normalize_menu(payload: dict[str, Any]) -> dict[str, Any]:
    menu = require_dict(payload.get("menu"), "menu")
    raw_icon = menu.get("icon")
    return {
        "displayName": require_string(menu.get("displayName"), "menu.displayName"),
        "icon": require_string(raw_icon, "menu.icon") if raw_icon is not None else DEFAULT_MENU_ICON,
        "placements": normalize_placements(menu.get("placements")),
    }


def normalize_crd_page(page: dict[str, Any], force_namespaced: bool) -> tuple[dict[str, Any], dict[str, str]]:
    crd = require_dict(page.get("crdTable"), "pages[].crdTable")
    group = require_string(crd.get("group"), "pages[].crdTable.group")
    version = require_string(crd.get("version"), "pages[].crdTable.version")
    names = require_dict(crd.get("names"), "pages[].crdTable.names")
    plural = require_string(names.get("plural"), "pages[].crdTable.names.plural")
    scope = normalize_scope(crd.get("scope"))
    if force_namespaced:
        scope = "Namespaced"

    key = page.get("key")
    normalized_key = require_string(key, "pages[].key") if key is not None else plural
    display_name = require_string(page.get("displayName"), "pages[].displayName")

    normalized_crd: dict[str, Any] = {
        "columns": [],
        "group": group,
        "names": {
            "plural": plural,
        },
        "scope": scope,
        "version": version,
    }
    if "kind" in names and names["kind"] is not None:
        normalized_crd["names"]["kind"] = require_string(
            names["kind"], "pages[].crdTable.names.kind"
        )
    if "authKey" in crd and crd["authKey"] is not None:
        normalized_crd["authKey"] = require_string(crd["authKey"], "pages[].crdTable.authKey")

    if "columns" in crd and crd["columns"] is not None:
        raw_columns = require_list(crd["columns"], "pages[].crdTable.columns")
        validated_columns = [validate_column(column, index) for index, column in enumerate(raw_columns)]
        normalized_crd["columns"] = normalize_columns(validated_columns, scope)
    else:
        normalized_crd["columns"] = default_columns(scope)

    page_output = {
        "key": normalized_key,
        "type": "crdTable",
        "crdTable": normalized_crd,
    }
    child = {
        "key": normalized_key,
        "displayName": display_name,
    }
    return page_output, child


def normalize_iframe_page(page: dict[str, Any]) -> tuple[dict[str, Any], dict[str, str]]:
    iframe = require_dict(page.get("iframe"), "pages[].iframe")
    key = require_string(page.get("key"), "pages[].key")
    display_name = require_string(page.get("displayName"), "pages[].displayName")
    src = require_string(iframe.get("src"), "pages[].iframe.src")
    page_output = {
        "key": key,
        "type": "iframe",
        "iframe": {
            "src": src,
        },
    }
    child = {
        "key": key,
        "displayName": display_name,
    }
    return page_output, child


def normalize_pages(payload: dict[str, Any], placements: list[str]) -> tuple[list[dict[str, Any]], list[dict[str, str]]]:
    raw_pages = require_list(payload.get("pages"), "pages")
    if not raw_pages:
        raise ValidationError("pages must contain at least one page.")

    force_namespaced = "workspace" in placements
    pages: list[dict[str, Any]] = []
    children: list[dict[str, str]] = []
    seen_keys: set[str] = set()

    for index, item in enumerate(raw_pages):
        page = require_dict(item, f"pages[{index}]")
        page_type = require_string(page.get("type"), f"pages[{index}].type")
        if page_type not in VALID_PAGE_TYPES:
            raise ValidationError(
                f"pages[{index}].type must be one of {', '.join(sorted(VALID_PAGE_TYPES))}."
            )
        if page_type == "crdTable":
            normalized_page, child = normalize_crd_page(page, force_namespaced)
        else:
            normalized_page, child = normalize_iframe_page(page)

        page_key = normalized_page["key"]
        if page_key in seen_keys:
            raise ValidationError(f"pages contains a duplicate key: {page_key}.")
        seen_keys.add(page_key)
        pages.append(normalized_page)
        children.append(child)

    return pages, children


def build_canonical_resource(payload: dict[str, Any]) -> dict[str, Any]:
    metadata = normalize_metadata(payload)
    spec_options = normalize_spec_options(payload)
    menu = normalize_menu(payload)
    pages, children = normalize_pages(payload, menu["placements"])

    menus = [
        {
            "key": f"{metadata['name']}-{placement}",
            "displayName": menu["displayName"],
            "icon": menu["icon"],
            "placement": placement,
            "type": "organization",
            "children": children,
        }
        for placement in menu["placements"]
    ]

    spec: dict[str, Any] = {}
    if "builder" in spec_options:
        spec["builder"] = spec_options["builder"]
    if "displayName" in spec_options:
        spec["displayName"] = spec_options["displayName"]
    spec["enabled"] = spec_options["enabled"]
    spec["menus"] = menus
    spec["pages"] = pages
    if "locales" in spec_options:
        spec["locales"] = spec_options["locales"]

    return {
        "apiVersion": API_VERSION,
        "kind": RESOURCE_KIND,
        "metadata": metadata,
        "spec": spec,
    }


def parse_input(raw_text: str) -> dict[str, Any]:
    try:
        data = json.loads(raw_text)
    except json.JSONDecodeError as error:
        raise ValidationError(f"Input must be valid JSON: {error.msg}.") from error
    return require_dict(data, "root")


def has_control_characters(value: str) -> bool:
    return any(ord(char) < 0x20 for char in value)


def is_yaml_keyword(value: str) -> bool:
    return value.lower() in YAML_RESERVED_WORDS


def looks_like_number(value: str) -> bool:
    return bool(NUMERIC_LIKE_RE.fullmatch(value))


def needs_quotes(value: str, *, pattern: re.Pattern[str]) -> bool:
    if value == "":
        return True
    if value.strip() != value:
        return True
    if has_control_characters(value):
        return True
    if is_yaml_keyword(value):
        return True
    if looks_like_number(value):
        return True
    return not bool(pattern.fullmatch(value))


def format_string(value: str) -> str:
    if not needs_quotes(value, pattern=PLAIN_VALUE_RE):
        return value
    return json.dumps(value, ensure_ascii=False)


def format_key(key: str) -> str:
    if not needs_quotes(key, pattern=PLAIN_KEY_RE):
        return key
    return format_string(key)


def dump_scalar(value: Any) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if value is None:
        return "null"
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        return str(value)
    if isinstance(value, str):
        return format_string(value)
    raise TypeError(f"Unsupported scalar value: {value!r}")


def dump_yaml(value: Any, indent: int = 0) -> list[str]:
    prefix = " " * indent
    if isinstance(value, dict):
        lines: list[str] = []
        for key, item in value.items():
            formatted_key = format_key(str(key))
            if isinstance(item, (dict, list)):
                lines.append(f"{prefix}{formatted_key}:")
                lines.extend(dump_yaml(item, indent + 2))
            else:
                lines.append(f"{prefix}{formatted_key}: {dump_scalar(item)}")
        return lines
    if isinstance(value, list):
        lines = []
        for item in value:
            if isinstance(item, (dict, list)):
                lines.append(f"{prefix}-")
                lines.extend(dump_yaml(item, indent + 2))
            else:
                lines.append(f"{prefix}- {dump_scalar(item)}")
        return lines
    return [f"{prefix}{dump_scalar(value)}"]


def read_input(path: str) -> str:
    if path == "-":
        return sys.stdin.read()
    return Path(path).read_text(encoding="utf-8")


def write_output(path: str | None, text: str) -> None:
    if path:
        Path(path).write_text(text, encoding="utf-8")
        return
    sys.stdout.write(text)
    if not text.endswith("\n"):
        sys.stdout.write("\n")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Generate canonical FrontendIntegration YAML from a simplified JSON authoring spec."
    )
    parser.add_argument(
        "--input",
        default="-",
        help="Path to a JSON authoring spec. Use - or omit to read from stdin.",
    )
    parser.add_argument(
        "--output",
        help="Optional path for the generated YAML. Defaults to stdout.",
    )
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    try:
        raw_input = read_input(args.input)
        payload = parse_input(raw_input)
        resource = build_canonical_resource(payload)
        yaml_text = "\n".join(dump_yaml(resource)) + "\n"
        write_output(args.output, yaml_text)
    except (OSError, ValidationError, TypeError) as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
