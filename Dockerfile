# Build stage
FROM golang:latest AS build-env

WORKDIR /app
COPY . .

RUN go build -o nuclearpond .

# Runtime stage
FROM gcr.io/distroless/base

COPY --from=build-env /app/nuclearpond /app/nuclearpond

ENTRYPOINT ["/app/nuclearpond"]

# Run commands on startup
CMD ["service"]