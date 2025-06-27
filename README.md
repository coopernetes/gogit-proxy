# Go Reverse Proxy

This project implements an HTTP reverse proxy that maps requests from `/github.com/*` to `https://github.com/*`. 

## Project Structure

```
go-reverse-proxy
├── src
│   ├── main.go        # Entry point of the application
│   └── proxy
│       └── proxy.go   # Implementation of the reverse proxy
├── go.mod             # Module definition
└── README.md          # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd go-reverse-proxy
   ```

2. Navigate to the `src` directory:
   ```
   cd src
   ```

3. Install dependencies:
   ```
   go mod tidy
   ```

### Running the Application

To run the application, execute the following command from the `src` directory:

```
go run main.go
```

The server will start on `localhost:8080`. You can access the reverse proxy by navigating to `http://localhost:8080/github.com/`.

### Example Usage

To test the reverse proxy, you can use a web browser or a tool like `curl`:

```
curl http://localhost:8080/github.com/user/repo
```

This will forward the request to `https://github.com/user/repo`.

### License

This project is licensed under the MIT License. See the LICENSE file for details.