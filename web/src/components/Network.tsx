import React, { useEffect, useState } from 'react';
import { useApi } from '../hooks/useApi';
import { Peer, Health } from '../types';
import {
  Card,
  Typography,
  Grid,
  List,
  ListItem,
  ListItemText,
  Button,
  TextField,
  Alert,
  CircularProgress,
  Chip
} from '@mui/material';

export const Network: React.FC = () => {
  const api = useApi();
  const [peers, setPeers] = useState<Peer[]>([]);
  const [health, setHealth] = useState<Health | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newPeerAddress, setNewPeerAddress] = useState('');
  const [connecting, setConnecting] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const [peersData, healthData] = await Promise.all([
          api.getPeers(),
          api.getHealth()
        ]);
        setPeers(peersData);
        setHealth(healthData);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch network data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [api]);

  const handleConnect = async () => {
    try {
      setConnecting(true);
      await api.connectPeer(newPeerAddress);
      setNewPeerAddress('');
      const updatedPeers = await api.getPeers();
      setPeers(updatedPeers);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to connect to peer');
    } finally {
      setConnecting(false);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress data-testid="loading-spinner" />
      </div>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <div style={{ padding: '20px' }}>
      <Typography variant="h4" gutterBottom>
        Network
      </Typography>

      <Grid container spacing={3}>
        {/* Network Health */}
        <Grid item xs={12} md={6}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Network Health
            </Typography>
            {health && (
              <div>
                <Chip
                  label={health.status}
                  color={
                    health.status === 'healthy'
                      ? 'success'
                      : health.status === 'degraded'
                      ? 'warning'
                      : 'error'
                  }
                  style={{ marginBottom: '10px' }}
                />
                {health.issues.length > 0 && (
                  <List>
                    {health.issues.map((issue, index) => (
                      <ListItem key={index}>
                        <ListItemText primary={issue} />
                      </ListItem>
                    ))}
                  </List>
                )}
                <Typography variant="body2" color="textSecondary">
                  Last Check: {new Date(health.lastCheck).toLocaleString()}
                </Typography>
              </div>
            )}
          </Card>
        </Grid>

        {/* Connect to Peer */}
        <Grid item xs={12} md={6}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Connect to Peer
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs>
                <TextField
                  fullWidth
                  label="Peer Address"
                  value={newPeerAddress}
                  onChange={(e) => setNewPeerAddress(e.target.value)}
                  placeholder="host:port"
                />
              </Grid>
              <Grid item>
                <Button
                  variant="contained"
                  color="primary"
                  onClick={handleConnect}
                  disabled={connecting || !newPeerAddress}
                >
                  {connecting ? <CircularProgress size={24} /> : 'Connect'}
                </Button>
              </Grid>
            </Grid>
          </Card>
        </Grid>

        {/* Connected Peers */}
        <Grid item xs={12}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Connected Peers ({peers.length})
            </Typography>
            <List>
              {peers.map((peer) => (
                <ListItem key={peer.address}>
                  <ListItemText
                    primary={peer.address}
                    secondary={
                      <>
                        Version: {peer.version} | Latency: {peer.latency}ms |
                        Last Seen: {new Date(peer.lastSeen).toLocaleString()}
                      </>
                    }
                  />
                </ListItem>
              ))}
            </List>
          </Card>
        </Grid>
      </Grid>
    </div>
  );
}; 