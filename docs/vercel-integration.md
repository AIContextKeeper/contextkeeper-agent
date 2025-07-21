# Vercel Integration for Private Agent Distribution

## Setup for ContextKeeper.dev (NextJS)

### 1. Add Download API Routes

Create these files in `/Users/samu/Development/ContextKeeper/pages/api/`:

#### `/pages/api/download/[platform].js`
```javascript
import fs from 'fs';
import path from 'path';

export default async function handler(req, res) {
  const { platform } = req.query;
  
  // Track download analytics
  await trackDownload(platform, req);
  
  const binaryMap = {
    'macos-arm64': 'contextkeeper-agent-darwin-arm64',
    'macos-amd64': 'contextkeeper-agent-darwin-amd64', 
    'linux-amd64': 'contextkeeper-agent-linux-amd64',
    'windows-amd64': 'contextkeeper-agent-windows-amd64.exe'
  };
  
  const binaryName = binaryMap[platform];
  if (!binaryName) {
    return res.status(404).json({ error: 'Platform not supported' });
  }
  
  const filePath = path.join(process.cwd(), 'public', 'downloads', binaryName);
  
  if (!fs.existsSync(filePath)) {
    return res.status(404).json({ error: 'Binary not found' });
  }
  
  const stat = fs.statSync(filePath);
  
  res.setHeader('Content-Type', 'application/octet-stream');
  res.setHeader('Content-Disposition', `attachment; filename="${binaryName}"`);
  res.setHeader('Content-Length', stat.size);
  
  const stream = fs.createReadStream(filePath);
  stream.pipe(res);
}

async function trackDownload(platform, req) {
  // Add to your analytics
  console.log(`Download: ${platform} from ${req.headers['user-agent']}`);
  
  // Could send to your analytics service
  // await analytics.track('agent_downloaded', { platform, userAgent: req.headers['user-agent'] });
}
```

#### `/pages/api/download/latest.js`
```javascript
export default function handler(req, res) {
  // Auto-detect user's platform
  const userAgent = req.headers['user-agent'] || '';
  
  let platform = 'linux-amd64'; // default
  
  if (userAgent.includes('Mac')) {
    // Detect Apple Silicon vs Intel
    if (userAgent.includes('arm64') || userAgent.includes('Apple Silicon')) {
      platform = 'macos-arm64';
    } else {
      platform = 'macos-amd64';
    }
  } else if (userAgent.includes('Windows')) {
    platform = 'windows-amd64';
  } else if (userAgent.includes('Linux')) {
    platform = 'linux-amd64';
  }
  
  // Redirect to platform-specific download
  res.redirect(302, `/api/download/${platform}`);
}
```

### 2. Add Download Page

Create `/Users/samu/Development/ContextKeeper/pages/download.js`:

