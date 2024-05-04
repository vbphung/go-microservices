FROM golang:1.18-alpine as builder
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN CGO_ENABLE=0 go build -o brokerApp ./cmd/api && chmod +x /app/brokerApp

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /app/brokerApp /app
CMD [ "/app/brokerApp" ]
