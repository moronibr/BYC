# BYC Blockchain

A blockchain implementation with support for multiple coin types and special coins.

## Features

- Multiple coin types (Leah, Shiblum, Senum)
- Special coins (Ephraim, Manasseh, Joseph)
- P2P network support
- RESTful API
- Secure wallet management
- LevelDB storage

## Installation

### Prerequisites

- Go 1.16 or higher
- LevelDB development files

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/byc.git
cd byc

# Build the project
make build

# Run tests
make test
```

## Usage

### Starting a Node

```bash
# Start a node with default configuration
./bin/byc node start

# Start a node with custom port
./bin/byc node start --port 8000
```

### API Endpoints

#### Wallet Operations

- `GET /balance/{address}/{coinType}` - Get balance for specific coin
- `GET /balances/{address}` - Get all coin balances
- `POST /transaction` - Create a new transaction
- `POST /special/ephraim` - Create Ephraim coin
- `POST /special/manasseh` - Create Manasseh coin
- `POST /special/joseph` - Create Joseph coin

#### Blockchain Operations

- `GET /block/{hash}` - Get block information
- `GET /transaction/{id}` - Get transaction information

### Example API Usage

```bash
# Get balance
curl http://localhost:8000/balance/address123/Leah

# Create transaction
curl -X POST http://localhost:8000/transaction \
  -H "Content-Type: application/json" \
  -d '{"to": "recipient123", "amount": 10.0, "coinType": "Leah"}'

# Create special coin
curl -X POST http://localhost:8000/special/ephraim
```

## Development

### Project Structure

```
.
├── cmd/            # Command-line interface
├── internal/       # Internal packages
│   ├── api/       # REST API implementation
│   ├── blockchain/# Core blockchain logic
│   ├── network/   # P2P network implementation
│   ├── storage/   # Data storage
│   └── wallet/    # Wallet management
├── pkg/           # Public packages
└── tests/         # Test files
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test package
go test ./internal/blockchain/...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors
- Inspired by various blockchain implementations
