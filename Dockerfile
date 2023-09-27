FROM golang:1.17-buster as builder
 
# Create and change to the app directory.
WORKDIR /app
 
# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./

# Set Private repository access
ENV GOPRIVATE=github.com/DX-standarization-Team/common-service
RUN echo "GOPRIVATE set: $GOPRIVATE"
ENV GONOPROXY=github.com/DX-standarization-Team/common-service
RUN echo "GONOPROXY set: $GOPRIVATE"

ARG ACCESS_TOKEN_PRIVATE_REPO
RUN echo "git config set"
RUN git config --global url."https://$ACCESS_TOKEN_PRIVATE_REPO:x-oauth-basic@github.com/DX-standarization-Team/common-service/".insteadOf "https://github.com/DX-standarization-Team/common-service/"

RUN go mod download

# Copy local code to the container image.
COPY . ./
 
# Build the binary.
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
CMD ["/app/server"]
