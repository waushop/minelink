# Use golang:1.20 as the base image
FROM golang:1.20

# Set the working directory to /app
WORKDIR /app

# Copy the project files into the container
COPY . .

# Build the Go application
RUN go build -o minelink

# Set the entrypoint
ENTRYPOINT ["./minelink"]
