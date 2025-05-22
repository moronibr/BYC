import React from 'react';
import { Box, Container, Typography, Paper } from '@mui/material';
import './Whitepaper.css';

const Whitepaper: React.FC = () => {
  return (
    <Container className="whitepaper-container">
      <Paper 
        elevation={3} 
        sx={{ 
          p: 4, 
          my: 4,
          backgroundColor: '#FFFFFF',
          border: '2px solid #00004d',
          borderRadius: '12px'
        }}
      >
        <Typography 
          variant="h2" 
          component="h1" 
          className="whitepaper-title"
          sx={{ 
            textShadow: '2px 2px 4px rgba(0, 0, 77, 0.2)',
            borderBottom: '3px solid #f9c209',
            paddingBottom: '1rem'
          }}
        >
          BYC: Book of Mormon Youth Coin
        </Typography>
        <Typography 
          variant="h4" 
          component="h2" 
          className="whitepaper-subtitle"
          sx={{ 
            marginTop: '1rem',
            textShadow: '1px 1px 2px rgba(249, 194, 9, 0.2)'
          }}
        >
          A Digital Currency Rooted in Faith and Innovation
        </Typography>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>Abstract</Typography>
          <Typography paragraph>
            BYC (Book of Mormon Youth Coin) is a revolutionary cryptocurrency that combines the principles of blockchain technology with the timeless values of the Book of Mormon. This whitepaper outlines the vision, technology, and unique value proposition of BYC, demonstrating how it serves as both a digital asset and a tool for strengthening faith and community.
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>1. Introduction</Typography>
          <Typography paragraph>
            BYC emerges from a vision to create a cryptocurrency that embodies the principles of faith, integrity, and community service. Inspired by the Book of Mormon and its teachings, BYC represents a new paradigm in digital currency - one that combines technological innovation with spiritual values.
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>2. Vision and Mission</Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li>Strengthens faith and community</li>
              <li>Promotes financial literacy and responsibility</li>
              <li>Supports charitable causes and community service</li>
              <li>Provides a secure and transparent financial system</li>
            </ul>
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>3. Technical Architecture</Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li>Secure transactions through advanced cryptography</li>
              <li>Transparent and immutable record-keeping</li>
              <li>Efficient consensus mechanisms</li>
              <li>Scalable network architecture</li>
            </ul>
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>4. Token Economics</Typography>
          <Typography paragraph>
            BYC implements a unique token system with special coins:
          </Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li><strong>Ephraim Coin:</strong> Representing leadership and responsibility</li>
              <li><strong>Manasseh Coin:</strong> Symbolizing unity and strength</li>
              <li><strong>Joseph Coin:</strong> Embodying wisdom and guidance</li>
            </ul>
          </Typography>
          <Typography paragraph>
            Each coin serves a specific purpose in the ecosystem while maintaining the core values of the platform.
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>5. Security and Trust</Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li>Advanced cryptographic algorithms</li>
              <li>Transparent transaction verification</li>
              <li>Regular security audits</li>
              <li>Community-driven governance</li>
            </ul>
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>6. Use Cases</Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li>Community donations and charitable giving</li>
              <li>Educational initiatives</li>
              <li>Youth programs and activities</li>
              <li>Digital asset management</li>
              <li>Cross-border transactions</li>
            </ul>
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>7. Community and Governance</Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li>Transparent decision-making processes</li>
              <li>Community voting mechanisms</li>
              <li>Regular updates and improvements</li>
              <li>Active community participation</li>
            </ul>
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>8. Future Development</Typography>
          <Typography component="div">
            <ul className="whitepaper-list">
              <li>Enhanced security features</li>
              <li>Expanded use cases</li>
              <li>Integration with more platforms</li>
              <li>Community growth initiatives</li>
            </ul>
          </Typography>
        </Box>

        <Box className="whitepaper-section">
          <Typography variant="h5" gutterBottom>9. Conclusion</Typography>
          <Typography paragraph>
            BYC represents a unique convergence of technology and faith, offering a secure, transparent, and value-driven digital currency. By combining blockchain innovation with spiritual principles, BYC creates a platform for financial transactions that strengthens communities and promotes positive values.
          </Typography>
        </Box>
      </Paper>
    </Container>
  );
};

export default Whitepaper; 