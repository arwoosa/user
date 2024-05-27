# Get the latest golang image
FROM golang:1.21 as base

# Set the Current Working Directory inside the container
WORKDIR /go/src/oosa_rewild

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go mod download -x

# Install compile daemon for hot reloading
# RUN go install -mod=mod github.com/githubnemo/CompileDaemon
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

FROM gcr.io/distroless/static-debian11

COPY --from=base /main .

# Expose port 80 to the outside world
EXPOSE 6080

# Command to run the executable
# ENTRYPOINT CompileDaemon -build="go build main.go" -command="./main"
CMD ["./main"]