#!/usr/bin/env python3
"""Import a machine, ecosystem, and IPv4 CSV into Great Firewall.

The input CSV must contain the columns ``machine``, ``ecosystem``, and ``ip``.
The ecosystem mapping converts readable CSV names to existing GFW ecosystem
UUIDs. Only Python's standard library is required.
"""

from __future__ import annotations

import argparse
import csv
import io
import ipaddress
import json
import os
import sys
import uuid
from collections.abc import Iterator
from pathlib import Path
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


DEFAULT_BATCH_SIZE = 100
DEFAULT_TIMEOUT_SECONDS = 30.0
REQUIRED_COLUMNS = {"machine", "ecosystem", "ip"}


def load_source_x(location: str, token: str | None, timeout: float) -> str:
    """Load a Source X CSV export from an HTTP URL or a local file."""
    if location.startswith(("http://", "https://")):
        headers = {"Accept": "text/csv"}
        if token:
            headers["Authorization"] = f"Bearer {token}"
        request = Request(location, headers=headers)
        with urlopen(request, timeout=timeout) as response:
            return response.read().decode("utf-8-sig")

    return Path(location).read_text(encoding="utf-8-sig")


def load_ecosystem_mapping(location: str) -> dict[str, str]:
    """Load and validate the ecosystem-name to GFW-ID mapping."""
    with Path(location).open(encoding="utf-8") as mapping_file:
        document = json.load(mapping_file)
    if not isinstance(document, dict):
        raise ValueError("the ecosystem mapping must be a JSON object")

    mapping: dict[str, str] = {}
    for name, ecosystem_id in document.items():
        if not isinstance(name, str) or not name.strip():
            raise ValueError("ecosystem mapping names must be non-empty strings")
        if not isinstance(ecosystem_id, str):
            raise ValueError(f"ecosystem mapping value for {name!r} must be a UUID string")
        try:
            normalized_id = str(uuid.UUID(ecosystem_id))
        except ValueError as error:
            raise ValueError(f"ecosystem mapping value for {name!r} is not a UUID") from error
        mapping[name.strip()] = normalized_id
    return mapping


def normalize_inventory(csv_text: str, ecosystem_mapping: dict[str, str]) -> list[dict[str, str]]:
    """Convert CSV rows into the GFW machine ingestion shape."""
    try:
        dialect = csv.Sniffer().sniff(csv_text[:4096], delimiters=",;\t")
    except csv.Error as error:
        raise ValueError("CSV delimiter could not be detected") from error
    reader = csv.DictReader(io.StringIO(csv_text), dialect=dialect)
    columns = set(reader.fieldnames or [])
    missing = REQUIRED_COLUMNS - columns
    if missing:
        raise ValueError(f"CSV is missing required column(s): {', '.join(sorted(missing))}")

    machines: list[dict[str, str]] = []
    for row_number, row in enumerate(reader, start=2):
        hostname = required_cell(row, "machine", row_number)
        ecosystem = required_cell(row, "ecosystem", row_number)
        ipv4 = required_cell(row, "ip", row_number)
        ecosystem_id = ecosystem_mapping.get(ecosystem)
        if ecosystem_id is None:
            raise ValueError(
                f"CSV row {row_number}: ecosystem {ecosystem!r} is absent from the mapping"
            )
        try:
            address = ipaddress.ip_address(ipv4)
        except ValueError as error:
            raise ValueError(f"CSV row {row_number}: invalid IP address {ipv4!r}") from error
        if address.version != 4:
            raise ValueError(f"CSV row {row_number}: {ipv4!r} is not an IPv4 address")
        machines.append(
            {"hostname": hostname, "ecosystemID": ecosystem_id, "ipv4": str(address)}
        )
    return machines


def required_cell(row: dict[str, Any], column: str, row_number: int) -> str:
    """Return one required, non-empty CSV cell."""
    value = row.get(column)
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"CSV row {row_number}: column {column!r} must not be empty")
    return value.strip()


def batches(items: list[dict[str, str]], size: int) -> Iterator[list[dict[str, str]]]:
    """Yield bounded ingestion batches."""
    for offset in range(0, len(items), size):
        yield items[offset : offset + size]


def ingest_batch(
    base_url: str,
    api_key: str,
    machines: list[dict[str, str]],
    timeout: float,
) -> int:
    """Send one managed-machine batch and return the reported upsert count."""
    endpoint = f"{base_url.rstrip('/')}/api/v1/ingestion/machines:upsert"
    body = json.dumps({"machines": machines}).encode()
    request = Request(
        endpoint,
        data=body,
        method="POST",
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
    )
    with urlopen(request, timeout=timeout) as response:
        result = json.load(response)
    upserted = result.get("upserted") if isinstance(result, dict) else None
    if not isinstance(upserted, int):
        raise ValueError("GFW returned an invalid ingestion response")
    return upserted


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments and environment-backed defaults."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--source",
        default=os.getenv("SOURCE_X_URL", "scripts/source_x.example.csv"),
        help="Source X CSV URL or file (default: SOURCE_X_URL or example file)",
    )
    parser.add_argument(
        "--ecosystem-map",
        default=os.getenv("GFW_ECOSYSTEM_MAP", "scripts/ecosystem_map.example.json"),
        help="ecosystem-name to GFW-ID JSON file (default: GFW_ECOSYSTEM_MAP)",
    )
    parser.add_argument(
        "--source-token",
        default=os.getenv("SOURCE_X_TOKEN"),
        help="Source X bearer token (default: SOURCE_X_TOKEN)",
    )
    parser.add_argument(
        "--gfw-url",
        default=os.getenv("GFW_URL", "http://localhost:8080"),
        help="Great Firewall base URL (default: GFW_URL or http://localhost:8080)",
    )
    parser.add_argument(
        "--gfw-key",
        default=os.getenv("GFW_INGESTION_KEY"),
        help="Great Firewall ingestion key (default: GFW_INGESTION_KEY)",
    )
    parser.add_argument("--batch-size", type=int, default=DEFAULT_BATCH_SIZE)
    parser.add_argument("--timeout", type=float, default=DEFAULT_TIMEOUT_SECONDS)
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="validate and print the GFW payload without sending it",
    )
    args = parser.parse_args()
    if args.batch_size < 1:
        parser.error("--batch-size must be greater than zero")
    if args.timeout <= 0:
        parser.error("--timeout must be greater than zero")
    if not args.dry_run and not args.gfw_key:
        parser.error("--gfw-key or GFW_INGESTION_KEY is required unless --dry-run is used")
    return args


def main() -> int:
    """Load, normalize, and ingest the Source X CSV inventory."""
    args = parse_args()
    try:
        csv_text = load_source_x(args.source, args.source_token, args.timeout)
        ecosystem_mapping = load_ecosystem_mapping(args.ecosystem_map)
        machines = normalize_inventory(csv_text, ecosystem_mapping)
        if args.dry_run:
            print(json.dumps({"machines": machines}, indent=2))
            print(f"Validated {len(machines)} CSV row(s).", file=sys.stderr)
            return 0

        total = 0
        for batch in batches(machines, args.batch_size):
            total += ingest_batch(args.gfw_url, args.gfw_key, batch, args.timeout)
        print(f"Upserted {total} machine(s) into Great Firewall.")
        return 0
    except HTTPError as error:
        detail = error.read().decode(errors="replace")
        print(f"HTTP {error.code}: {detail}", file=sys.stderr)
    except (OSError, URLError, ValueError, json.JSONDecodeError) as error:
        print(f"Ingestion failed: {error}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
