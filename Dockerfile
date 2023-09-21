# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -p 8 -o rancher-import main.go

# Final stage
FROM alpine:latest


#ADD https://releases.hashicorp.com/vault/1.14.2/vault_1.14.2_linux_amd64.zip /tmp/
#
## Install ca-certificates, bash, curl, and libc6-compat
#RUN apk add --no-cache ca-certificates bash curl sed libc6-compat && \
#    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
#    chmod +x kubectl && \
#    mv kubectl /usr/local/bin/  \
#    && unzip /tmp/vault_1.14.2_linux_amd64.zip -d /tmp/ \
#    && mv /tmp/vault /usr/local/bin/  \
#    && rm -rf /tmp/* \
#    && rm -rf /var/cache/apk/*

RUN apk add --no-cache ca-certificates curl  && \
    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Copy binary and bash scripts
COPY --from=builder /app/rancher-import /app/rancher-import

# Set the working directory
WORKDIR /app

CMD ["/app/rancher-import"]
