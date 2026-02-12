FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN apk add --no-cache git
RUN go mod download || true
COPY . .
RUN go build -o /app/textify ./src

FROM alpine:edge
RUN apk add --no-cache ca-certificates
COPY --from=build /app/textify /textify
EXPOSE 3000
ENTRYPOINT ["/textify"]