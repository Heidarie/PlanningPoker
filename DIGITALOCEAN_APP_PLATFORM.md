# DigitalOcean App Platform Setup Guide

## 🌊 Quick Setup for DigitalOcean App Platform

Since you're using **DigitalOcean App Platform** (not Droplets), the setup is much simpler! No SSH or server management needed.

## 📋 What You Need

1. **DigitalOcean Account**: [Sign up here](https://www.digitalocean.com/)
2. **GitHub Repository**: Your Planning Poker code (already done!)
3. **CLIENT_SECRET**: A secure secret key for authentication

## 🚀 Step-by-Step Setup

### Step 1: Create App on DigitalOcean

1. **Go to DigitalOcean Console**:
   - Visit [DigitalOcean Apps](https://cloud.digitalocean.com/apps)
   - Click **"Create App"**

2. **Choose Source**:
   - Select **"GitHub"**
   - Authorize DigitalOcean to access your GitHub
   - Choose your repository: `Heidarie/PlanningPoker`
   - Branch: `main`
   - Auto-deploy: ✅ **Enabled**

3. **Configure Build Settings**:
   - **Source Directory**: `/` (root)
   - **Build Command**: `go build -o bin/server ./cmd/server`
   - **Run Command**: `./bin/server`
   - **Port**: `8080`

### Step 2: Set Environment Variables

In the App Platform configuration, add these environment variables:

```
CLIENT_SECRET = your_secure_secret_key_here
PORT = 8080
```

**How to set them**:
1. In your app settings, go to **"Environment Variables"**
2. Click **"Add Variable"**
3. Add `CLIENT_SECRET` as a **Secret** (not plain text)
4. Add `PORT` as `8080`

### Step 3: Deploy

1. Click **"Create Resources"**
2. Wait for deployment (usually 2-3 minutes)
3. Your app will be available at: `https://your-app-name.ondigitalocean.app`

### Step 4: Configure GitHub Secrets

For automated client builds, set this in your GitHub repository secrets:

```
CLIENT_SECRET = same_secret_as_above
DIGITALOCEAN_APP_URL = https://your-app-name.ondigitalocean.app
```

**How to add GitHub secrets**:
1. Go to your GitHub repo → Settings → Secrets and variables → Actions
2. Click **"New repository secret"**
3. Add both secrets above

### Step 5: Build and Release Clients

Once your app is deployed:

1. **Create a release**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions will**:
   - Build client binaries for all platforms
   - Embed your App Platform URL in the clients
   - Create a GitHub release with downloadable binaries

## 🔧 App Platform Configuration

I've included a `.do/app.yaml` file in your repository that defines:

- **Build process**: Compiles your Go server
- **Runtime**: Sets up environment and port
- **Auto-deployment**: Deploys when you push to main branch
- **Resources**: Uses the smallest (cheapest) instance size

## 💰 Pricing

**App Platform Costs**:
- **Basic**: $5/month - Perfect for small teams (what we configured)
- **Professional**: $12/month - For larger usage
- **No surprise costs** - Fixed monthly pricing

## 🎯 Testing Your Deployment

### Test the Server
```bash
# Check if your server is running
curl https://your-app-name.ondigitalocean.app/health

# Should return: {"status":"OK","timestamp":"..."}
```

### Test Client Connection
1. Download a client binary from your GitHub release
2. Run it: `./planning_poker_secure_windows_amd64.exe`
3. Create a room - it should connect to your App Platform server

## 🔄 Making Updates

### Server Updates
Simply push to your `main` branch:
```bash
git add .
git commit -m "Update server"
git push origin main
```
App Platform will automatically redeploy!

### Client Updates (New Release)
```bash
git tag v1.0.1
git push origin v1.0.1
```
This creates new client binaries with the latest code.

## 🆘 Troubleshooting

### App Won't Start
1. Check **"Activity"** tab in App Platform console
2. Look for build errors
3. Verify environment variables are set

### Client Can't Connect
1. Verify `CLIENT_SECRET` matches between app and client
2. Check your app URL is correct
3. Try the health endpoint: `https://your-app.ondigitalocean.app/health`

### GitHub Actions Failing
1. Make sure `CLIENT_SECRET` is set in GitHub secrets
2. Verify `DIGITALOCEAN_APP_URL` points to your app

## 📋 Required GitHub Secrets Summary

```
CLIENT_SECRET = your_secure_secret_key
DIGITALOCEAN_APP_URL = https://your-app-name.ondigitalocean.app
```

That's it! No SSH keys, no server management, no complex deployment scripts. App Platform handles everything automatically! 🎉

## 🌟 Benefits of App Platform

- ✅ **Zero server management**
- ✅ **Automatic deployments**
- ✅ **Built-in load balancing**
- ✅ **SSL certificates included**
- ✅ **Monitoring and logs**
- ✅ **Easy scaling**
- ✅ **Fixed monthly cost**

Your Planning Poker server will be running on production-grade infrastructure with minimal setup! 🚀
