FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bsl .

FROM node:22-alpine AS web-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/bsl .
COPY --from=go-builder /app/config.yaml .
COPY --from=web-builder /app/dist ./web/dist
EXPOSE 8080
CMD ["./bsl"]
