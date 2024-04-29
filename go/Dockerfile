# Build stage
# Use Rocky Linux as the base image
FROM rockylinux:8.9 as builder

# Install necessary packages
RUN dnf -y update && \
    #dnf -y groupinstall "Development Tools" && \
    dnf -y install wget tar cmake gcc gcc-c++

# Set environment variables for Go installation
ENV GOLANG_VERSION 1.18
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Create the directory and install Go
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && \
    wget -q "https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz" -O go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

# Set the working directory inside the container
#WORKDIR $GOPATH

WORKDIR /app
RUN pwd
COPY . .
RUN ls -al
COPY ./cpp /app/cpp


RUN mkdir ./cpp/build
RUN cd ./cpp/build && cmake .. -DCMAKE_BUILD_TYPE=Release && make
RUN cp ./cpp/build/libpwctrlbe.so /lib64/

RUN go mod init grida/m2m/pwctrl-be-cpp-go
RUN go get github.com/gin-gonic/gin
RUN go get github.com/sirupsen/logrus

RUN CGO_ENABLED=1 go build -o pwctrl-be-app .

#EXPOSE 8080
#ENTRYPOINT [ "/app/pwctrl-be-app", "ttyUSB", "100", "0" ]

#
FROM alpine

RUN apk update --no-cache && apk upgrade --no-cache && apk add --no-cache bash libstdc++ libc6-compat
COPY --from=builder /app/pwctrl-be-app ./pwctrl-be-app
COPY --from=builder /lib64/libpwctrlbe.so /lib/libpwctrlbe.so

EXPOSE 8080
ENTRYPOINT [ "./pwctrl-be-app", "ttyUSB", "100", "0" ]