FROM golang:alpine AS builder

COPY . /build
WORKDIR /build
RUN go mod vendor && go build -o main ./cmd/server

FROM alpine:latest  

COPY --from=builder /build/main .

EXPOSE 8000
CMD ["./main"]
