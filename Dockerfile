FROM golang AS builder
WORKDIR /app
COPY . .
RUN go get -v
RUN CGO_ENABLED=0 go build -o replicator

FROM ubuntu AS certs
RUN  apt-get update && apt-get install -y ca-certificates

FROM scratch
COPY --from=builder /app/replicator /app/replicator
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
ENTRYPOINT ["/app/replicator"]