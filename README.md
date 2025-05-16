# Brigham Young Chain (BYC)

A dual-blockchain system inspired by the Nephite monetary system as described in the Book of Mormon (Alma 11). This implementation features two interconnected blockchains with unique coin relationships and mining mechanics.

## Architecture

### Golden Block (Ephraim)
- **Mineable Coins**: Leah, Shiblum, Shiblon
- **Non-mineable Coins**: Senine, Seon, Shum, Limnah, Antion
- **Final Form**: Ephraim Coin (15 million limit)

### Silver Block (Manasseh)
- **Mineable Coins**: Leah, Shiblum, Shiblon
- **Non-mineable Coins**: Senum, Amnor, Ezrom, Onti, Antion
- **Final Form**: Manasseh Coin (15 million limit)

### Mining Difficulty (Proof of Work)
- Leah: Base difficulty
- Shiblum: 2x Leah's difficulty
- Shiblon: 2x Shiblum's difficulty (4x Leah's difficulty)

### Special Features
- Antion is the only coin transferable between blocks
- Joseph Coin (3 million limit) can be created by combining Ephraim and Manasseh coins
- Terminal-based interface for node operation, mining, and transactions

## Project Structure
```
byc/
├── cmd/
│   └── byc/           # Main executable
├── internal/
│   ├── blockchain/    # Core blockchain implementation
│   ├── network/       # P2P networking
│   ├── mining/        # Mining implementation
│   ├── wallet/        # Wallet and transaction management
│   └── consensus/     # Consensus mechanisms
├── pkg/
│   ├── crypto/        # Cryptographic utilities
│   └── utils/         # General utilities
├── go.mod
└── README.md
```

## Building and Running

### Prerequisites
- Go 1.20 or higher
- Git

### Installation
```bash
git clone https://github.com/yourusername/byc.git
cd byc
go build -o byc cmd/byc/main.go
```

### Running a Node
```bash
./byc node start
```

### Mining
```bash
./byc mine --coin leah    # Mine Leah coins
./byc mine --coin shiblum # Mine Shiblum coins
./byc mine --coin shiblon # Mine Shiblon coins
```

### Wallet Operations
```bash
./byc wallet create      # Create a new wallet
./byc wallet balance     # Check balance
./byc wallet send        # Send coins
```

## License
MIT License

## Contributing
Contributions are welcome! Please read our contributing guidelines before submitting pull requests.
