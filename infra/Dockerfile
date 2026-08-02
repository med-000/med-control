FROM golang:1.25-alpine AS build

WORKDIR /src

COPY infra/go.mod ./infra/go.mod
COPY shared/go.mod ./shared/go.mod

WORKDIR /src/infra
RUN go mod download

WORKDIR /src
COPY infra ./infra
COPY shared ./shared

WORKDIR /src/infra
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/infra ./cmd/app

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /out/infra /app/infra

ENTRYPOINT ["/app/infra"]
