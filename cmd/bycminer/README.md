# BYC Miner

BYC Miner is a command-line tool for mining Brigham Young Chain (BYC) cryptocurrency. It supports both solo mining and pool mining, and can mine different coin types (Leah, Shiblum, Shiblon).

## Features

- Multi-threaded CPU mining
- Support for different coin types with varying difficulties
- Solo and pool mining modes
- Real-time mining statistics
- Wallet integration for receiving mining rewards

## Installation

### Prerequisites

- Go 1.16 or higher
- A BYC wallet address

### Building from Source

1. Clone the Brigham Young Chain repository:
   ```
   git clone https://github.com/yourusername/brigham-young-chain.git
   cd brigham-young-chain
   ```

2. Build the mining client:
   ```
   go build -o bycminer ./cmd/bycminer
   ```

## Usage

### Basic Usage

```
./bycminer -wallet YOUR_WALLET_ADDRESS
```

### Command-line Options

- `-type`: Mining type (solo or pool), default: solo
- `-pool`: Pool address (required for pool mining)
- `-wallet`: Wallet address to receive mining rewards (required)
- `-threads`: Number of mining threads, default: number of CPU cores
- `-coin`: Coin type to mine (leah, shiblum, shiblon), default: leah
- `-node`: BYC node address to connect to, default: localhost:8333

### Examples

#### Solo Mining Leah Coins

```
./bycminer -wallet YOUR_WALLET_ADDRESS -coin leah
```

#### Pool Mining Shiblum Coins with 8 Threads

```
./bycminer -type pool -pool pool.byc.com:3333 -wallet YOUR_WALLET_ADDRESS -coin shiblum -threads 8
```

## Mining Difficulty

BYC has three different coin types with varying mining difficulties:

1. **Leah**: Base difficulty (1x)
2. **Shiblum**: Medium difficulty (2x harder)
3. **Shiblon**: High difficulty (4x harder)

The difficulty automatically adjusts based on the network's hashrate to maintain a target block time of 10 minutes.

## Mining Rewards

Mining rewards are distributed as follows:

- **Leah**: 50 BYC per block
- **Shiblum**: 25 BYC per block
- **Shiblon**: 12.5 BYC per block

Rewards are sent to the wallet address specified when starting the miner.

## Troubleshooting

### Common Issues

- **"Wallet not found"**: Make sure you're using a valid wallet address. If you don't have one, the miner will create a new wallet for you.
- **"Failed to connect to node"**: Check that the BYC node is running and accessible at the specified address.
- **Low hashrate**: Try increasing the number of threads or using a more powerful CPU.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 