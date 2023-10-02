# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.20-buster as builder
 
# Create and change to the app directory.
WORKDIR /app
 
# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./

# Set Private repository access
ARG GIT_HUB_ACCOUNT
ENV GOPRIVATE=github.com/${GIT_HUB_ACCOUNT}/*

ARG TOKEN
RUN git config --global url."https://x-access-token:${TOKEN}@github.com/${GIT_HUB_ACCOUNT}/".insteadOf "https://github.com/${GIT_HUB_ACCOUNT}/"

# RUN echo "set GOPRIVATE, GONOPROXY, GONOSUMDB: ${GOPRIVATE}"

RUN go mod download

# Copy local code to the container image.
COPY . ./
 
# Build the binary.（-x: 詳細ログ出力）
RUN go build -v -o server
 
# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
 
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server
 
# Run the web service on container startup.
ARG RUNNING_ENV
ENV RUNNING_ENV2=$RUNNING_ENV
RUN echo $RUNNING_ENV2
# CMD ["/app/server", "-runningEnv=dev1234"]
CMD ["/app/server","-runningEnv=$RUNNING_ENV2"]
# CMD ["/app/server"]
