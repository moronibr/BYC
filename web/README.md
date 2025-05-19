# Book of Mormon Coin Web Interface

This is the web interface for the Book of Mormon Coin (BMC) node. It provides a user-friendly way to interact with your BMC node, manage wallets, and monitor the network.

## Features

- **Dashboard**: Monitor node status, network statistics, and system metrics
- **Wallet Management**: Create wallets, view balances, and send transactions
- **Network Monitoring**: View connected peers and network health
- **Real-time Updates**: Automatic refresh of data
- **Responsive Design**: Works on desktop and mobile devices

## Prerequisites

- Node.js 16.x or later
- npm 8.x or later
- BMC node running on localhost:8545 (or configure API_URL)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/book-of-mormon-coin.git
   cd book-of-mormon-coin/web
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Create a `.env` file:
   ```bash
   cp .env.example .env
   ```

4. Configure environment variables:
   ```
   REACT_APP_API_URL=http://localhost:8545
   ```

## Development

Start the development server:
```bash
npm start
```

The application will be available at http://localhost:3000.

## Building for Production

Build the application:
```bash
npm run build
```

The build artifacts will be stored in the `build/` directory.

## Testing

Run tests:
```bash
npm test
```

Run end-to-end tests:
```bash
npm run test:e2e
```

## Project Structure

```
web/
├── public/              # Static files
├── src/
│   ├── components/      # React components
│   ├── hooks/          # Custom React hooks
│   ├── types/          # TypeScript type definitions
│   ├── App.tsx         # Main application component
│   └── index.tsx       # Application entry point
├── package.json        # Project dependencies
└── tsconfig.json      # TypeScript configuration
```

## Components

### Dashboard
- Node status
- Network statistics
- System metrics
- Recent transactions

### Wallet
- Wallet information
- Balance display
- Transaction history
- Send/receive functionality

### Network
- Network health
- Peer connections
- Connection management
- Network statistics

## API Integration

The web interface communicates with the BMC node through a REST API. The API endpoints are:

- Node Management: `/api/v1/node/*`
- Wallet Management: `/api/v1/wallet/*`
- Transaction Management: `/api/v1/tx/*`
- Network Management: `/api/v1/network/*`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details. 