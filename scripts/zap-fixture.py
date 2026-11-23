#!/usr/bin/env python3
"""Reproduce the Juice Shop fixture or derive its v1 observation partition offline."""

import argparse
import hashlib
import json
from pathlib import Path
import re
from urllib.parse import unquote_plus, urlsplit, urlunsplit
from uuid import UUID


ORIGIN = "http://juice-shop:3000"


def redact_fields(record, names):
    for name in names:
        if record.get(name):
            record[name] = "[redacted]"


def sanitize(report):
    for site in report["site"]:
        if site["@name"] != ORIGIN:
            raise ValueError("Expected the recorded Juice Shop origin")
        for alert in site["alerts"]:
            redact_fields(alert, ("otherinfo",))
            for instance in alert["instances"]:
                url = urlsplit(instance["uri"])
                require_origin(url)
                query = "&".join(
                    f"{value.partition('=')[0]}=redacted" if "=" in value else value
                    for value in url.query.split("&")
                )
                instance["uri"] = urlunsplit(url._replace(query=query))
                instance.pop("nodeName", None)
                redact_fields(instance, ("attack", "evidence", "otherinfo"))
    return report


def require_origin(url):
    if f"{url.scheme}://{url.netloc}" != ORIGIN or url.fragment:
        raise ValueError("Unexpected observation location")


def decode_unreserved(match):
    char = chr(int(match.group()[1:], 16))
    return char if re.fullmatch(r"[A-Za-z0-9._~-]", char) else match.group()


def groups(report, target_id):
    result = {}
    for s, site in enumerate(report["site"]):
        for a, alert in enumerate(site["alerts"]):
            for i, instance in enumerate(alert["instances"]):
                url = urlsplit(instance["uri"])
                require_origin(url)
                names = sorted({
                    unquote_plus(part.partition("=")[0], errors="strict")
                    for part in url.query.split("&") if part
                })
                path = re.sub(r"%[0-9A-Fa-f]{2}", decode_unreserved, url.path or "/")
                identity = [
                    target_id, "zap", str(alert["pluginid"]), instance["method"].upper(),
                    ORIGIN, path, names, instance.get("param", ""),
                ]
                encoded = json.dumps(identity, ensure_ascii=False, separators=(",", ":")).encode()
                fingerprint = hashlib.sha256(encoded).hexdigest()
                result.setdefault(fingerprint, []).append(f"/site/{s}/alerts/{a}/instances/{i}")
    return dict(sorted(result.items()))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("operation", choices=("sanitize", "groups"))
    parser.add_argument("report", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--target", type=UUID)
    args = parser.parse_args()
    report = json.loads(args.report.read_bytes())
    if args.operation == "groups":
        if args.target is None:
            parser.error("groups requires --target")
        result = groups(report, str(args.target))
    else:
        result = sanitize(report)
    with args.output.open("x", encoding="utf-8", newline="\n") as handle:
        json.dump(result, handle, indent=2, ensure_ascii=False)
        handle.write("\n")


if __name__ == "__main__":
    main()
