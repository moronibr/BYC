# Getting Started with Book of Mormon Coin

This guide will help you get started with the Book of Mormon Coin (BMC) blockchain implementation. You'll learn how to install, configure, and run your first BMC node.

## Prerequisites

- Go 1.21 or later
- Git
- Make
- Docker (optional, for containerized deployment)
- 4GB RAM minimum (8GB recommended)
- 20GB free disk space
- Linux/macOS/Windows with WSL2

## Installation

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/book-of-mormon-coin.git
   cd book-of-mormon-coin
   ```

2. Build the project:
   ```bash
   make build
   ```

3. Install the binary:
   ```bash
   make install
   ```

### Using Docker

1. Pull the Docker image:
   ```bash
   docker pull yourusername/book-of-mormon-coin:latest
   ```

2. Run the container:
   ```bash
   docker run -d \
     --name bmc-node \
     -p 8545:8545 \
     -p 30303:30303 \
     -v bmc-data:/data \
     yourusername/book-of-mormon-coin:latest
   ```

## Quick Start

1. Initialize a new node:
   ```bash
   bmc init --network mainnet
   ```

2. Start the node:
   ```bash
   bmc start
   ```

3. Check node status:
   ```bash
   bmc status
   ```

## Basic Configuration

The default configuration is suitable for most users. However, you can customize it by editing the `config.json` file:

```json
{
  "network": {
    "type": "mainnet",
    "rpc_url": "http://localhost:8545",
    "p2p_port": 30303
  },
  "security": {
    "encryption_algorithm": "aes-256-gcm",
    "key_derivation_cost": 32768,
    "password_min_length": 12
  }
}
```

See the [Configuration Guide](./configuration.md) for detailed options.

## First Steps

1. **Create a Wallet**
   ```bash
   bmc wallet create
   ```

2. **Get Your Address**
   ```bash
   bmc wallet address
   ```

3. **Check Balance**
   ```bash
   bmc wallet balance
   ```

4. **Send Transaction**
   ```bash
   bmc wallet send --to <address> --amount <amount>
   ```

## Network Types

- **Mainnet**: Production network
- **Testnet**: Testing network with test coins
- **Devnet**: Development network for testing

## Common Commands

```bash
# View help
bmc --help

# Check version
bmc version

# View logs
bmc logs

# Backup wallet
bmc wallet backup

# Restore wallet
bmc wallet restore
```

## Next Steps

1. Read the [Architecture Guide](./architecture.md) to understand the system
2. Check the [Security Guide](./security.md) for best practices
3. Explore the [API Reference](./api-reference.md) for development
4. Join our [Community](https://github.com/yourusername/book-of-mormon-coin/discussions)

## Troubleshooting

If you encounter issues:

1. Check the [Troubleshooting Guide](./troubleshooting.md)
2. View the logs: `bmc logs`
3. Check system requirements
4. Verify network connectivity
5. Create an issue on GitHub

## Support

- [GitHub Issues](https://github.com/yourusername/book-of-mormon-coin/issues)
- [Discord Community](https://discord.gg/your-discord)
- [Documentation](./README.md) 