```javascript
import { useState, useEffect } from 'react';
import Head from 'next/head';

export default function Download() {
  const [detectedPlatform, setDetectedPlatform] = useState('');
  
  useEffect(() => {
    // Auto-detect platform
    const userAgent = navigator.userAgent;
    if (userAgent.indexOf('Mac') !== -1) {
      setDetectedPlatform('macOS');
    } else if (userAgent.indexOf('Windows') !== -1) {
      setDetectedPlatform('Windows');
    } else if (userAgent.indexOf('Linux') !== -1) {
      setDetectedPlatform('Linux');
    }
  }, []);
  
  const handleDownload = (platform) => {
    // Track download event
    if (typeof gtag !== 'undefined') {
      gtag('event', 'download', {
        'event_category': 'Agent',
        'event_label': platform
      });
    }
    
    // Trigger download
    window.location.href = `/api/download/${platform}`;
  };
  
  return (
    <>
      <Head>
        <title>Download ContextKeeper Agent</title>
        <meta name="description" content="Download the ContextKeeper agent for automatic AI session capture" />
      </Head>
      
      <div className="download-page">
        <div className="container">
          <h1>Download ContextKeeper Agent</h1>
          <p>Automatically capture and sync AI conversations across all your tools</p>
          
          {/* Primary download button */}
          <div className="primary-download">
            <button 
              onClick={() => handleDownload('latest')}
              className="btn-large btn-primary"
            >
              📦 Download for {detectedPlatform || 'Your Platform'}
            </button>
            <p className="download-note">
              Free: 50 sessions • Pro: Unlimited automation ($29/month)
            </p>
          </div>
          
          {/* All platforms */}
          <div className="all-platforms">
            <h3>All Platforms</h3>
            <div className="platform-grid">
              <button onClick={() => handleDownload('macos-arm64')}>
                🍎 macOS (Apple Silicon)
              </button>
              <button onClick={() => handleDownload('macos-amd64')}>
                🍎 macOS (Intel)
              </button>
              <button onClick={() => handleDownload('linux-amd64')}>
                🐧 Linux (x64)
              </button>
              <button onClick={() => handleDownload('windows-amd64')}>
                🪟 Windows (x64)
              </button>
            </div>
          </div>
          
          {/* Install instructions */}
          <div className="install-section">
            <h3>Quick Install</h3>
            <div className="install-tabs">
              <div className="tab-content">
                <h4>macOS / Linux</h4>
                <pre><code>curl -fsSL https://contextkeeper.dev/install.sh | bash</code></pre>
              </div>
              <div className="tab-content">
                <h4>Manual Install</h4>
                <ol>
                  <li>Download the binary for your platform</li>
                  <li>Make it executable: <code>chmod +x contextkeeper-agent</code></li>
                  <li>Move to PATH: <code>mv contextkeeper-agent /usr/local/bin/</code></li>
                  <li>Run: <code>contextkeeper-agent</code></li>
                </ol>
              </div>
            </div>
          </div>
          
          {/* Features */}
          <div className="features">
            <h3>What you get:</h3>
            <div className="feature-grid">
              <div className="feature">
                <h4>🤖 Auto-Detection</h4>
                <p>Automatically captures Claude, Gemini, ChatGPT sessions</p>
              </div>
              <div className="feature">
                <h4>🔄 Real-time Sync</h4>
                <p>Sessions appear instantly in your web dashboard</p>
              </div>
              <div className="feature">
                <h4>🛡️ Privacy First</h4>
                <p>Your code never leaves your machine</p>
              </div>
              <div className="feature">
                <h4>⚡ Lightweight</h4>
                <p>Minimal resource usage, runs in background</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
```

### 3. Update Install Script

Create `/Users/samu/Development/ContextKeeper/public/install.sh`:

```bash
#!/bin/bash
# Installation script hosted on contextkeeper.dev

REPO_BASE="https://contextkeeper.dev/api/download"

# Auto-detect and download
curl -fsSL "$REPO_BASE/latest" -o /tmp/contextkeeper-agent
chmod +x /tmp/contextkeeper-agent

# Install
if [ "$EUID" -eq 0 ]; then
    mv /tmp/contextkeeper-agent /usr/local/bin/
else
    mkdir -p "$HOME/.local/bin"
    mv /tmp/contextkeeper-agent "$HOME/.local/bin/"
fi

echo "✅ ContextKeeper Agent installed!"
echo "Run: contextkeeper-agent"
```

## 3. Homepage Integration

Update the hero section to point to your download page:

```html
<a href="/download" class="btn-primary">
  📦 Download Agent
</a>
```

## 4. Build Process

1. Build binaries: `./scripts/build-for-vercel.sh`
2. Commit to ContextKeeper repo
3. Deploy to Vercel
4. Binaries available at `contextkeeper.dev/downloads/`

## 5. Analytics Integration

Track downloads in your existing analytics:
- Download button clicks
- Platform distribution
- Download completion rates
- Install script usage

This gives you full control over distribution while keeping the repos private!