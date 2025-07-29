# Migration from Vercel to DigitalOcean

## 🔄 Migration Overview

This guide will help you migrate your Planning Poker deployment from Vercel to DigitalOcean for better control, performance, and cost-effectiveness.

## 📋 Why Migrate to DigitalOcean?

### Benefits of DigitalOcean:
- **Better WebSocket Support**: Native WebSocket handling without serverless limitations
- **Persistent Connections**: No cold starts or timeout issues
- **Full Server Control**: Complete control over server configuration
- **Better Performance**: Dedicated resources vs shared serverless
- **Cost Predictability**: Fixed monthly cost vs per-request pricing
- **Custom Domain**: Easy SSL setup with Let's Encrypt
- **Monitoring**: Built-in monitoring and alerting

### Vercel Limitations:
- 10-second execution limit for serverless functions
- WebSocket connections can be unstable
- Cold starts affect user experience
- Limited control over server configuration

## 🚀 Migration Steps

### Step 1: Backup Current Configuration

Before migrating, document your current setup:

```bash
# Note your current settings
echo "Current SERVER_URL: $(echo $SERVER_URL)"
echo "Current CLIENT_SECRET: $(echo $CLIENT_SECRET | cut -c1-8)..."

# Backup your GitHub secrets
# - SERVER_URL or VERCEL_PROJECT_URL
# - CLIENT_SECRET  
# - VERCEL_TOKEN (can be removed after migration)
```

### Step 2: Set Up DigitalOcean

Follow the [DigitalOcean Setup Guide](DIGITALOCEAN_SETUP.md) to:

1. **Create DigitalOcean account and droplet**
2. **Configure GitHub secrets for DigitalOcean**
3. **Set up domain and SSL** (recommended)

### Step 3: Update GitHub Secrets

Replace Vercel secrets with DigitalOcean secrets:

#### Remove (after migration):
```
VERCEL_TOKEN
VERCEL_PROJECT_ID
SERVER_URL (if using old name)
```

#### Add for DigitalOcean:
```
DIGITALOCEAN_ACCESS_TOKEN=your_do_api_token
DIGITALOCEAN_DROPLET_IP=your_droplet_ip
DIGITALOCEAN_SERVER_URL=https://your-domain.com
CLIENT_SECRET=same_secret_as_before (keep this!)
SSH_PRIVATE_KEY=your_ssh_private_key
```

### Step 4: Test Deployment

1. **Deploy to DigitalOcean**:
   ```bash
   git tag v2.0.0-do
   git push origin v2.0.0-do
   ```

2. **Test the deployment**:
   - Check if server is running: `https://your-domain.com/health`
   - Download and test a client binary
   - Create a room and test WebSocket connections

### Step 5: Update Documentation

1. **Update README.md** (already done in this migration)
2. **Update any internal documentation**
3. **Notify your team** about the new server URL

### Step 6: DNS and Domain Setup

If you're switching domains or URLs:

1. **Update DNS records** to point to your DigitalOcean droplet
2. **Set up SSL certificate** with Let's Encrypt
3. **Test HTTPS access**

### Step 7: Retire Vercel Deployment

After confirming DigitalOcean works well:

1. **Remove Vercel deployment** (optional)
2. **Clean up GitHub secrets** (remove VERCEL_TOKEN, etc.)
3. **Archive old workflow files**

## 🔧 Configuration Changes

### Workflow Changes

The new workflow (`build.yml`) includes:
- DigitalOcean deployment via SSH
- Systemd service configuration
- Nginx reverse proxy setup
- Automatic SSL certificate management
- Cross-platform client builds

### Environment Variable Changes

| Variable | Vercel | DigitalOcean |
|----------|--------|--------------|
| Server URL | `SERVER_URL` | `DIGITALOCEAN_SERVER_URL` |
| Deployment Token | `VERCEL_TOKEN` | `DIGITALOCEAN_ACCESS_TOKEN` |
| Server Access | N/A | `SSH_PRIVATE_KEY` |
| Server IP | N/A | `DIGITALOCEAN_DROPLET_IP` |

### Client Configuration

The client now defaults to:
- Server URL: `https://your-app.domain.com` (instead of Vercel)
- Same authentication system
- Improved WebSocket stability

## 🔍 Testing Checklist

After migration, test these features:

- [ ] **Server Health**: `curl https://your-domain.com/health`
- [ ] **Room Creation**: Create a new room via client
- [ ] **WebSocket Connection**: Join room and test real-time updates
- [ ] **Authentication**: Verify client secret authentication works
- [ ] **Rate Limiting**: Test that rate limiting is working
- [ ] **SSL Certificate**: Verify HTTPS is working properly
- [ ] **Cross-Platform Clients**: Test Windows, Linux, and macOS binaries

## 🆘 Rollback Plan

If you need to rollback to Vercel:

1. **Revert GitHub secrets** to Vercel configuration
2. **Use legacy workflow**: Rename `deploy-vercel-legacy.yml` back to `build.yml`
3. **Update client defaults** back to Vercel URL
4. **Redeploy** with a new tag

## 📊 Performance Comparison

### Expected Improvements:
- **WebSocket Stability**: 99.9% vs 95% connection success
- **Latency**: 50-100ms improvement due to no cold starts
- **Reliability**: No 10-second timeout limits
- **Concurrent Users**: Better handling of multiple simultaneous users

### Monitoring:
```bash
# Monitor server performance on DigitalOcean
htop                    # CPU and memory usage
sudo systemctl status planning-poker  # Service status
sudo journalctl -u planning-poker -f  # Live logs
```

## 💰 Cost Comparison

### Vercel (Hobby):
- **Free tier**: Limited to hobby projects
- **Pro tier**: $20/month per member
- **Usage-based**: Can spike with heavy usage

### DigitalOcean:
- **Basic Droplet**: $6/month (fixed cost)
- **Standard Droplet**: $12/month (better performance)
- **Predictable costs**: No usage spikes

## 🎯 Next Steps After Migration

1. **Monitor Performance**: Keep an eye on server metrics for the first week
2. **Team Testing**: Have your team test the new deployment
3. **Documentation Updates**: Update any team documentation with new URLs
4. **SSL Renewal**: Set up automatic SSL certificate renewal
5. **Backup Strategy**: Implement regular backups of your droplet
6. **Scaling Plan**: Plan for horizontal scaling if needed

---

**Migration Complete! 🌊** Your Planning Poker server is now running on DigitalOcean with improved performance and reliability!
