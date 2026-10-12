FROM golang:1.26-alpine AS base
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

# docker build --target test .
FROM base AS test
RUN go test -v ./...

FROM base AS build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/server"]
