# JWT Token Generator

This is a simple JWT token generator written in Go. It creates a JWT token signed with an RSA private key and includes a custom expiration date.

## Prerequisites

- Go 1.15 or later
- RSA key pair (private key in PEM format)

## Installation

1. **Clone the repository (or download the jwt_generator.go file):**

    ```sh
    git clone https://github.com/yourusername/jwt-generator.git
    cd jwt-generator
    ```

2. **Install dependencies:**

    Ensure you have the `golang-jwt/jwt/v5` library installed:

    ```sh
    go get github.com/golang-jwt/jwt/v5
    ```

3. **Generate an RSA key pair:**

    Use OpenSSL or a similar tool to generate an RSA key pair (in PEM format):

    ```sh
    openssl genrsa -out private.pem 2048
    openssl rsa -in private.pem -pubout -out public.pem
    ```

4. **Compile the program:**

    ```sh
    go build -o jwt_generator main.go
    ```

## Usage

Run the compiled binary to generate a JWT token:

```sh
./jwt_generator -private-key=private.pem -sub=user@example.com -exp=24h -kid=key1
```
