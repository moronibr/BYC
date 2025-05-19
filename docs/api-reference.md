# API Reference

This document provides detailed information about the Book of Mormon Coin (BMC) APIs, including REST, RPC, and WebSocket endpoints.

## REST API

The REST API is available at `http://localhost:8545` by default.

### Authentication

All API endpoints require authentication using a JWT token. Include the token in the `Authorization` header:

```http
Authorization: Bearer <your-token>
```

### Endpoints

#### Node Information

```http
GET /api/v1/node/info
```

Response:
```json
{
  "version": "1.0.0",
  "network": "mainnet",
  "block_height": 12345,
  "peers": 25,
  "uptime": "5d 12h 30m"
}
```

#### Wallet Operations

##### Create Wallet
```http
POST /api/v1/wallet/create
Content-Type: application/json

{
  "password": "your-password"
}
```

Response:
```json
{
  "address": "0x123...",
  "public_key": "0x456...",
  "encrypted": true
}
```

##### Get Balance
```http
GET /api/v1/wallet/balance
```

Response:
```json
{
  "balance": "100.5",
  "unconfirmed": "0.5",
  "currency": "BMC"
}
```

##### Send Transaction
```http
POST /api/v1/wallet/send
Content-Type: application/json

{
  "to": "0x789...",
  "amount": "10.5",
  "fee": "0.001"
}
```

Response:
```json
{
  "tx_hash": "0xabc...",
  "status": "pending"
}
```

#### Blockchain Operations

##### Get Block
```http
GET /api/v1/block/{height}
```

Response:
```json
{
  "height": 12345,
  "hash": "0xdef...",
  "timestamp": "2024-03-15T12:00:00Z",
  "transactions": 10,
  "size": 1024
}
```

##### Get Transaction
```http
GET /api/v1/tx/{hash}
```

Response:
```json
{
  "hash": "0xabc...",
  "from": "0x123...",
  "to": "0x789...",
  "amount": "10.5",
  "fee": "0.001",
  "status": "confirmed",
  "block_height": 12345
}
```

## RPC API

The RPC API is available at `http://localhost:8545/rpc` by default.

### Methods

#### Node Methods

```json
{
  "jsonrpc": "2.0",
  "method": "bmc_getNodeInfo",
  "params": [],
  "id": 1
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "version": "1.0.0",
    "network": "mainnet",
    "block_height": 12345,
    "peers": 25,
    "uptime": "5d 12h 30m"
  },
  "id": 1
}
```

#### Wallet Methods

##### Create Wallet
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_createWallet",
  "params": ["your-password"],
  "id": 1
}
```

##### Get Balance
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_getBalance",
  "params": ["0x123..."],
  "id": 1
}
```

##### Send Transaction
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_sendTransaction",
  "params": [{
    "to": "0x789...",
    "amount": "10.5",
    "fee": "0.001"
  }],
  "id": 1
}
```

#### Blockchain Methods

##### Get Block
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_getBlock",
  "params": [12345],
  "id": 1
}
```

##### Get Transaction
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_getTransaction",
  "params": ["0xabc..."],
  "id": 1
}
```

## WebSocket API

The WebSocket API is available at `ws://localhost:8545/ws` by default.

### Events

#### New Block
```json
{
  "type": "new_block",
  "data": {
    "height": 12345,
    "hash": "0xdef...",
    "timestamp": "2024-03-15T12:00:00Z"
  }
}
```

#### New Transaction
```json
{
  "type": "new_transaction",
  "data": {
    "hash": "0xabc...",
    "from": "0x123...",
    "to": "0x789...",
    "amount": "10.5"
  }
}
```

#### Peer Connected
```json
{
  "type": "peer_connected",
  "data": {
    "address": "192.168.1.1:30303",
    "version": "1.0.0"
  }
}
```

### Subscriptions

#### Subscribe to New Blocks
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_subscribe",
  "params": ["new_blocks"],
  "id": 1
}
```

#### Subscribe to Transactions
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_subscribe",
  "params": ["transactions"],
  "id": 1
}
```

#### Subscribe to Peers
```json
{
  "jsonrpc": "2.0",
  "method": "bmc_subscribe",
  "params": ["peers"],
  "id": 1
}
```

## Error Codes

| Code | Description |
|------|-------------|
| 1000 | Invalid request |
| 1001 | Authentication failed |
| 1002 | Insufficient funds |
| 1003 | Invalid address |
| 1004 | Invalid amount |
| 1005 | Transaction failed |
| 1006 | Network error |
| 1007 | Rate limit exceeded |

## Rate Limiting

API endpoints are rate-limited to prevent abuse. The default limits are:

- 100 requests per minute for authenticated users
- 10 requests per minute for unauthenticated users

## Best Practices

1. **Error Handling**
   - Always check response status codes
   - Handle rate limiting errors
   - Implement retry logic for network errors

2. **Security**
   - Use HTTPS in production
   - Rotate API keys regularly
   - Validate all input data

3. **Performance**
   - Use WebSocket for real-time updates
   - Implement caching where appropriate
   - Batch requests when possible

## Client Libraries

### Go
```go
import "github.com/yourusername/book-of-mormon-coin/client"

client := client.NewClient("http://localhost:8545")
balance, err := client.GetBalance("0x123...")
```

### JavaScript
```javascript
import { BMCClient } from '@book-of-mormon-coin/client';

const client = new BMCClient('http://localhost:8545');
const balance = await client.getBalance('0x123...');
```

### Python
```python
from bmc_client import BMCClient

client = BMCClient('http://localhost:8545')
balance = client.get_balance('0x123...')
```

## Examples

### Creating a Wallet
```bash
curl -X POST http://localhost:8545/api/v1/wallet/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"password": "your-password"}'
```

### Sending a Transaction
```bash
curl -X POST http://localhost:8545/api/v1/wallet/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "to": "0x789...",
    "amount": "10.5",
    "fee": "0.001"
  }'
```

### Subscribing to New Blocks
```javascript
const ws = new WebSocket('ws://localhost:8545/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({
    jsonrpc: '2.0',
    method: 'bmc_subscribe',
    params: ['new_blocks'],
    id: 1
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('New block:', data);
};
```

## Support

For API support:
1. Check the [Troubleshooting Guide](./troubleshooting.md)
2. Review the [API Examples](./api-examples.md)
3. Join our [Developer Community](https://github.com/yourusername/book-of-mormon-coin/discussions)
4. Create an issue on GitHub 