import React, { useEffect, useState } from 'react';
import { NodeStatus, NetworkStats, Transaction, SystemMetrics } from '../types';
import { useApi } from '../hooks/useApi';
import { Card, Grid, Typography, CircularProgress, Alert } from '@mui/material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';

interface DashboardProps {
  refreshInterval?: number;
}

export const Dashboard: React.FC<DashboardProps> = ({ refreshInterval = 5000 }) => {
  const api = useApi();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [nodeStatus, setNodeStatus] = useState<NodeStatus | null>(null);
  const [networkStats, setNetworkStats] = useState<NetworkStats | null>(null);
  const [recentTxs, setRecentTxs] = useState<Transaction[]>([]);
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const [status, stats, txs, systemMetrics] = await Promise.all([
          api.getNodeStatus(),
          api.getNetworkStats(),
          api.getRecentTransactions(),
          api.getSystemMetrics()
        ]);

        setNodeStatus(status);
        setNetworkStats(stats);
        setRecentTxs(txs);
        setMetrics(systemMetrics);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch dashboard data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, refreshInterval);

    return () => clearInterval(interval);
  }, [api, refreshInterval]);

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </div>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <div style={{ padding: '20px' }}>
      <Typography variant="h4" gutterBottom>
        Dashboard
      </Typography>

      <Grid container spacing={3}>
        {/* Node Status */}
        <Grid item xs={12} md={6}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Node Status
            </Typography>
            {nodeStatus && (
              <div>
                <Typography>Version: {nodeStatus.version}</Typography>
                <Typography>Network: {nodeStatus.network}</Typography>
                <Typography>Block Height: {nodeStatus.blockHeight}</Typography>
                <Typography>Peers: {nodeStatus.peers}</Typography>
                <Typography>Uptime: {nodeStatus.uptime}</Typography>
              </div>
            )}
          </Card>
        </Grid>

        {/* Network Stats */}
        <Grid item xs={12} md={6}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Network Statistics
            </Typography>
            {networkStats && (
              <div>
                <Typography>Total Nodes: {networkStats.totalNodes}</Typography>
                <Typography>Active Nodes: {networkStats.activeNodes}</Typography>
                <Typography>Average Block Time: {networkStats.avgBlockTime}s</Typography>
                <Typography>Network Hash Rate: {networkStats.hashRate} H/s</Typography>
              </div>
            )}
          </Card>
        </Grid>

        {/* System Metrics */}
        <Grid item xs={12}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              System Metrics
            </Typography>
            {metrics && (
              <LineChart
                width={800}
                height={300}
                data={metrics.history}
                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="cpu" stroke="#8884d8" />
                <Line type="monotone" dataKey="memory" stroke="#82ca9d" />
                <Line type="monotone" dataKey="network" stroke="#ffc658" />
              </LineChart>
            )}
          </Card>
        </Grid>

        {/* Recent Transactions */}
        <Grid item xs={12}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Recent Transactions
            </Typography>
            {recentTxs.map((tx) => (
              <div key={tx.hash} style={{ marginBottom: '10px' }}>
                <Typography>Hash: {tx.hash}</Typography>
                <Typography>From: {tx.from}</Typography>
                <Typography>To: {tx.to}</Typography>
                <Typography>Amount: {tx.amount} BMC</Typography>
                <Typography>Status: {tx.status}</Typography>
              </div>
            ))}
          </Card>
        </Grid>
      </Grid>
    </div>
  );
}; 