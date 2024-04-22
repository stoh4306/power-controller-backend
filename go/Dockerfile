# Build stage
FROM golang:1.19-buster AS builder
RUN apt-get update && apt-get install -y \
    g++ \
    cmake \
    && rm -rf /var/lib/apt/list/*
WORKDIR /app
COPY . .
RUN ls -al
COPY ./cpp /app/cpp


RUN mkdir /app/cpp/build
RUN cd /app/cpp/build && cmake .. -DCMAKE_BUILD_TYPE=Release && make && cp libpwctrlbe.so /lib/x86_64-linux-gnu/

RUN go mod init grida/m2m/pwctrl-be-cpp-go
RUN go get github.com/gin-gonic/gin

RUN CGO_ENABLED=1 go build -o pwctrl-be-app .

#ENTRYPOINT [ "/app/pwctrl-be-app", "ttyUSB", "100", "0" ]


# Final stage
#FROM scratch
#WORKDIR /
#COPY --from=builder /app/pwctrl-be-app /app
#COPY --from=builder /lib/x86_64-linux-gnu/libpwctrlbe.so /lib/x86_64-linux-gnu/
#RUN ls -al
#EXPOSE 8080
#ENTRYPOINT ["/app", "ttyUSB", "100", "0"]
