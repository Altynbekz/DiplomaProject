#!/usr/bin/env python3
"""Create ignored local secrets once without printing them."""
from pathlib import Path
import os
import secrets
import sys

root = Path(__file__).resolve().parent.parent
os.umask(0o077)
secret_dir = root / ".secrets"
secret_dir.mkdir(mode=0o700, exist_ok=True)
secret_dir.chmod(0o700)
compose = "--compose" in sys.argv
if compose:
    compose_dir = secret_dir / "compose"
    compose_dir.mkdir(mode=0o700, exist_ok=True)

for name in ("api_token", "db_password", "runtime_password"):
    secret = secret_dir / name
    if not secret.exists():
        with secret.open("x") as handle:
            handle.write(secrets.token_hex(32))
    if compose:
        # Compose bind-mounts each file. Its PostgreSQL and Go service UIDs
        # need read access inside their containers. The host parent stays 0700.
        mounted_secret = compose_dir / name
        if not mounted_secret.exists():
            with mounted_secret.open("xb") as handle:
                handle.write(secret.read_bytes())
            mounted_secret.chmod(0o444)
        elif mounted_secret.read_bytes() != secret.read_bytes():
            raise SystemExit("Compose secret differs from the local copy. Resolve credential rotation before continuing.")

env = root / ".env"
if not env.exists():
    with env.open("x") as handle:
        handle.write("APP_OPERATOR_ID=local-operator\nJUICE_SHOP_ORIGIN=http://juice-shop:3000\n")

print("Local secret files ready under .secrets (contents not printed).")
