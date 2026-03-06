FROM golang:1.25 AS build

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notify ./cmd/notify

FROM alpine:3.22

RUN adduser -D -g '' appuser
USER appuser

COPY --from=build /out/notify /usr/local/bin/notify

ENTRYPOINT ["/usr/local/bin/notify"]