# ---------- builder ----------
FROM golang:1.25-bookworm AS builder

WORKDIR /build

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /mantis-market-data ./cmd/marketdata

# ---------- runtime ----------
FROM gcr.io/distroless/static-debian12

COPY --from=builder /mantis-market-data /mantis-market-data

EXPOSE 8081

CMD ["/mantis-market-data"]
