# Build docker image using Dockerfile
docker build -t go-rocky89 .

# Run docker container with privileged option which enables the app access the serial port
docker run -d --rm --privileged -v /dev:/dev -p 8080:8080 go-rocky89