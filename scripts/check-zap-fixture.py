#!/usr/bin/env python3
"""Check a local ZAP fixture candidate's digests, provenance and report structure."""

import argparse
from datetime import datetime
import hashlib
import json
from pathlib import Path
import re


REPORT_LIMIT = 20 << 20
PROVENANCE_LIMIT = 64 << 10
OPTIONAL_TEXT = {"zap_version", "juice_shop_version", "juice_shop_image"}
PROVENANCE_FIELDS = OPTIONAL_TEXT | {
    "schema_version", "addon_versions", "report_template", "authorization_ref",
    "scan_time", "plan_sha256", "raw_sha256", "sanitized_sha256", "sanitization_steps",
}
RFC3339 = re.compile(
    r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]+)?"
    r"(?:Z|[+-](?:[01][0-9]|2[0-3]):[0-5][0-9])"
)


class FixtureArguments(argparse.ArgumentParser):
    def error(self, _message):
        self.exit(2, "Expected original, derivative and provenance file paths. Use --help.\n")


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("Duplicate JSON key")
        result[key] = value
    return result


def invalid_constant(_value):
    raise ValueError("Invalid JSON constant")


def read_json(path, label, limit):
    try:
        with path.open("rb") as handle:
            raw = handle.read(limit + 1)
    except OSError:
        raise ValueError(f"Cannot read {label}") from None
    if len(raw) > limit:
        raise ValueError(f"{label} exceeds its byte limit")
    try:
        value = json.loads(raw.decode("utf-8"), object_pairs_hook=unique_object, parse_constant=invalid_constant)
    except (ValueError, RecursionError):
        raise ValueError(f"{label} must be UTF-8 JSON with no duplicate keys") from None
    return hashlib.sha256(raw).hexdigest(), value


def nonempty_text(value):
    return isinstance(value, str) and bool(value.strip()) and "\x00" not in value


def validate_provenance(value):
    if not isinstance(value, dict) or value.keys() - PROVENANCE_FIELDS:
        raise ValueError("Provenance must be an object with recognized fields")
    if value.get("schema_version", "report-provenance/v2") != "report-provenance/v2":
        raise ValueError("Invalid provenance schema_version")
    if value.get("report_template") != "traditional-json":
        raise ValueError("Provenance report_template must be traditional-json")
    if not nonempty_text(value.get("authorization_ref")):
        raise ValueError("Provenance needs a lab authorization reference")
    scan_time = value.get("scan_time")
    if not isinstance(scan_time, str) or not RFC3339.fullmatch(scan_time):
        raise ValueError("Provenance scan_time must be RFC3339")
    try:
        datetime.fromisoformat(scan_time.replace("Z", "+00:00"))
    except ValueError:
        raise ValueError("Provenance scan_time must be a valid RFC3339 timestamp") from None

    for field in sorted(OPTIONAL_TEXT):
        if value.get(field) is not None and not nonempty_text(value[field]):
            raise ValueError(f"Provenance {field} must be nonempty text or null")
    addons = value.get("addon_versions")
    if addons is not None and (not isinstance(addons, dict) or not all(
        nonempty_text(name) and nonempty_text(version) for name, version in addons.items()
    )):
        raise ValueError("Provenance addon_versions must contain nonempty text versions or be null")
    for field in ("raw_sha256", "sanitized_sha256", "plan_sha256"):
        digest = value.get(field)
        if field == "plan_sha256" and digest is None:
            continue
        if not isinstance(digest, str) or not re.fullmatch(r"[0-9a-f]{64}", digest):
            raise ValueError(f"Provenance {field} must be a lowercase SHA-256 digest")
    steps = value.get("sanitization_steps")
    if not isinstance(steps, list) or not steps or not all(map(nonempty_text, steps)):
        raise ValueError("Provenance needs ordered, nonempty sanitization_steps")


def read_report_layout(path, label, provenance, digest_field):
    digest, report = read_json(path, label, REPORT_LIMIT)
    if digest != provenance[digest_field]:
        raise ValueError(f"Provenance {digest_field} differs from the {label} bytes")
    if not isinstance(report, dict) or not nonempty_text(report.get("@version")):
        raise ValueError(f"{label} needs a Traditional JSON @version")
    if report["@version"].strip().casefold().startswith("synthetic"):
        raise ValueError(f"{label} is labeled synthetic")
    sites = report.get("site")
    if not isinstance(sites, list) or not 1 <= len(sites) <= 10:
        raise ValueError(f"{label} needs between 1 and 10 sites")
    layout = []
    alert_count = observation_count = 0
    for site_index, site in enumerate(sites):
        pointer = f"/site/{site_index}"
        if not isinstance(site, dict) or not nonempty_text(site.get("@name")):
            raise ValueError(f"{label} needs a site name at {pointer}/@name")
        alerts = site.get("alerts")
        if not isinstance(alerts, list):
            raise ValueError(f"{label} needs an alerts array at {pointer}/alerts")
        alert_count += len(alerts)
        if alert_count > 5000:
            raise ValueError(f"{label} exceeds 5,000 alerts")
        counts = []
        for alert_index, alert in enumerate(alerts):
            alert_pointer = f"{pointer}/alerts/{alert_index}"
            if not isinstance(alert, dict):
                raise ValueError(f"{label} needs an alert object at {alert_pointer}")
            instances = alert.get("instances")
            if instances is None:
                counts.append(0)
                continue
            if not isinstance(instances, list):
                raise ValueError(f"{label} needs an instances array at {alert_pointer}/instances")
            observation_count += len(instances)
            if observation_count > 20000:
                raise ValueError(f"{label} exceeds 20,000 instances")
            for instance_index, instance in enumerate(instances):
                instance_pointer = f"{alert_pointer}/instances/{instance_index}"
                if not isinstance(instance, dict):
                    raise ValueError(f"{label} needs an instance object at {instance_pointer}")
                for field in ("uri", "method"):
                    if not nonempty_text(instance.get(field)):
                        raise ValueError(f"{label} needs {field} at {instance_pointer}/{field}")
            counts.append(len(instances))
        layout.append(counts)
    return layout


def main():
    parser = FixtureArguments(description=__doc__)
    parser.add_argument("original", type=Path)
    parser.add_argument("derivative", type=Path)
    parser.add_argument("provenance", type=Path)
    args = parser.parse_args()

    try:
        _, provenance = read_json(args.provenance, "provenance", PROVENANCE_LIMIT)
        validate_provenance(provenance)
        original_layout = read_report_layout(args.original, "original report", provenance, "raw_sha256")
        derivative_layout = read_report_layout(
            args.derivative, "sanitized derivative", provenance, "sanitized_sha256"
        )
        if original_layout != derivative_layout:
            raise ValueError("Sanitization changed site, alert or instance counts by source position")
        if provenance["raw_sha256"] == provenance["sanitized_sha256"]:
            raise ValueError("The sanitized derivative must differ from the original bytes")
    except ValueError as error:
        parser.exit(1, f"{error}\n")

    counts = [count for site in original_layout for count in site]
    print(json.dumps({
        "site_count": len(original_layout),
        "alert_count": len(counts),
        "observation_count": sum(counts),
        "unlocated_alert_count": counts.count(0),
    }, sort_keys=True))


if __name__ == "__main__":
    main()
