FROM golang:1.25-alpine AS builder

#RUN apk update && apk add --no-cache gcc libc-dev make

WORKDIR /app

COPY . .

RUN go mod download && go build -o ./bin/watchtower ./cmd/watchtower


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/configs /app/configs

ENTRYPOINT [ "/app/bin/watchtower" ]

EXPOSE 2893
