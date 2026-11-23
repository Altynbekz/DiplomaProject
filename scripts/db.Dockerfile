FROM postgres:15.19@sha256:6fd00320c0d70da5caf73c9382162d5a2b0fc718f3861f94687badb6ef8ab1c4
COPY scripts/db-init.sh /docker-entrypoint-initdb.d/10-runtime.sh
