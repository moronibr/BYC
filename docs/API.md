# BYC Blockchain API Documentation

## Overview

The BYC Blockchain API provides a RESTful interface for interacting with the blockchain network. All endpoints return JSON responses and use standard HTTP status codes.

## Base URL

```
http://localhost:8000
```

## Authentication

Currently, the API does not require authentication. Future versions will implement API key authentication.

## Endpoints

### Blockchain Operations

#### Get Block Information

```http
GET /block/{hash}
```

Returns information about a specific block.

**Parameters:**
- `hash` (path parameter): The hash of the block to retrieve

**Response:**
```json
{
    "timestamp": 1234567890,
    "transactions": [...],
    "prevHash": "0x...",
    "hash": "0x...",
    "nonce": 123,
    "blockType": "GOLDEN",
    "difficulty": 1
}
```

#### Get Transaction Information

```http
GET /transaction/{id}
```

Returns information about a specific transaction.

**Parameters:**
- `id` (path parameter): The ID of the transaction to retrieve

**Response:**
```json
{
    "id": "0x...",
    "inputs": [...],
    "outputs": [...],
    "timestamp": "2024-03-20T12:00:00Z",
    "blockType": "GOLDEN"
}
```

### Wallet Operations

#### Get Balance

```http
GET /balance/{address}/{coinType}
```

Returns the balance for a specific coin type.

**Parameters:**
- `address` (path parameter): The wallet address
- `coinType` (path parameter): The type of coin (e.g., "LEAH", "SENUM")

**Response:**
```json
{
    "address": "0x...",
    "coinType": "LEAH",
    "balance": 100.0
}
```

#### Get All Balances

```http
GET /balances/{address}
```

Returns balances for all coin types.

**Parameters:**
- `address` (path parameter): The wallet address

**Response:**
```json
{
    "address": "0x...",
    "balances": {
        "LEAH": 100.0,
        "SENUM": 50.0,
        ...
    }
}
```

#### Create Transaction

```http
POST /transaction
```

Creates a new transaction.

**Request Body:**
```json
{
    "from": "0x...",
    "to": "0x...",
    "amount": 10.0,
    "coinType": "LEAH"
}
```

**Response:**
```json
{
    "id": "0x...",
    "status": "pending"
}
```

### Special Coin Operations

#### Create Ephraim Coin

```http
POST /special/ephraim
```

Creates a new Ephraim coin.

**Request Body:**
```json
{
    "address": "0x..."
}
```

**Response:**
```json
{
    "success": true,
    "transactionId": "0x..."
}
```

#### Create Manasseh Coin

```http
POST /special/manasseh
```

Creates a new Manasseh coin.

**Request Body:**
```json
{
    "address": "0x..."
}
```

**Response:**
```json
{
    "success": true,
    "transactionId": "0x..."
}
```

#### Create Joseph Coin

```http
POST /special/joseph
```

Creates a new Joseph coin.

**Request Body:**
```json
{
    "address": "0x..."
}
```

**Response:**
```json
{
    "success": true,
    "transactionId": "0x..."
}
```

### Network Operations

#### Get Node Information

```http
GET /node/info
```

Returns information about the current node.

**Response:**
```json
{
    "address": "0x...",
    "peers": [...],
    "blockType": "GOLDEN",
    "isMining": true,
    "miningCoin": "LEAH"
}
```

#### Get Network Status

```http
GET /network/status
```

Returns the current network status.

**Response:**
```json
{
    "totalNodes": 10,
    "connectedPeers": 5,
    "blockHeight": 1000,
    "networkDifficulty": 1
}
```

## Error Responses

All endpoints return standard HTTP status codes and error messages in the following format:

```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message"
    }
}
```

Common error codes:
- `INVALID_PARAMETER`: Invalid parameter provided
- `INSUFFICIENT_BALANCE`: Not enough balance for transaction
- `INVALID_TRANSACTION`: Invalid transaction data
- `NETWORK_ERROR`: Network-related error
- `INTERNAL_ERROR`: Internal server error

## Rate Limiting

API requests are limited to:
- 100 requests per minute per IP address
- 1000 requests per hour per IP address

Rate limit headers are included in the response:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1234567890
```

## WebSocket API

The API also provides WebSocket endpoints for real-time updates:

```
ws://localhost:8000/ws
```

### WebSocket Events

1. **New Block**
```json
{
    "type": "new_block",
    "data": {
        "hash": "0x...",
        "blockType": "GOLDEN"
    }
}
```

2. **New Transaction**
```json
{
    "type": "new_transaction",
    "data": {
        "id": "0x...",
        "status": "pending"
    }
}
```

3. **Balance Update**
```json
{
    "type": "balance_update",
    "data": {
        "address": "0x...",
        "coinType": "LEAH",
        "balance": 100.0
    }
}
```

## Examples

### Creating a Transaction

```bash
curl -X POST http://localhost:8000/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0x123...",
    "to": "0x456...",
    "amount": 10.0,
    "coinType": "LEAH"
  }'
```

### Getting Balance

```bash
curl http://localhost:8000/balance/0x123.../LEAH
```

### Creating Special Coin

```bash
curl -X POST http://localhost:8000/special/ephraim \
  -H "Content-Type: application/json" \
  -d '{
    "address": "0x123..."
  }'
```

## SDK Support

The API is compatible with the following SDKs:
- Go SDK
- JavaScript SDK
- Python SDK

Please refer to the respective SDK documentation for implementation details. 