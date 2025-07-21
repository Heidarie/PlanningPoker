# Vercel Setup Guide (Hobby Tier)

## 🚀 Setting up Automatic Vercel Deployment for Personal Accounts

### Step 1: Get Vercel Token
1. Go to [Vercel Account Settings](https://vercel.com/account/tokens)
2. Click "Create Token"
3. Name it "GitHub Actions Deploy"
4. Copy the token (save it for GitHub Secrets)

### Step 2: Configure GitHub Secrets
Go to your GitHub repository → Settings → Secrets and variables → Actions

Add these secrets:
```
VERCEL_TOKEN=your_vercel_token_here
CLIENT_SECRET=your_secure_secret_key_here
SERVER_URL=https://your-app.vercel.app (optional)
```

**Note**: For Vercel hobby tier, you only need the `VERCEL_TOKEN`. The system will automatically link to your project.

### Step 3: Test the Deployment
```bash
# Create and push a tag to trigger deployment
git tag v1.0.0
git push origin v1.0.0
```

This will:
1. ✅ Deploy server to Vercel automatically
2. ✅ Build client binaries with the correct server URL
3. ✅ Create a GitHub release with downloadable binaries
4. ✅ Everything pre-configured and ready to use!

## 🔧 Manual Vercel Setup (Alternative)

If you prefer to deploy manually first:

```bash
# Install Vercel CLI
npm i -g vercel

# Login to Vercel
vercel login

# Deploy the project
vercel --prod

# Set environment variables
vercel env add CLIENT_SECRET
```

## 🎯 Benefits of Automated Deployment

- **Zero-touch deployment**: Push a tag → everything deploys
- **Consistent builds**: Same environment every time
- **Pre-configured binaries**: Users get working apps immediately  
- **Version synchronization**: Client and server always match
- **Production security**: Environment variables properly configured

---

**Ready to go! 🚀** Once configured, your entire deployment process is a single `git push`!
