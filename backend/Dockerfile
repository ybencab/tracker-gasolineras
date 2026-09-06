FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api

FROM alpine:latest
RUN apk add --no-cache curl ca-certificates
WORKDIR /app
COPY --from=build /app/api ./api
ENTRYPOINT ["./api"]
