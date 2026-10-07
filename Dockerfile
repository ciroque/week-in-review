FROM golang:1.23-alpine AS build

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/week-in-review ./cmd/week-in-review

FROM alpine:3.21

RUN apk add --no-cache git ca-certificates tzdata

COPY --from=build /out/week-in-review /usr/local/bin/week-in-review

WORKDIR /repo
ENTRYPOINT ["week-in-review"]
