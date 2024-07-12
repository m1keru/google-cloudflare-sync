FROM golang:alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . /usr/src/app
RUN  cd cmd/google-cloudflare-sync && go build -v -o /usr/local/bin/google-cloudflare-sync .

FROM scratch
COPY --from=builder /usr/local/bin/google-cloudflare-sync /usr/local/bin/google-cloudflare-sync

ENTRYPOINT ["/usr/local/bin/google-cloudflare-sync"]
