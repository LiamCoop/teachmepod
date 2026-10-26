FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /podcastd ./cmd/podcastd

FROM debian:trixie-slim
# ffmpeg mixes episodes; ca-certificates for TLS to Postgres, the bucket and the APIs.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /podcastd /usr/local/bin/podcastd
USER nobody
ENTRYPOINT ["podcastd"]
CMD ["serve"]
