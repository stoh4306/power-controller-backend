# Building 
FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pwctl-be .

# Creating the final docker image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pwctl-be .
CMD ["./pwctl-be", "ttyUSB", "5", "0"]