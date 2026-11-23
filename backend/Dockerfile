FROM golang:1.27.1@sha256:21758bb7e93fe251a31f1c2390cdaaea748885814d1a8f00fcb37480dc0df0ec AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /api ./cmd/api && go install github.com/pressly/goose/v3/cmd/goose@v3.30.1
FROM golang:1.27.1@sha256:21758bb7e93fe251a31f1c2390cdaaea748885814d1a8f00fcb37480dc0df0ec
WORKDIR /app
COPY --from=build /api /app/api
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY db/migrations /app/migrations
COPY docker-entrypoint.sh /app/entrypoint.sh
USER 65532:65532
ENTRYPOINT ["/app/entrypoint.sh"]
