# Use the official Go image as a parent image.
FROM centos:7

# Set the working directory inside the container to /app.
WORKDIR /app

# Install necessary packages for Go installation and your project
RUN yum update -y && \
    yum install -y wget tar gcc && \
    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz && \
    rm go1.21.0.linux-amd64.tar.gz

# Set PATH so that the Go binary is usable
ENV PATH=$PATH:/usr/local/go/bin

# Set environment variables.
ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct

# Although we're not copying the source code into the image,
# you might want to copy other files such as a go.mod and go.sum to cache dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Your build command will be specified in docker-compose.yml
