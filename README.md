# BYC CLI

BYC CLI is a command-line interface for the BYC blockchain network. It provides tools for mining, wallet management, and network operations.

## Installation

### Windows
1. Download `byc-windows.zip`
2. Extract the zip file
3. Double-click `byc.exe` to start the application

### Linux
1. Download `byc-linux.tar.gz`
2. Extract the archive:
   ```bash
   tar -xzf byc-linux.tar.gz
   ```
3. Make the binary executable:
   ```bash
   chmod +x byc
   ```
4. Run the application:
   ```bash
   ./byc
   ```

### macOS
1. Download `byc-macos.tar.gz`
2. Extract the archive:
   ```bash
   tar -xzf byc-macos.tar.gz
   ```
3. Make the binary executable:
   ```bash
   chmod +x byc-mac
   ```
4. Run the application:
   ```bash
   ./byc-mac
   ```

## Usage

The BYC CLI provides a menu-driven interface with the following options:

1. **Network Operations**
   - Start a node
   - Monitor network
   - Connect to peers

2. **Wallet Operations**
   - Create new wallet
   - Check balance
   - Send coins
   - Backup wallet

3. **Dashboard**
   - View network statistics
   - Monitor mining performance
   - Track transactions

4. **Mining**
   - Start mining
   - Configure mining parameters
   - View mining statistics

## Security Features

- Rate limiting for API endpoints
- Strong wallet encryption with Argon2 and AES-GCM
- Transaction signing verification
- Multi-signature support
- HD wallet support

## Requirements

- Windows 10/11, Linux, or macOS
- 4GB RAM minimum
- 10GB free disk space
- Internet connection for network operations

## Support

For issues and feature requests, please visit our GitHub repository.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
