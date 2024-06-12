# Building 
FROM golang:1.22 as builder
RUN go install github.com/swaggo/swag/cmd/swag@latest
WORKDIR /app
COPY . .
RUN swag init
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pwctl-be .

# Creating the final docker image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pwctl-be .
#COPY --from=builder /app/docs/* ./docs
CMD ["./pwctl-be", "ttyUSB", "5", "0"]
#CMD ["./pwctl-be", "ttyACM", "5", "0"]