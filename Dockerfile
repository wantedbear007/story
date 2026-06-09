FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/story ./cmd/story/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -h /home/story story

COPY --from=builder /usr/local/bin/story /usr/local/bin/story
COPY --from=builder /src/configs /etc/story/configs
COPY --from=builder /src/migrations /etc/story/migrations

USER story
WORKDIR /home/story

ENTRYPOINT ["story"]
CMD ["--help"]
