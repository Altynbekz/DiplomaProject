#!/usr/bin/env python3
"""Exercise Compose with a new project, disposable credentials and ephemeral ports."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import secrets
import subprocess
import tempfile
from urllib.error import HTTPError
from urllib.request import ProxyHandler, Request, build_opener

ROOT = Path(__file__).resolve().parent.parent
ORIGIN = "http://juice-shop:3000"
FIXTURE = ROOT / "backend/testdata/zap/synthetic-traditional.json"


def command(args, *, capture=False):
    result = subprocess.run(
        args, cwd=ROOT, check=True, text=True, stdout=subprocess.PIPE if capture else None
    )
    return result.stdout.strip() if capture else None


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def compose_override(work):
    # The temporary directory is private on the host. In containers the
    # per-file bind mounts need to be readable by the non-root service UIDs.
    paths = {}
    for key in ("api_token", "db_password", "runtime_password"):
        path = work / key
        path.write_text(secrets.token_hex(32), encoding="ascii")
        path.chmod(0o444)
        paths[key] = str(path)
    override = work / "smoke.yaml"
    lines = [
        "services:",
        "  api:",
        "    environment:",
        "      APP_OPERATOR_ID: compose-smoke",
        f"      JUICE_SHOP_ORIGIN: {ORIGIN}",
        "    ports: !override",
        "      - {target: 8080, host_ip: 127.0.0.1}",
        "  frontend:",
        "    ports: !override",
        "      - {target: 5173, host_ip: 127.0.0.1}",
        "secrets:",
    ]
    for name, path in paths.items():
        lines.extend((f"  {name}:", f"    file: {json.dumps(path)}"))
    override.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return override, paths


def request(base, path, expected, *, token=None, data=None, content_type=None):
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if content_type:
        headers["Content-Type"] = content_type
    opener = build_opener(ProxyHandler({}))
    try:
        response = opener.open(Request(base + path, data=data, headers=headers), timeout=20)
    except HTTPError as exc:
        response = exc
    with response:
        require(response.status == expected, f"{path}: expected HTTP {expected}, got {response.status}")
        return response.read()


def exercise(compose, token):
    def listen(service, port):
        binding = compose("port", service, str(port), capture=True)
        host, _, published = binding.rpartition(":")
        require(host == "127.0.0.1" and published.isdigit() and int(published) > 0,
                f"Expected an ephemeral loopback port for {service}, got {binding!r}")
        return f"http://127.0.0.1:{published}"

    expected_revision = max(int(p.name.split("_", 1)[0]) for p in (ROOT / "backend/db/migrations").glob("*.sql"))

    def query(sql):
        return compose("exec", "-T", "db", "psql", "-XAt", "-v", "ON_ERROR_STOP=1",
                       "-U", "diploma_migrate", "-d", "diploma", "-c", sql, capture=True)

    def check_stack():
        api, frontend = listen("api", 8080), listen("frontend", 5173)
        request(api, "/healthz", 200)
        ready = json.loads(request(api, "/readyz", 200))
        require(ready["status"] == "ready", "Unexpected readiness response")
        compose("exec", "-T", "frontend", "node", "-e",
                "const fs = require('node:fs'); fs.accessSync('node_modules', fs.constants.W_OK)")
        request(frontend, "/", 200)
        request(frontend, "/src/app/main.tsx", 200)
        request(frontend, "/api/v1/targets", 401)
        revision = query("SELECT version_id FROM goose_db_version WHERE is_applied ORDER BY id DESC LIMIT 1")
        require(revision == str(expected_revision), "Unexpected goose migration revision")
        return frontend

    frontend = check_stack()
    target = json.loads(request(frontend, "/api/v1/targets", 201, token=token,
        data=b'{"name":"Synthetic Compose lab","lab_target_key":"juice-shop"}',
        content_type="application/json"))
    require(target["canonical_origin"] == ORIGIN, "Incorrect registered target")
    raw = FIXTURE.read_bytes()
    boundary = secrets.token_hex(20)
    body = (f"--{boundary}\r\n"
            'Content-Disposition: form-data; name="report"; filename="synthetic.json"\r\n'
            "Content-Type: application/json\r\n\r\n").encode() + raw + f"\r\n--{boundary}--\r\n".encode()
    upload = f"/api/v1/targets/{target['id']}/imports/zap"
    args = {"token": token, "data": body, "content_type": f"multipart/form-data; boundary={boundary}"}
    imported = json.loads(request(frontend, upload, 201, **args))
    duplicate = json.loads(request(frontend, upload, 200, **args))
    digest = hashlib.sha256(raw).hexdigest()
    require(imported["report_sha256"] == digest, "Unexpected report hash")
    require((imported["finding_count"], imported["observation_count"], imported["unlocated_alert_count"]) == (4, 5, 1),
        "Unexpected counts for the synthetic fixture")
    require(duplicate["duplicate"] and duplicate["id"] == imported["id"], "Duplicate created a new import")
    require(query("SELECT count(*) FROM report_imports") == "1", "Expected one report in PostgreSQL")
    require(query("SELECT count(*) FROM observations") == "5", "Expected five observations in PostgreSQL")
    print("Compose clean startup, migration, readiness, auth, synthetic import and duplicate checks passed.")

    compose("down", "--timeout", "15")
    require(not compose("ps", "--all", "--quiet", capture=True), "Test containers remain after shutdown")
    compose("up", "--wait", "--wait-timeout", "180")
    frontend = check_stack()
    persisted = json.loads(request(frontend, f"/api/v1/imports/{imported['id']}", 200, token=token))
    require(persisted["report_sha256"] == digest, "Report did not persist across restart")
    findings = json.loads(request(frontend, f"/api/v1/targets/{target['id']}/findings", 200, token=token))
    require(len(findings["items"]) == 4, "Findings did not persist across restart")
    print("Compose shutdown and persistence across restart passed (synthetic fixture).")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config-only", action="store_true", help="Check effective isolation settings without starting Docker")
    options = parser.parse_args()
    project = f"diploma-smoke-{secrets.token_hex(10)}"
    with tempfile.TemporaryDirectory(prefix="diploma-compose-") as directory:
        work = Path(directory)
        override, paths = compose_override(work)
        env_file = work / "isolated.env"
        env_file.touch()
        base = ["docker", "compose", "--project-name", project, "--env-file", str(env_file),
                "-f", str(ROOT / "compose.yaml"), "-f", str(override)]

        def compose(*args, capture=False):
            return command([*base, *args], capture=capture)

        config = json.loads(compose("config", "--format", "json", capture=True))
        require(config["name"] == project, "Refusing to operate on a different Compose project")
        require(set(config["services"]) == {"db", "migrate", "api", "frontend"}, "Unexpected Compose services")
        require(set(config["volumes"]) == {"pgdata"}, "Unexpected Compose volumes")
        volume = config["volumes"]["pgdata"]
        require(volume["name"] == f"{project}_pgdata" and not volume.get("external") and not volume.get("driver_opts"),
                "Refusing to use a shared volume")
        networks = config["networks"]
        require(set(networks) == {"storage", "app"}, "Unexpected Compose networks")
        require(all(network["name"] == f"{project}_{name}" and not network.get("external")
                    for name, network in networks.items()), "Refusing to use a shared network")
        require(networks["storage"].get("internal") and networks["app"].get("driver") == "bridge"
                and not networks["app"].get("internal"), "Unexpected storage or app network configuration")
        for name in ("db", "migrate"):
            require(set(config["services"][name]["networks"]) == {"storage"},
                    f"{name} must use only the internal storage network")
        for key, path in paths.items():
            require(config["secrets"][key]["file"] == path, "Refusing to use non-disposable credentials")
        for name in ("api", "frontend"):
            ports = config["services"][name]["ports"]
            require(len(ports) == 1 and ports[0].get("host_ip") == "127.0.0.1"
                    and ports[0].get("published") in (None, "", "0"), "Expected one ephemeral loopback port")
        require(not config["services"]["db"].get("ports"), "The test database must not publish a port")
        if options.config_only:
            print("Compose test configuration is isolated. Container startup was not exercised.")
            return
        endpoint = os.environ.get("DOCKER_HOST")
        if env_context := os.environ.get("DOCKER_CONTEXT"):
            endpoint = command(["docker", "context", "inspect", env_context,
                "--format", "{{.Endpoints.docker.Host}}"], capture=True)
        if not endpoint:
            endpoint = command(["docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}"], capture=True)
        require(endpoint.startswith("unix://"), "The smoke test requires a local Unix-socket Docker engine")
        architecture = command(["docker", "info", "--format", "{{.Architecture}}"], capture=True)
        require(architecture in {"aarch64", "arm64"}, "Compose images are pinned to Linux arm64")
        require(not compose("ps", "--all", "--quiet", capture=True), "The test project already has containers")
        volumes = command(["docker", "volume", "ls", "--format", "{{.Name}}"], capture=True)
        require(volume["name"] not in volumes.splitlines(), "The test volume already exists")
        try:
            compose("up", "--build", "--wait", "--wait-timeout", "180")
            exercise(compose, Path(paths["api_token"]).read_text(encoding="ascii"))
        except (OSError, RuntimeError, subprocess.CalledProcessError):
            # Only logs from this disposable project are requested.
            subprocess.run([*base, "logs", "--tail", "30"], cwd=ROOT, check=False)
            raise
        finally:
            # This unique project and its verified name are owned by this run.
            compose("down", "--volumes", "--timeout", "15")
            images = command(["docker", "image", "ls", "--filter", f"reference={project}-*",
                              "--format", "{{.Repository}}:{{.Tag}}"], capture=True).splitlines()
            owned_images = {f"{project}-{name}:latest" for name in config["services"]}
            images = sorted(owned_images.intersection(images))
            if images:
                command(["docker", "image", "rm", *images])


if __name__ == "__main__":
    try:
        main()
    except (OSError, RuntimeError, subprocess.CalledProcessError) as exc:
        raise SystemExit(str(exc)) from None
