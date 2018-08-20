# Build dev image first
FROM golang:1.8.3-jessie as builder
# Copy files to appropriate locations
WORKDIR /go/
COPY . .
# Prepare build
RUN go-wrapper download .
RUN go-wrapper install .
# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o tictactoe-server-linux-amd64 .

# Build prod image
FROM alpine as runner
RUN apk --no-cache add bash
RUN apk --no-cache add ca-certificates && update-ca-certificates
WORKDIR /root
COPY --from=builder /go//tictactoe-server-linux-amd64 .
ENTRYPOINT ["./tictactoe-server-linux-amd64"]
