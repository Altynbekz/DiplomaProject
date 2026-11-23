"""Exercise the preflight CLI with synthetic test data only."""

import hashlib
import json
from pathlib import Path
import subprocess
import sys
from tempfile import TemporaryDirectory
import unittest


SCRIPT = Path(__file__).with_name("check-zap-fixture.py")
SYNTHETIC_FIXTURE = SCRIPT.parent.parent / "backend/testdata/zap/synthetic-traditional.json"


class FixturePreflightTests(unittest.TestCase):
    def setUp(self):
        folder = TemporaryDirectory()
        self.addCleanup(folder.cleanup)
        self.original, self.derivative, self.provenance = (
            Path(folder.name) / name for name in ("original.json", "derivative.json", "provenance.json")
        )
        location = {"uri": "http://juice-shop:3000/test-only", "method": "GET"}
        self.report = {
            "@version": "UNIT-TEST-ONLY",
            "site": [{"@name": "http://juice-shop:3000", "alerts": [
                {"pluginid": "1", "instances": [{**location, "evidence": "private-unit-test-value"}, location]},
                {},
                {"instances": None},
            ]}],
        }
        self.original.write_text(json.dumps(self.report), encoding="utf-8")
        self.report["site"][0]["alerts"][0]["instances"][0]["evidence"] = ""
        self.derivative.write_text(json.dumps(self.report), encoding="utf-8")
        self.meta = {
            "report_template": "traditional-json",
            "authorization_ref": "unit-test-placeholder",
            "scan_time": "2020-01-01T12:00:00Z",
            "raw_sha256": self.digest(self.original),
            "sanitized_sha256": self.digest(self.derivative),
            "sanitization_steps": ["Unit test removes the evidence value"],
        }

    @staticmethod
    def digest(path):
        return hashlib.sha256(path.read_bytes()).hexdigest()

    def invoke(self, provenance_text=None, *, extra_arguments=()):
        self.provenance.write_text(provenance_text or json.dumps(self.meta), encoding="utf-8")
        return subprocess.run(
            [sys.executable, SCRIPT, self.original, self.derivative, self.provenance, *extra_arguments],
            capture_output=True, text=True, check=False,
        )

    def assert_rejected(self, result, detail):
        self.assertEqual(result.returncode, 1)
        self.assertEqual(result.stdout, "")
        self.assertIn(detail, result.stderr)
        self.assertNotIn("private-unit-test-value", result.stderr)

    def test_counts_and_keeps_original_bytes(self):
        result = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(result.stdout), {
            "site_count": 1, "alert_count": 3, "observation_count": 2, "unlocated_alert_count": 2,
        })
        self.assertEqual(self.digest(self.original), self.meta["raw_sha256"])
        self.assertEqual(self.digest(self.derivative), self.meta["sanitized_sha256"])

    def test_rejects_both_digest_mismatches(self):
        for field in ("raw_sha256", "sanitized_sha256"):
            with self.subTest(field=field):
                original = self.meta[field]
                self.meta[field] = "0" * 64
                self.assert_rejected(self.invoke(), f"{field} differs")
                self.meta[field] = original

    def test_rejects_unchanged_derivative(self):
        self.derivative.write_bytes(self.original.read_bytes())
        self.meta["sanitized_sha256"] = self.meta["raw_sha256"]
        self.assert_rejected(self.invoke(), "must differ from the original")

    def test_rejects_changed_instance_counts_at_same_total(self):
        alerts = self.report["site"][0]["alerts"]
        alerts[1]["instances"] = [alerts[0]["instances"].pop()]
        self.derivative.write_text(json.dumps(self.report), encoding="utf-8")
        self.meta["sanitized_sha256"] = self.digest(self.derivative)
        self.assert_rejected(self.invoke(), "changed site, alert or instance counts by source position")

    def test_rejects_repository_synthetic_fixture(self):
        self.original.write_bytes(SYNTHETIC_FIXTURE.read_bytes())
        self.meta["raw_sha256"] = self.digest(self.original)
        self.assert_rejected(self.invoke(), "original report is labeled synthetic")

    def test_reports_location_of_a_missing_required_field(self):
        del self.report["site"][0]["alerts"][0]["instances"][1]["method"]
        self.derivative.write_text(json.dumps(self.report), encoding="utf-8")
        self.meta["sanitized_sha256"] = self.digest(self.derivative)
        self.assert_rejected(self.invoke(), "/site/0/alerts/0/instances/1/method")

    def test_requires_authorization_and_rfc3339_scan_time(self):
        for field, value, message in (
            ("authorization_ref", None, "authorization reference"),
            ("scan_time", "2020-02-30T12:00:00Z", "valid RFC3339"),
            ("scan_time", "2020-01-01T12:00:00+00:60", "must be RFC3339"),
            ("scan_time", None, "must be RFC3339"),
        ):
            with self.subTest(field=field, value=value):
                original = self.meta[field]
                self.meta[field] = value
                self.assert_rejected(self.invoke(), message)
                self.meta[field] = original

    def test_rejects_duplicate_provenance_keys(self):
        payload = json.dumps(self.meta).removesuffix("}") + ',"raw_sha256":"duplicate"}'
        self.assert_rejected(self.invoke(payload), "no duplicate keys")

    def test_cli_error_keeps_unexpected_path_private(self):
        unexpected_path = "/fictional/private-path-marker.json"
        result = self.invoke(extra_arguments=(unexpected_path,))
        self.assertEqual(result.returncode, 2)
        self.assertEqual(result.stdout, "")
        self.assertIn("Use --help", result.stderr)
        self.assertNotIn(unexpected_path, result.stderr)


if __name__ == "__main__":
    unittest.main()
