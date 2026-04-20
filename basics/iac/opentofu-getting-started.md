# Getting Started with OpenTofu: Mac & Docker Setup

> **Goal:** Go from zero to a working OpenTofu environment in 15 minutes.
> Two paths: native macOS install or Docker container. Pick one (or both).

---

## Which Path Should You Pick?

| | **Mac Native** | **Docker** |
|---|---|---|
| **Best for** | Daily driver, your main dev machine | CI/CD, team consistency, keeping your machine clean |
| **Setup time** | ~5 minutes | ~5 minutes |
| **Pros** | Faster execution, tab completion, direct filesystem access | Reproducible, isolated, no brew clutter, works on any OS |
| **Cons** | Ties to your machine, version managed by brew | Slight overhead, need to mount volumes for state/config |
| **Verdict** | **Start here** if you're learning | **Start here** if you're setting up a team pipeline |

---

## Path 1: Native macOS Install

### 1.1 Install OpenTofu via Homebrew

```bash
# Add the OpenTofu tap (one-time setup)
brew install opentofu

# Verify it works
tofu --version
# OpenTofu v1.9.x on darwin_arm64   (or darwin_amd64 for Intel Macs)
```

That's it. One command. Homebrew handles the binary, PATH, and updates.

> **Why not install manually?** You can — download from [opentofu.org/docs/intro/install](https://opentofu.org/docs/intro/install/) — but Homebrew gives you `brew upgrade opentofu` for updates, which is worth the 10 seconds of setup.

### 1.2 Install the AWS CLI

OpenTofu doesn't talk to AWS directly — it uses the **AWS provider plugin**, which relies on the same credential chain as the AWS CLI. So you need credentials configured.

```bash
brew install awscli

aws --version
# aws-cli/2.x.x Python/3.x.x Darwin/...
```

### 1.3 Configure AWS Credentials

You have three options. Pick the one that fits your situation:

#### Option A: `aws configure` (simplest for learning)

```bash
aws configure
# AWS Access Key ID [None]: AKIA...............
# AWS Secret Access Key [None]: wJalr.........
# Default region name [None]: us-east-1
# Default output format [None]: json
```

This writes credentials to `~/.aws/credentials` and config to `~/.aws/config`. OpenTofu reads these automatically.

**Where to get the keys:** AWS Console > IAM > Users > your user > Security credentials > Create access key.

#### Option B: Environment variables (good for CI, temporary sessions)

```bash
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="wJalr..."
export AWS_DEFAULT_REGION="us-east-1"

# Optional: for temporary credentials (SSO, assumed roles)
export AWS_SESSION_TOKEN="FwoGZX..."
```

#### Option C: AWS SSO (the way most companies do it)

```bash
# One-time setup
aws configure sso
# SSO session name: my-company
# SSO start URL: https://my-company.awsapps.com/start
# SSO region: us-east-1
# ... follow the browser prompts ...

# Daily login (tokens expire, usually every 8-12 hours)
aws sso login --profile my-profile

# Tell OpenTofu which profile to use
export AWS_PROFILE=my-profile
```

### 1.4 Verify Everything Works

```bash
# Check AWS credentials are valid
aws sts get-caller-identity
# {
#     "UserId": "AIDA...",
#     "Account": "123456789012",
#     "Arn": "arn:aws:iam::123456789012:user/your-name"
# }

# Create a test project
mkdir ~/tofu-test && cd ~/tofu-test

cat > main.tf << 'EOF'
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# A simple data source — reads info, creates nothing
data "aws_caller_identity" "current" {}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
EOF

tofu init       # downloads the AWS provider
tofu plan       # should show "no changes" (data sources are read-only)
tofu apply      # outputs your AWS account ID

# If you see your account ID, you're good to go!
```

### 1.5 Shell Completion (Optional but Nice)

```bash
# For zsh (default on modern macOS)
tofu -install-autocomplete

# Restart your shell or:
source ~/.zshrc

# Now try:
# tofu pl<tab>  →  tofu plan
# tofu ap<tab>  →  tofu apply
```

### 1.6 Editor Setup

**VS Code:**
```bash
# Install the HashiCorp HCL extension (works with OpenTofu)
code --install-extension hashicorp.HCL
```

**IntelliJ / GoLand / WebStorm:**
- Settings > Plugins > Marketplace > search "HCL" > Install "HashiCorp Terraform / HCL"

Both give you syntax highlighting, formatting, and autocomplete for `.tf` files.

---

## Path 2: Docker Setup

This approach runs OpenTofu inside a container. Your `.tf` files live on your Mac; the container mounts them and runs `tofu` commands.

### 2.1 Prerequisites

```bash
# Install Docker Desktop if you haven't
brew install --cask docker

# Start Docker Desktop (needs to be running)
open -a Docker

# Verify
docker --version
# Docker version 27.x.x
```

### 2.2 The OpenTofu Docker Image

The official image is `ghcr.io/opentofu/opentofu`. Pull it:

```bash
docker pull ghcr.io/opentofu/opentofu:latest

# Verify
docker run --rm ghcr.io/opentofu/opentofu:latest --version
# OpenTofu v1.9.x
```

### 2.3 Running OpenTofu Commands via Docker

The pattern is always the same — mount your project directory and pass AWS credentials:

```bash
# The base command (you'll alias this below)
docker run --rm \
  -v "$(pwd):/workspace" \       # mount your .tf files
  -w /workspace \                 # set working directory
  -e AWS_ACCESS_KEY_ID \          # pass through env vars (no value = use host's)
  -e AWS_SECRET_ACCESS_KEY \
  -e AWS_DEFAULT_REGION \
  -e AWS_SESSION_TOKEN \
  ghcr.io/opentofu/opentofu:latest \
  init                            # ← the tofu command goes here
```

### 2.4 Make It Usable: Shell Alias

Nobody wants to type that every time. Add this to your `~/.zshrc`:

```bash
# ~/.zshrc — add this at the end

tofu() {
  docker run --rm -it \
    -v "$(pwd):/workspace" \
    -w /workspace \
    -v "$HOME/.aws:/root/.aws:ro" \
    -e AWS_ACCESS_KEY_ID \
    -e AWS_SECRET_ACCESS_KEY \
    -e AWS_DEFAULT_REGION \
    -e AWS_SESSION_TOKEN \
    -e AWS_PROFILE \
    ghcr.io/opentofu/opentofu:latest \
    "$@"
}
```

```bash
source ~/.zshrc

# Now you can use it like a native install:
tofu --version
tofu init
tofu plan
tofu apply
```

> **What the alias does:**
> - `-v "$(pwd):/workspace"` — mounts your current directory so OpenTofu can see your `.tf` files
> - `-v "$HOME/.aws:/root/.aws:ro"` — mounts your AWS credentials read-only (so `aws configure` works from the host)
> - `-e AWS_*` — passes through any environment variable credentials
> - `-it` — interactive mode (needed for the `yes` prompt on `apply`)
> - `--rm` — auto-removes the container after each run (no cruft)

### 2.5 Docker Compose Wrapper (For Teams)

If you want a reproducible setup the whole team can use, add this to your project:

```yaml
# docker-compose.yml — in the root of your OpenTofu project

services:
  tofu:
    image: ghcr.io/opentofu/opentofu:1.9
    working_dir: /workspace
    volumes:
      - .:/workspace                    # your .tf files
      - ~/.aws:/root/.aws:ro            # AWS credentials (read-only)
      - tofu-plugins:/workspace/.terraform   # cache providers between runs
    environment:
      - AWS_ACCESS_KEY_ID
      - AWS_SECRET_ACCESS_KEY
      - AWS_DEFAULT_REGION
      - AWS_SESSION_TOKEN
      - AWS_PROFILE
    entrypoint: ["tofu"]

volumes:
  tofu-plugins:                         # named volume = providers survive container restarts
```

Usage:

```bash
# Init, plan, apply — just like native
docker compose run --rm tofu init
docker compose run --rm tofu plan
docker compose run --rm tofu apply

# The tofu-plugins volume caches the AWS provider (~300MB)
# so `init` is fast after the first run.
```

### 2.6 Verify the Docker Setup

```bash
mkdir ~/tofu-docker-test && cd ~/tofu-docker-test

cat > main.tf << 'EOF'
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

data "aws_caller_identity" "current" {}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
EOF

# If you set up the alias:
tofu init
tofu apply

# If you're using docker-compose:
docker compose run --rm tofu init
docker compose run --rm tofu apply

# See your AWS account ID? You're set.
```

---

## Credential Security Checklist

Regardless of which path you chose, follow these rules:

| Rule | Why |
|---|---|
| **Never put credentials in `.tf` files** | They get committed to git. Someone will find them |
| **Add `*.tfstate` to `.gitignore`** | State files contain every resource attribute, including secrets |
| **Add `*.tfvars` to `.gitignore` if they contain secrets** | Better yet: don't put secrets in tfvars — use Secrets Manager or SSM Parameter Store |
| **Use IAM roles in CI/CD** | No long-lived keys. GitHub Actions has OIDC federation with AWS |
| **Rotate access keys periodically** | If you're using access keys, rotate every 90 days. Or better: use SSO |
| **Use least-privilege IAM policies** | Don't give your OpenTofu user `AdministratorAccess`. Scope to what it needs |

---

## Quick Reference: Commands You'll Run Daily

```bash
# ── Project lifecycle ──
tofu init                  # first time, or after changing providers
tofu plan                  # preview changes (ALWAYS run before apply)
tofu apply                 # execute changes
tofu destroy               # tear everything down

# ── Formatting & validation ──
tofu fmt                   # auto-format .tf files
tofu fmt -check            # CI-friendly: exits non-zero if unformatted
tofu validate              # syntax check without talking to AWS

# ── Inspect state ──
tofu show                  # show current state in human-readable format
tofu state list            # list all managed resources
tofu state show <address>  # details of one resource
tofu output                # show all outputs
tofu output <name>         # show one output

# ── Debugging ──
TF_LOG=DEBUG tofu plan     # verbose logging (also: TRACE, INFO, WARN, ERROR)
tofu plan -out=plan.bin    # save plan to file
tofu show plan.bin         # review a saved plan
tofu apply plan.bin        # apply a saved plan (no re-confirmation needed)
```

---

## Troubleshooting

| Problem | Fix |
|---|---|
| `Error: No valid credential sources found` | Run `aws sts get-caller-identity` — if that fails, your credentials aren't set up. Re-run `aws configure` or check your env vars |
| `Error: Failed to query available provider packages` | Network issue. Check internet/proxy. If behind corporate VPN, you may need a provider mirror |
| `tofu init` is slow | It's downloading the AWS provider (~300MB). Normal on first run. Cached after that |
| Docker: `permission denied` on mounted volume | Docker Desktop > Settings > File Sharing — ensure your project directory is shared |
| Docker: `Error: configuring Terraform AWS Provider: no valid credential sources` | AWS env vars aren't being passed through. Check your alias or docker-compose.yml has all the `-e AWS_*` flags |
| `Error: Error acquiring the state lock` | Someone else (or a crashed process) is holding the lock. Wait, or if you're sure it's stale: `tofu force-unlock <LOCK_ID>` |
| `terraform` block name? | Yes, the block is literally called `terraform {}` even in OpenTofu. Historical naming from the fork. It works fine |

---

## Next Step

You've got a working OpenTofu setup. Head to [opentofu-tutorial.md](opentofu-tutorial.md) for your first real infrastructure — S3 buckets and an Aurora database.
