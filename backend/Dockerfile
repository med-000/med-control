FROM golang:1.25-alpine AS build

WORKDIR /src

COPY backend/go.mod ./backend/go.mod
COPY infra/go.mod ./infra/go.mod
COPY shared/go.mod ./shared/go.mod

WORKDIR /src/backend
RUN go mod download

WORKDIR /src
COPY backend ./backend
COPY infra ./infra
COPY shared ./shared

WORKDIR /src/backend
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/backend ./cmd/app

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /out/backend /app/backend

EXPOSE 8080

ENTRYPOINT ["/app/backend"]
