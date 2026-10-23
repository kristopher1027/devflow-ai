FROM golang:1.26-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/devflow-api ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates     && addgroup -S devflow     && adduser -S -G devflow devflow
WORKDIR /app
COPY --from=build /out/devflow-api /app/devflow-api
USER devflow
EXPOSE 8080
ENTRYPOINT ["/app/devflow-api"]
