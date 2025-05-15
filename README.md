# Brigham Young Chain (BYC)

Brigham Young Chain is a blockchain implementation with a unique three-tier mining system featuring Leah, Shiblum, and Shiblon coins, this blockchain is based on the Nephites monetary system found in The Book of Mormon, Alma 11.

## Quick Start

### Download Binaries

Download the appropriate binary for your operating system from the [releases page](https://github.com/yourusername/byc/releases).

### Running a BYC Node

To run a BYC node:

```bash
# Linux/macOS
./bycnode_linux_amd64_0.1.0 -type full -port 8333

# Windows
bycnode_windows_amd64_0.1.0.exe -type full -port 8333
```

Options:
- `-type`: Node type (full, miner, light)
- `-port`: Port to listen on (default: 8333)

### Running a BYC Miner

To run a BYC miner:

```bash
# Linux/macOS
./bycminer_linux_amd64_0.1.0 -node localhost:8333 -coin leah -threads 4

# Windows
bycminer_windows_amd64_0.1.0.exe -node localhost:8333 -coin leah -threads 4
```

Options:
- `-node`: BYC node address to connect to (default: localhost:8333)
- `-coin`: Coin type to mine (leah, shiblum, shiblon)
- `-threads`: Number of mining threads (default: number of CPU cores)
- `-type`: Mining type (solo, pool)
- `-pool`: Pool address (required for pool mining)
- `-wallet`: Wallet address to receive mining rewards (required for pool mining)

## Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/yourusername/byc.git
cd byc

# Build the node
go build -o bycnode ./cmd/youngchain

# Build the miner
go build -o bycminer ./cmd/bycminer
```

## Development

### Prerequisites

- Go 1.16 or higher
- Git

### Project Structure

- `cmd/`: Command-line applications
  - `youngchain/`: BYC node implementation
  - `bycminer/`: BYC miner implementation
- `internal/`: Internal packages
  - `core/`: Core blockchain functionality
  - `network/`: Network communication
  - `storage/`: Data storage

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Features

### Dual Blockchain System
- Golden Block Chain
- Silver Block Chain
- Independent but interconnected economies

### Mining System
- Three-tier mining difficulty:
  - Leah (easiest)
  - Shiblum (medium)
  - Shiblon (hardest)
- Progressive difficulty adjustment
- Proof of Work consensus

### Coin System
#### Mineable Coins (Both Chains)
- Leah (base unit)
- Shiblum (2x difficulty)
- Shiblon (4x difficulty)

#### Gold-based Derived Units
- Senine
- Seon
- Shum
- Limnah
- Antion (cross-chain transfer)

#### Silver-based Derived Units
- Senum
- Amnor
- Ezrom
- Onti

#### Special Block Completion Coins
- Ephraim (Golden Block completion)
- Manasseh (Silver Block completion)
- Joseph (1 Ephraim + 1 Manasseh)

### Supply Limits
- 15 million Ephraim coins (only 3 million can be converted to Joseph)
- 15 million Manasseh coins (only 3 million can be converted to Joseph)
- 3 million Joseph coins (created by combining 1 Ephraim + 1 Manasseh)

### Network Features
- P2P network
- Node types (full, miner, light)
- Transparent transactions
- No pre-mining
- Open source

## Getting Started

### Prerequisites
- Go 1.16 or higher
- Git

### Installation
```bash
git clone https://github.com/yourusername/brigham-young-chain.git
cd brigham-young-chain
go mod download
```

### Running a Node
```bash
# Full node
go run cmd/youngchain/main.go -type=full -port=8333

# Mining node
go run cmd/youngchain/main.go -type=miner -port=8334

# Light node
go run cmd/youngchain/main.go -type=light -port=8335
```

## Architecture

### Block Structure
- Block header
- Transaction list
- Mining information
- Chain-specific data

### Mining Process
1. Select mining difficulty (Leah, Shiblum, Shiblon)
2. Create block with transactions
3. Perform proof of work
4. Broadcast to network

### Cross-Chain Transfers
- Only Antion can move between chains
- 1 Antion = 3 Shiblons
- Special validation for cross-chain transactions

## Contributing
Contributions are welcome! Please read our contributing guidelines before submitting pull requests.

## Acknowledgments
- The Book of Mormon was translated by the power of God by The Prophet Joseph Smith Junior
- This block chain has no link to The Church of Jesus Christ of Latter Day Saints
- Inspired by the Nephite monetary system
- Built on blockchain principles
- Community-driven development # BYC
# BYC
