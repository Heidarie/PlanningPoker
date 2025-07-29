# Planning Poker - Secure Remote Estimation Tool

A beautiful, secure planning poker application with real-time collaboration for agile teams.

## 🚀 Features

- ✨ **Beautiful TUI Interface** - Modern terminal UI with smooth navigation and styling
- 🔐 **Secure Authentication** - Secret key authentication prevents unauthorized access
- 🛡️ **DDoS Protection** - Rate limiting and IP-based protection
- 🏠 **Host Controls** - Only hosts can start new voting sessions and control the game
- 👥 **Real-time Participant Tracking** - See who's joined and their voting status
- 🔒 **Vote Locking** - Prevents vote changes after submission for fair estimation
- 🧹 **Auto Cleanup** - Inactive rooms are automatically cleaned up after 10 minutes
- 🌊 **Cloud Hosted** - Deployed on DigitalOcean for global accessibility

## 📥 Download

Download the latest release from the [Releases](../../releases) section.

### Windows
```powershell
# Download the exe file and run directly
./cli_planning_poker_secure.exe
```

### macOS/Linux
```bash
# Download and make executable
chmod +x cli_planning_poker_secure
./cli_planning_poker_secure
```

## 🎮 How to Use

1. **Run the application** - Execute the downloaded binary
2. **Choose your role**:
   - **Create a new room** - Become the host and get a 4-character room code
   - **Join existing room** - Enter the room code and your name
3. **Start estimation** - Host can start voting sessions for story points
4. **Vote** - Participants select their estimates (1, 2, 3, 5, 8, 13, 21)
5. **View results** - Once everyone votes, results are revealed
6. **New round** - Host can start a new voting session

## 🔐 Security Features

- **Authentication**: All clients must have a valid secret key
- **Rate Limiting**: 1 request per second per IP address
- **DDoS Protection**: Automatic IP blocking for excessive requests
- **Room Validation**: Cannot join non-existent rooms
- **Auto Cleanup**: Inactive rooms are cleaned up automatically

## 🏗️ For Developers

### Quick Setup

1. **Clone the repository**:
```bash
git clone https://github.com/Heidarie/PlanningPoker.git
cd PlanningPoker
```

2. **Setup development environment**:
```bash
make setup-dev
```

3. **Configure your environment**:
Edit the `.env` file that was created:
```env
SERVER_URL=http://localhost:8080
CLIENT_SECRET=your-dev-secret-key
DEV_MODE=true
```

4. **Run in development**:
```bash
# Terminal 1: Start server
make run-server

# Terminal 2: Start client
make run-client
```

### Building for Production

1. **Set production environment**:
```bash
# Create production .env
SERVER_URL=https://your-app.vercel.app CLIENT_SECRET=your-production-secret make build-secure
```

2. **Build for all platforms**:
```bash
SERVER_URL=https://your-app.vercel.app CLIENT_SECRET=your-production-secret make build-all
```

### Deployment to DigitalOcean

1. **Set up DigitalOcean droplet**:
```bash
# Create Ubuntu 22.04 LTS droplet
# Configure firewall and SSH access
```

2. **Set environment variables in GitHub Secrets**:
```bash
# Required secrets:
DIGITALOCEAN_ACCESS_TOKEN=your_do_api_token
DIGITALOCEAN_DROPLET_IP=your_droplet_ip
DIGITALOCEAN_SERVER_URL=https://your-domain.com
CLIENT_SECRET=your_production_secret
SSH_PRIVATE_KEY=your_ssh_private_key
```

3. **Configure automated deployment**:
   - Go to your repository settings → Secrets and variables → Actions
   - Add all required secrets listed above
   
   **For DigitalOcean Auto-Deployment:**
   - Get API token from DigitalOcean Console → API → Tokens
   - Add your droplet's IP address
   - Add your SSH private key for secure deployment

4. **Deploy and build automatically**:
```bash
# Push a tag to trigger automatic DigitalOcean deployment + GitHub releases
git tag v1.0.0
git push origin v1.0.0
```

This will:
- Deploy the server to your DigitalOcean droplet with production settings
- Build secure client binaries for all platforms
- Create a GitHub release with pre-configured binaries

**For detailed setup instructions, see [DIGITALOCEAN_SETUP.md](DIGITALOCEAN_SETUP.md)**

### Building from Source

```bash
# Build client
go build -o planning_poker.exe ./cmd/client

# Build server (for local hosting)
go build -o server.exe ./cmd/server
```

## 🛠️ Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
# Server Configuration
SERVER_URL=https://your-app.vercel.app
# For local development:
# SERVER_URL=http://localhost:8080

# Security (REQUIRED)
CLIENT_SECRET=your-secure-secret-key-here

# Development Settings
DEV_MODE=false
```

### Configuration Priority

1. **Build-time variables** (highest priority) - Set via `-ldflags` during build
2. **Environment variables** - From `.env` file or system environment
3. **Default values** - Fallback defaults

### Available Make Commands

```bash
# Development
make setup-dev          # Setup development environment
make run-server         # Run development server
make run-client         # Run development client
make build-dev          # Build development version

# Production
make build-secure       # Build with embedded config
make build-all          # Build for all platforms

# Legacy
make build              # Legacy build without env vars
make clean              # Clean build artifacts
```

## 🎯 Game Rules

1. **Host Privileges**:
   - Create rooms
   - Start new voting sessions
   - Reset votes
   - Cannot vote (facilitator role)

2. **Participant Rights**:
   - Join rooms with valid codes
   - Vote once per session
   - Cannot change votes after submission
   - See voting progress

3. **Voting Process**:
   - Story points: 1, 2, 3, 5, 8, 13, 21
   - No vote changes after submission
   - Results shown when everyone votes
   - Host starts new rounds

## 🔒 Security Notes

- The secret key is embedded in the client binary
- Change the `CLIENT_SECRET` before building for production
- Rate limiting prevents API abuse
- Room codes are 4 characters (alphanumeric)
- WebSocket connections are secured with the same authentication

## 📝 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📞 Support

For issues, questions, or feature requests, please open an issue on GitHub.

---

Made with ❤️ for agile teams worldwide
