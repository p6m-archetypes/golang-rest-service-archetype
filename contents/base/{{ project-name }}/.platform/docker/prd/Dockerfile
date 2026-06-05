FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o server ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/server /server
EXPOSE {{ service_port }}
EXPOSE {{ management_port }}
ENTRYPOINT ["/server"]
