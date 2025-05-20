import '@testing-library/jest-dom';
import { configure } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { rest } from 'msw';

// Configure testing-library
configure({
  testIdAttribute: 'data-testid',
});

// Setup MSW for API mocking
export const server = setupServer(
  // Node endpoints
  rest.get('/api/v1/node/status', (req, res, ctx) => {
    return res(
      ctx.json({
        status: 'online',
        lastBlock: 1000,
        lastBlockTime: Date.now(),
        syncProgress: 100,
      })
    );
  }),

  // Wallet endpoints
  rest.get('/api/v1/wallet/:address', (req, res, ctx) => {
    return res(
      ctx.json({
        address: req.params.address,
        balance: 1000,
        nonce: 1,
        createdAt: Date.now(),
        lastUsed: Date.now(),
      })
    );
  }),

  // Network endpoints
  rest.get('/api/v1/network/peers', (req, res, ctx) => {
    return res(
      ctx.json([
        {
          address: '127.0.0.1:8545',
          version: '1.0.0',
          lastSeen: Date.now(),
          height: 1000,
          latency: 50,
        },
      ])
    );
  }),
);

// Start server before all tests
beforeAll(() => server.listen());

// Reset handlers after each test
afterEach(() => server.resetHandlers());

// Close server after all tests
afterAll(() => server.close()); 