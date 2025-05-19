import React, { useEffect, useState } from 'react';
import { useApi } from '../hooks/useApi';
import { Wallet as WalletType, Transaction } from '../types';
import {
  Card,
  Typography,
  Button,
  TextField,
  Grid,
  List,
  ListItem,
  ListItemText,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  CircularProgress
} from '@mui/material';

interface WalletProps {
  address: string;
}

export const Wallet: React.FC<WalletProps> = ({ address }) => {
  const api = useApi();
  const [wallet, setWallet] = useState<WalletType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sendDialogOpen, setSendDialogOpen] = useState(false);
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [sending, setSending] = useState(false);

  useEffect(() => {
    const fetchWallet = async () => {
      try {
        setLoading(true);
        const data = await api.getWallet(address);
        setWallet(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch wallet data');
      } finally {
        setLoading(false);
      }
    };

    fetchWallet();
    const interval = setInterval(fetchWallet, 5000);
    return () => clearInterval(interval);
  }, [api, address]);

  const handleSend = async () => {
    try {
      setSending(true);
      await api.sendTransaction({
        to: recipient,
        amount: parseFloat(amount),
      });
      setSendDialogOpen(false);
      setRecipient('');
      setAmount('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send transaction');
    } finally {
      setSending(false);
    }
  };

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

  if (!wallet) {
    return <Alert severity="warning">Wallet not found</Alert>;
  }

  return (
    <div style={{ padding: '20px' }}>
      <Typography variant="h4" gutterBottom>
        Wallet
      </Typography>

      <Grid container spacing={3}>
        {/* Wallet Info */}
        <Grid item xs={12} md={6}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Wallet Information
            </Typography>
            <Typography>Address: {wallet.address}</Typography>
            <Typography>Balance: {wallet.balance} BMC</Typography>
            <Button
              variant="contained"
              color="primary"
              onClick={() => setSendDialogOpen(true)}
              style={{ marginTop: '20px' }}
            >
              Send BMC
            </Button>
          </Card>
        </Grid>

        {/* Recent Transactions */}
        <Grid item xs={12} md={6}>
          <Card style={{ padding: '20px' }}>
            <Typography variant="h6" gutterBottom>
              Recent Transactions
            </Typography>
            <List>
              {wallet.transactions.map((tx: Transaction) => (
                <ListItem key={tx.hash}>
                  <ListItemText
                    primary={`${tx.amount} BMC ${tx.from === address ? 'to' : 'from'} ${tx.from === address ? tx.to : tx.from}`}
                    secondary={`Status: ${tx.status} | ${new Date(tx.timestamp).toLocaleString()}`}
                  />
                </ListItem>
              ))}
            </List>
          </Card>
        </Grid>
      </Grid>

      {/* Send Dialog */}
      <Dialog open={sendDialogOpen} onClose={() => setSendDialogOpen(false)}>
        <DialogTitle>Send BMC</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Recipient Address"
            type="text"
            fullWidth
            value={recipient}
            onChange={(e) => setRecipient(e.target.value)}
          />
          <TextField
            margin="dense"
            label="Amount"
            type="number"
            fullWidth
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSendDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleSend}
            color="primary"
            disabled={sending || !recipient || !amount}
          >
            {sending ? <CircularProgress size={24} /> : 'Send'}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}; 