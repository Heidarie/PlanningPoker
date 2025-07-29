# DigitalOcean Deployment Guide

## 🌊 Setting up Planning Poker on DigitalOcean

This guide will help you deploy the Planning Poker server on DigitalOcean and configure the automated deployment pipeline.

## 📋 Prerequisites

1. **DigitalOcean Account**: Sign up at [DigitalOcean](https://www.digitalocean.com/)
2. **Domain Name** (optional but recommended): For SSL/HTTPS setup
3. **GitHub Repository**: With proper secrets configured

## 🚀 Quick Setup

### Step 1: Create DigitalOcean Droplet

1. **Create a new Droplet**:
   - Go to [DigitalOcean Console](https://cloud.digitalocean.com/)
   - Click "Create" → "Droplets"
   - Choose **Ubuntu 22.04 LTS**
   - Select **Basic** plan ($6/month minimum recommended)
   - Choose a datacenter region close to your users
   - Add your SSH key for secure access
   - Give it a meaningful name like "planning-poker-server"

2. **Note your Droplet's IP address** - you'll need this for GitHub secrets

### Step 2: Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add these secrets:

```
DIGITALOCEAN_ACCESS_TOKEN=your_digitalocean_api_token
DIGITALOCEAN_DROPLET_IP=your_droplet_ip_address
DIGITALOCEAN_SERVER_URL=https://your-domain.com (or http://your-droplet-ip:8080)
CLIENT_SECRET=your_secure_secret_key_here
SSH_PRIVATE_KEY=your_ssh_private_key
```

#### Getting DigitalOcean API Token:
1. Go to [DigitalOcean API Tokens](https://cloud.digitalocean.com/account/api/tokens)
2. Click "Generate New Token"
3. Name it "GitHub Actions"
4. Select "Full Access"
5. Copy the token to GitHub secrets as `DIGITALOCEAN_ACCESS_TOKEN`

#### SSH Private Key:
```bash
# Generate SSH key pair if you don't have one
ssh-keygen -t rsa -b 4096 -C "github-actions@your-domain.com"

# Copy the private key content to GitHub secret SSH_PRIVATE_KEY
cat ~/.ssh/id_rsa

# Add the public key to your DigitalOcean droplet
cat ~/.ssh/id_rsa.pub
```

### Step 3: Initial Server Setup (One-time)

SSH into your droplet and run initial setup:

```bash
# Connect to your droplet
ssh root@your-droplet-ip

# Update system
apt update && apt upgrade -y

# Install required packages
apt install -y nginx certbot python3-certbot-nginx ufw

# Configure firewall
ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable

# Exit for now - the deployment script will handle the rest
exit
```

### Step 4: Deploy

1. **Trigger deployment**:
   ```bash
   # Create and push a tag to trigger deployment
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Or use manual deployment**:
   - Go to GitHub Actions tab
   - Click "Deploy to DigitalOcean"
   - Click "Run workflow"

## 🔧 Manual Deployment

If you prefer to deploy manually:

1. **Build the server**:
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o planning_poker_server ./cmd/server
   ```

2. **Copy to your server**:
   ```bash
   scp planning_poker_server root@your-droplet-ip:/tmp/
   ```

3. **SSH into your server and install**:
   ```bash
   ssh root@your-droplet-ip
   
   # Create user and directories
   useradd -r -s /bin/false planning-poker
   mkdir -p /opt/planning-poker
   chown planning-poker:planning-poker /opt/planning-poker
   
   # Copy binary
   cp /tmp/planning_poker_server /opt/planning-poker/
   chown planning-poker:planning-poker /opt/planning-poker/planning_poker_server
   chmod +x /opt/planning-poker/planning_poker_server
   
   # Create systemd service
   cat > /etc/systemd/system/planning-poker.service << EOF
   [Unit]
   Description=Planning Poker Server
   After=network.target
   
   [Service]
   Type=simple
   User=planning-poker
   WorkingDirectory=/opt/planning-poker
   ExecStart=/opt/planning-poker/planning_poker_server
   Restart=always
   RestartSec=5
   Environment=CLIENT_SECRET=your_secret_here
   Environment=PORT=8080
   
   [Install]
   WantedBy=multi-user.target
   EOF
   
   # Enable and start service
   systemctl daemon-reload
   systemctl enable planning-poker
   systemctl start planning-poker
   ```

## 🔒 SSL/HTTPS Setup (Recommended)

### Option 1: With Domain Name (Recommended)

1. **Point your domain to the droplet**:
   - Create an A record pointing to your droplet's IP
   - Wait for DNS propagation (can take up to 24 hours)

2. **Setup Let's Encrypt SSL**:
   ```bash
   # SSH into your server
   ssh root@your-droplet-ip
   
   # Configure nginx for your domain
   cat > /etc/nginx/sites-available/planning-poker << EOF
   server {
       listen 80;
       server_name your-domain.com;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_http_version 1.1;
           proxy_set_header Upgrade \$http_upgrade;
           proxy_set_header Connection "upgrade";
           proxy_set_header Host \$host;
           proxy_set_header X-Real-IP \$remote_addr;
           proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto \$scheme;
           proxy_cache_bypass \$http_upgrade;
           proxy_read_timeout 86400;
       }
   }
   EOF
   
   # Enable the site
   ln -s /etc/nginx/sites-available/planning-poker /etc/nginx/sites-enabled/
   nginx -t && systemctl reload nginx
   
   # Get SSL certificate
   certbot --nginx -d your-domain.com
   ```

### Option 2: Without Domain (IP only)

If you don't have a domain, you can access via IP:
- HTTP: `http://your-droplet-ip:8080`
- Note: WebSocket connections may have issues with some browsers over HTTP

## 📊 Monitoring and Maintenance

### Check Service Status
```bash
# Check if service is running
sudo systemctl status planning-poker

# View logs
sudo journalctl -u planning-poker -f

# Restart service
sudo systemctl restart planning-poker
```

### Update Deployment
Just push a new tag or trigger the GitHub Action again:
```bash
git tag v1.0.1
git push origin v1.0.1
```

### Backup
```bash
# Backup your configuration
tar -czf planning-poker-backup.tar.gz /opt/planning-poker /etc/systemd/system/planning-poker.service /etc/nginx/sites-available/planning-poker
```

## 🔧 Configuration

### Environment Variables

The server reads these environment variables:

- `CLIENT_SECRET`: Required secret key for authentication
- `PORT`: Server port (default: 8080)
- `DEV_MODE`: Enable development mode (default: false)

### Client Configuration

Update the `DIGITALOCEAN_SERVER_URL` secret in GitHub to point to your server:
- With domain: `https://your-domain.com`
- Without domain: `http://your-droplet-ip:8080`

## 🆘 Troubleshooting

### Common Issues

1. **Connection refused**:
   ```bash
   # Check if service is running
   sudo systemctl status planning-poker
   
   # Check firewall
   sudo ufw status
   ```

2. **WebSocket connection failed**:
   - Ensure nginx is configured properly
   - Check if SSL certificate is valid
   - Verify proxy settings

3. **Authentication errors**:
   - Verify CLIENT_SECRET matches between client and server
   - Check server logs for authentication attempts

### Logs
```bash
# Server logs
sudo journalctl -u planning-poker -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## 💰 Cost Estimation

### DigitalOcean Droplet Costs:
- **Basic ($6/month)**: 1 vCPU, 1GB RAM, 25GB SSD - Good for small teams
- **Standard ($12/month)**: 1 vCPU, 2GB RAM, 50GB SSD - Better for larger teams
- **Optimized ($24/month)**: 2 vCPU, 4GB RAM, 80GB SSD - High availability

### Additional Costs:
- **Domain**: $10-15/year (optional)
- **Load Balancer**: $12/month (for high availability)
- **Monitoring**: $5/month (optional)

## 🚀 Next Steps

1. Deploy your first version
2. Test with your team
3. Set up monitoring and alerts
4. Consider adding a load balancer for high availability
5. Set up automated backups

---

**Ready to deploy! 🌊** Your Planning Poker server will be running on DigitalOcean with professional-grade reliability.
