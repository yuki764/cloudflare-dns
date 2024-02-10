FROM golang:1.22.0-bookworm as build

WORKDIR /app

COPY go.mod ./
COPY pkg/ ./pkg
COPY cmd/ddns/ ./

RUN CGO_ENABLED=0 go build -o ddns

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /app/ddns /cloudflare-dns-ddns
CMD ["/cloudflare-dns-ddns"]
