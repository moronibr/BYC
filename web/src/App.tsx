import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import {
  AppBar,
  Toolbar,
  Typography,
  Container,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  IconButton,
  useTheme,
  useMediaQuery
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  AccountBalance as WalletIcon,
  NetworkCheck as NetworkIcon,
  Menu as MenuIcon
} from '@mui/icons-material';
import { Dashboard } from './components/Dashboard';
import { Wallet } from './components/Wallet';
import { Network } from './components/Network';

const drawerWidth = 240;

export const App: React.FC = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const drawer = (
    <div>
      <Toolbar>
        <Typography variant="h6" noWrap>
          BMC Node
        </Typography>
      </Toolbar>
      <List>
        <ListItem button component={Link} to="/">
          <ListItemIcon>
            <DashboardIcon />
          </ListItemIcon>
          <ListItemText primary="Dashboard" />
        </ListItem>
        <ListItem button component={Link} to="/wallet">
          <ListItemIcon>
            <WalletIcon />
          </ListItemIcon>
          <ListItemText primary="Wallet" />
        </ListItem>
        <ListItem button component={Link} to="/network">
          <ListItemIcon>
            <NetworkIcon />
          </ListItemIcon>
          <ListItemText primary="Network" />
        </ListItem>
      </List>
    </div>
  );

  return (
    <Router>
      <div style={{ display: 'flex' }}>
        <AppBar
          position="fixed"
          style={{
            width: isMobile ? '100%' : `calc(100% - ${drawerWidth}px)`,
            marginLeft: isMobile ? 0 : drawerWidth,
          }}
        >
          <Toolbar>
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              style={{ marginRight: 2, display: isMobile ? 'block' : 'none' }}
            >
              <MenuIcon />
            </IconButton>
            <Typography variant="h6" noWrap>
              Book of Mormon Coin
            </Typography>
          </Toolbar>
        </AppBar>

        <Drawer
          variant={isMobile ? 'temporary' : 'permanent'}
          open={isMobile ? mobileOpen : true}
          onClose={handleDrawerToggle}
          style={{
            width: drawerWidth,
            flexShrink: 0,
          }}
          ModalProps={{
            keepMounted: true, // Better open performance on mobile.
          }}
        >
          {drawer}
        </Drawer>

        <Container
          style={{
            marginTop: 64,
            padding: theme.spacing(3),
            width: isMobile ? '100%' : `calc(100% - ${drawerWidth}px)`,
            marginLeft: isMobile ? 0 : drawerWidth,
          }}
        >
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/wallet" element={<Wallet address="your-wallet-address" />} />
            <Route path="/network" element={<Network />} />
          </Routes>
        </Container>
      </div>
    </Router>
  );
}; 