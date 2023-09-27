FROM golang:1.17-buster as builder
 
# Create and change to the app directory.
WORKDIR /app
 
# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./

# Set Private repository access
# RUN echo "set GOPRIVATE, GONOPROXY, GONOSUMDB: github.com/DX-standarization-Team/common-service-v2"
ENV GOPRIVATE=github.com/DX-standarization-Team/common-service-v2
ENV GONOPROXY=github.com/DX-standarization-Team/common-service-v2
ENV GONOSUMDB=github.com/DX-standarization-Team/common-service-v2
ENV GOPROXY=direct
RUN echo $GOPRIVATE

# ARG TOKEN
# RUN git config --global url."https://x-access-token:${TOKEN}@github.com/DX-standarization-Team/common-service-v2/".insteadOf "https://github.com/DX-standarization-Team/common-service-v2/"
# RUN git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"

# ACCESS TOKEN version
# ARG ACCESS_TOKEN_PRIVATE_REPO
# ARG USER_NAME
# RUN echo "git config set. USER_NAME: $USER_NAME"
# RUN git config --global url."https://$USER_NAME:$ACCESS_TOKEN_PRIVATE_REPO:x-oauth-basic@github.com/DX-standarization-Team/common-service-v2/".insteadOf "https://github.com/DX-standarization-Team/common-service-v2/"
RUN git config --global url."https://$ACCESS_TOKEN_PRIVATE_REPO:x-oauth-basic@github.com/DX-standarization-Team/common-service-v2/".insteadOf "https://github.com/DX-standarization-Team/common-service-v2/"

# SSH version
# RUN git config --global url."ssh://git@github.com".insteadOf "https://github.com"

RUN go mod download
# RUN go install github.com/DX-standarization-Team/common-service-v2@v0.1.0

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
