# Brigham Young Chain (Young Chain)

A dual-blockchain system inspired by the Nephite monetary system, featuring parallel Golden and Silver blockchains with unique mining mechanics and special coins.

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
- 11 million Ephraim coins
- 11 million Manasseh coins
- 11 million Joseph coins

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

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments
- Inspired by the Nephite monetary system
- Built on blockchain principles
- Community-driven development # BYC
# BYC
