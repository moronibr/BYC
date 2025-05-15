# Brigham Young Chain RPC API Documentation

This document describes the JSON-RPC 2.0 API endpoints available in the Brigham Young Chain (BYC) node.

## Base URL

- HTTP: `http://localhost:8332/rpc`
- WebSocket: `ws://localhost:8332/ws`

## Authentication

Currently, the API does not require authentication. In production environments, it is recommended to run the node behind a reverse proxy with proper authentication.

## JSON-RPC 2.0 Format

All requests must follow the JSON-RPC 2.0 specification:

```json
{
    "jsonrpc": "2.0",
    "method": "method_name",
    "params": {},
    "id": 1
}
```

Responses will be in the format:

```json
{
    "jsonrpc": "2.0",
    "result": {},
    "id": 1
}
```

Or in case of an error:

```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -32601,
        "message": "Method not found",
        "data": null
    },
    "id": 1
}
```

## Error Codes

- -32700: Parse error
- -32600: Invalid Request
- -32601: Method not found
- -32602: Invalid params
- -32603: Internal error

## Blockchain Methods

### getblockcount

Returns the current block count.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "getblockcount",
    "params": {},
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": 1234,
    "id": 1
}
```

### getbestblockhash

Returns the hash of the best (tip) block in the most-work fully-validated chain.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "getbestblockhash",
    "params": {},
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": "0000000000000000000000000000000000000000000000000000000000000000",
    "id": 1
}
```

### getblock

Returns information about a block by its hash.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "getblock",
    "params": {
        "hash": "0000000000000000000000000000000000000000000000000000000000000000"
    },
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "header": {
            "version": 1,
            "prev_hash": "0000000000000000000000000000000000000000000000000000000000000000",
            "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
            "timestamp": 1234567890,
            "bits": 486604799,
            "nonce": 0,
            "hash": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        "transactions": []
    },
    "id": 1
}
```

## Transaction Methods

### gettransaction

Returns information about a transaction by its hash.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "gettransaction",
    "params": {
        "hash": "0000000000000000000000000000000000000000000000000000000000000000"
    },
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "version": 1,
        "timestamp": 1234567890,
        "inputs": [],
        "outputs": [],
        "fee": 0,
        "hash": "0000000000000000000000000000000000000000000000000000000000000000",
        "coin_type": "golden"
    },
    "id": 1
}
```

### sendtransaction

Sends a transaction to the network.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "sendtransaction",
    "params": {
        "transaction": {
            "version": 1,
            "timestamp": 1234567890,
            "inputs": [],
            "outputs": [],
            "fee": 0,
            "coin_type": "golden"
        }
    },
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": "0000000000000000000000000000000000000000000000000000000000000000",
    "id": 1
}
```

## Wallet Methods

### getbalance

Returns the balance of a wallet address.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "getbalance",
    "params": {
        "address": "BYC1..."
    },
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "golden": 1000000,
        "silver": 500000,
        "shiblum": 100000,
        "shiblon": 50000,
        "antion": 10000
    },
    "id": 1
}
```

### createwallet

Creates a new wallet.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "createwallet",
    "params": {
        "coin_type": "golden"
    },
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "address": "BYC1...",
        "public_key": "...",
        "coin_type": "golden"
    },
    "id": 1
}
```

### loadwallet

Loads a wallet from a file.

**Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "loadwallet",
    "params": {
        "filename": "wallet.dat"
    },
    "id": 1
}
```

**Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "address": "BYC1...",
        "public_key": "...",
        "coin_type": "golden"
    },
    "id": 1
}
```

## WebSocket Events

The WebSocket connection supports all the same methods as the HTTP API, plus the following events:

### block

Emitted when a new block is added to the chain.

```json
{
    "jsonrpc": "2.0",
    "method": "block",
    "params": {
        "header": {
            "version": 1,
            "prev_hash": "0000000000000000000000000000000000000000000000000000000000000000",
            "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
            "timestamp": 1234567890,
            "bits": 486604799,
            "nonce": 0,
            "hash": "0000000000000000000000000000000000000000000000000000000000000000"
        },
        "transactions": []
    }
}
```

### transaction

Emitted when a new transaction is received.

```json
{
    "jsonrpc": "2.0",
    "method": "transaction",
    "params": {
        "version": 1,
        "timestamp": 1234567890,
        "inputs": [],
        "outputs": [],
        "fee": 0,
        "hash": "0000000000000000000000000000000000000000000000000000000000000000",
        "coin_type": "golden"
    }
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse. The default limits are:

- 100 requests per minute for HTTP
- 1000 requests per minute for WebSocket

## Best Practices

1. Always handle errors appropriately
2. Use WebSocket for real-time updates
3. Implement proper error handling and retry logic
4. Cache responses when appropriate
5. Use batch requests for multiple operations
6. Implement proper security measures in production

## Examples

### Node.js Example

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8332/ws');

ws.on('open', () => {
    ws.send(JSON.stringify({
        jsonrpc: '2.0',
        method: 'getblockcount',
        params: {},
        id: 1
    }));
});

ws.on('message', (data) => {
    console.log(JSON.parse(data));
});
```

### Python Example

```python
import websocket
import json

def on_message(ws, message):
    print(json.loads(message))

def on_error(ws, error):
    print(error)

def on_close(ws):
    print("### closed ###")

def on_open(ws):
    ws.send(json.dumps({
        "jsonrpc": "2.0",
        "method": "getblockcount",
        "params": {},
        "id": 1
    }))

websocket.enableTrace(True)
ws = websocket.WebSocketApp("ws://localhost:8332/ws",
                          on_message = on_message,
                          on_error = on_error,
                          on_close = on_close)
ws.on_open = on_open
ws.run_forever()
```

### Go Example

```go
package main

import (
    "fmt"
    "github.com/gorilla/websocket"
)

func main() {
    c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8332/ws", nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer c.Close()

    request := map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  "getblockcount",
        "params":  map[string]interface{}{},
        "id":      1,
    }

    if err := c.WriteJSON(request); err != nil {
        fmt.Println(err)
        return
    }

    var response map[string]interface{}
    if err := c.ReadJSON(&response); err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(response)
}
``` 