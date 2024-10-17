# Go Server for Live User Status

This project is a Go-based HTTP server that provides APIs to track and fetch live user statuses using Redis for storage. It is optimized to handle high traffic, supporting up to 300,000+ customers per minute concurrently, thanks to Redis connection pooling and batch operations.

## Features

- **Heartbeat API**: Receives a heartbeat from clients to update their last activity timestamp.
- **Status API**: Retrieves the last activity timestamps of multiple users in a single request using Redis batch MGET for performance optimization.
- **Redis Connection Pooling**: Efficient management of Redis connections using gomodule/redigo.

## APIs

### Heartbeat API

- **Endpoint**: `/heartbeat`
- **Method**: POST
- **Request Body**:
  ```json
  {
    "userId": "user456"
  }
  ```
- **Description**: Updates the last activity timestamp for the given `userId`.

Example using `curl`:
```bash
curl --location 'http://localhost:8080/heartbeat' \
--header 'Content-Type: application/json' \
--data '{
    "userId": "user456"
}'
```

### Status API

- **Endpoint**: `/status`
- **Method**: POST
- **Request Body**:
  ```json
  {
    "userIds": [
      "user123",
      "user456"
    ]
  }
  ```
- **Description**: Retrieves the last activity timestamps for the given list of `userIds`.

Example using `curl`:
```bash
curl --location 'http://localhost:8080/status' \
--header 'Content-Type: application/json' \
--data '{
    "userIds": [
        "user123",
        "user456"
    ]
}'
```

## Running the Server

### Prerequisites

- Docker installed.
- Docker Compose installed.

### Steps to Run

1. Clone this repository.
2. Build and run the server using Docker Compose:
   ```bash
   docker-compose up --build
   ```
3. The server will be available at `http://localhost:8080`.

## Load Testing

You can load test the server using the included Go load tester.

1. Navigate to the `load-test` directory.
2. Run the load tester:
   ```bash
   go run load-test/load_tester.go
   ```

This will simulate multiple concurrent requests to test the server's capacity to handle high traffic.

## License

This project is licensed under the MIT License.