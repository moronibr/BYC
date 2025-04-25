# Brigham Young Chain (BYC)

BYC is a blockchain implementation written in Go, featuring a unique coin system and modern blockchain architecture.

## Features

- Custom coin system (Leah, Shiblon, Ephraim, Manasseh)
- Proof of Work consensus
- P2P networking
- Wallet management
- Mining capabilities
- Security features

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Make

### Installation

1. Clone the repository:
```bash
git clone https://github.com/youngchain/brigham-young-chain.git
cd brigham-young-chain
```

2. Install dependencies:
```bash
make deps
```

3. Build the project:
```bash
make build
```

### Running the Node

To start a BYC node:

```bash
./bin/byc
```

To start mining:

```bash
./bin/bycminer
```

## Configuration

The default configuration file is located at `~/.byc/config.json`. You can specify a different configuration file using the `--config` flag.

Example configuration:

```json
{
    "network": "mainnet",
    "listen_addr": "0.0.0.0:8333",
    "rpc_addr": "127.0.0.1:8334",
    "max_peers": 50,
    "mining_enabled": true,
    "mining_threads": 4,
    "data_dir": "~/.byc/data"
}
```

## Network

BYC uses a P2P network for communication between nodes. The network protocol includes:

- Peer discovery
- Block synchronization
- Transaction propagation
- Mining coordination

## Mining

BYC uses a Proof of Work consensus mechanism. The mining process includes:

- Transaction selection
- Block creation
- Nonce finding
- Block propagation

## Wallet

The wallet system provides:

- Key generation
- Address management
- Transaction signing
- Balance tracking

## Security

Security features include:

- TLS encryption
- Rate limiting
- Input validation
- Security headers
- Authentication

## Development

### Testing

Run tests:

```bash
make test
```

Run linter:

```bash
make lint
```

### Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 