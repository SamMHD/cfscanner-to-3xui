# cfscanner-to-3xui

**Scan Cloudflare IPs → generate Xray outbounds → push to [3x-ui](https://github.com/MHSanaei/3x-ui).**

*Uses [Cloudflare-Clean-IP-Scanner](https://github.com/bia-pain-bache/Cloudflare-Clean-IP-Scanner) for IP testing and [3x-ui](https://github.com/MHSanaei/3x-ui) as the panel.*

---

## 🚀 Quick start

```bash
# One-shot: scan → generate → update
cfscanner-to-3xui run

# Or run in a loop every N minutes (default 60)
cfscanner-to-3xui run-cron
cfscanner-to-3xui run-cron -n 30
```

---

## 📋 Commands

| Command | Description |
|--------|--------------|
| `scan` | Run Cloudflare IP latency/speed test; writes `ip-scan-result.csv`. |
| `generate` | Read `ip-scan-result.csv` + JSON templates in `configs/` → write `generated-outbounds.json`. |
| `update` | Push `generated-outbounds.json` to 3x-ui panel (replace outbounds with tag prefix, restart Xray). |
| `run` | Run `scan` → `generate` → `update` once. |
| `run-cron` | Run `run` every N minutes (`-n` or `CRON_MINUTES`). |

---

## 🔧 Environment variables

### 3x-ui (required for `update` / `run` / `run-cron`)

| Variable | Description |
|----------|-------------|
| `XUI_URL` | Panel base URL (e.g. `https://panel.example.com`). |
| `XUI_USERNAME` | Login username. |
| `XUI_PASSWORD` | Login password. |
| `XUI_ALLOW_INSECURE` | `1` or `true` to skip TLS verify. |
| `OUTBOUND_PREFIX` | Tag prefix for generated outbounds (replaced on update). Default: `cf-clean-`. |

### Cron

| Variable | Description |
|----------|-------------|
| `CRON_MINUTES` | Interval in minutes for `run-cron` (default: `60`). |

### Scan (optional; see [CloudflareScanner](https://github.com/bia-pain-bache/Cloudflare-Clean-IP-Scanner))

| Variable | Default | Description |
|----------|---------|-------------|
| `SCAN_N` | `200` | Latency test threads. |
| `SCAN_T` | `4` | Latency test times per IP. |
| `SCAN_DN` | `10` | Number of IPs to download-test. |
| `SCAN_DT` | `10` | Max seconds per download test. |
| `SCAN_TP` | `443` | Test port. |
| `SCAN_URL` | (speed.cloudflare.com) | HTTPing/download test URL. |
| `SCAN_HTTPING` | `false` | Use HTTP latency test. |
| `SCAN_HTTPING_CODE` | - | Valid HTTP status (e.g. `200`). |
| `SCAN_CFCOLO` | - | Comma-separated colo codes (HTTPing). |
| `SCAN_TL` | `9999` | Max avg latency (ms). |
| `SCAN_TLL` | `0` | Min avg latency (ms). |
| `SCAN_TLR` | `1` | Max loss rate (0–1). |
| `SCAN_SL` | `0` | Min download speed (MB/s). |
| `SCAN_P` | `10` | Number of results to print (`0` = no print). |
| `SCAN_F` | `ip.txt` | IP range file path. |
| `SCAN_IP` | - | Inline IP ranges (comma-separated). |
| `SCAN_O` | `ip-scan-result.csv` | Output CSV path. |
| `SCAN_DD` | `false` | Disable download test. |
| `SCAN_ALLIP` | `false` | Test every IP in range (IPv4). |

---

## 📁 Config templates

Put Xray outbound JSON files in **`configs/`** (or **`/app/configs`** in Docker). Supported: **trojan**, **vless**. Each file is one outbound; `address` is set per scanned IP and tag becomes `{OUTBOUND_PREFIX}{protocol}-{ip}`.

Example `configs/trojan.json`:

```json
{
  "protocol": "trojan",
  "settings": {
    "servers": [{ "address": "0.0.0.0", "port": 443, "password": "your-pass" }]
  },
  "streamSettings": { ... }
}
```

---

## 📦 Releases & binaries

Prebuilt binaries for Linux (amd64, arm64, arm, riscv64, mips64, mips64le), Windows (amd64, arm64), macOS (amd64, arm64), and Android (arm64) are on [Releases](https://github.com/SamMHD/cfscanner-to-3xui/releases).

Download the zip for your platform, extract, and run:

```bash
./cfscanner-to-3xui run
# or
./cfscanner-to-3xui run-cron -n 60
```

---

## 🐳 Docker

Images are published on **GHCR** for **linux/amd64** and **linux/arm64**. The container runs **`run-cron`** by default.

```bash
# Pull (use your tag)
docker pull ghcr.io/sammhd/cfscanner-to-3xui:latest
```

**Example run:** mount configs and set 3x-ui env:

```bash
docker run -d --name cfscanner-to-3xui \
  -v /path/to/configs:/app/configs \
  -e XUI_URL=https://panel.example.com \
  -e XUI_USERNAME=admin \
  -e XUI_PASSWORD=your-password \
  -e CRON_MINUTES=60 \
  -e OUTBOUND_PREFIX=cf-clean- \
  ghcr.io/sammhd/cfscanner-to-3xui:latest
```

Optional: mount `ip.txt` and/or persist scan output:

```bash
-v /path/to/ip.txt:/app/ip.txt \
-v /path/to/data:/app
```

Dockerfile is in the repo; multi-arch build runs on release.

---

## 🔨 Build from source

```bash
git clone https://github.com/SamMHD/cfscanner-to-3xui.git
cd cfscanner-to-3xui
go build -ldflags "-s -w -X main.version=dev" -o cfscanner-to-3xui .
./cfscanner-to-3xui run
```

---

## 📄 License

Same as upstream dependencies (see [Cloudflare-Clean-IP-Scanner](https://github.com/bia-pain-bache/Cloudflare-Clean-IP-Scanner), [3x-ui](https://github.com/MHSanaei/3x-ui)).
