FROM golang:1.22-alpine AS build

WORKDIR /app
COPY . .
RUN go build -o mailshield .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/mailshield .
COPY config.yaml .
CMD ["./mailshield"]