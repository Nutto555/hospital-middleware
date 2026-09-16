FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -o /out/hismock ./cmd/hismock

FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
COPY --from=build /out/api /out/hismock /usr/local/bin/
USER app
CMD ["api"]
