# Build App Using golang Image  
FROM golang:latest as builder

# Add Source Code
WORKDIR /go/src/github.com/Tri125/HoP
COPY . .

# Get Dependencies
RUN go get -u github.com/golang/dep/cmd/dep
RUN $GOPATH/bin/dep ensure --vendor-only

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o HoP .

# Create Lightweight Docker Image 
FROM scratch

# Copy Binary From Builder
COPY --from=builder /go/src/github.com/Tri125/HoP/HoP /app/
WORKDIR /app

# Copy In Certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Entrypoint/CMD
CMD ["./HoP"]