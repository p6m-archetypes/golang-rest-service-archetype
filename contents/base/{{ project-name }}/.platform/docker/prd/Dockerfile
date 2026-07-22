FROM golang:1.23 AS builder
WORKDIR /app
# A fresh render carries no go.sum (the dependency set is decided at render time), so the
# lockfile is produced inside the build: `go.*` copies whichever of go.mod/go.sum exist, and
# `go mod tidy` completes go.sum before the compile. Run `make setup` and commit go.sum to pin.
COPY go.* ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o server ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/server /server
EXPOSE {{ service_port }}
EXPOSE {{ management_port }}
ENTRYPOINT ["/server"]
