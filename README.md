# ☁️ CloudDrop — Seamless Local File Sharing via Flutter & Go

**CloudDrop** is an in-progress cross-platform file sharing application that bridges local and mobile environments using a QR-code-enabled tunnel system. Built with **Flutter** for the frontend (mobile + desktop) and **Go** for the backend, it enables secure, self-hosted, QR-scannable file access from your devices.

---
### 🚧 In Progres
---
## 🔧 Current Functionality

- 📱 **Flutter Mobile App**: Access files over the local network or internet via a secure tunnel.
- 🖥️ **Flutter Desktop App**: Initiates the backend Go server and sets up a Cloudflare tunnel to expose your file directory.
- 🌐 **Go Server**:
  - Serves files via HTTP.
  - Reads upload/download metadata.
  - Interfaces with tunnel and exposes API.
- 📸 **QR Code Generation**: For easy pairing between devices.
- 🔒 Local-first, no third-party storage — your files stay on your system.

---

## 📦 Architecture

```text
[Flutter Desktop] ── starts ──▶ [Go Backend Server]
        │                             │
        │                             ├── Serves files from selected directory
        │                             ├── Starts Cloudflare tunnel
        │                             └── Displays public URL as QR code
        ▼
[Flutter Mobile] ◀── scans ── QR code with tunnel URL
        │
        └── Access/download/upload files over tunnel via Go API
