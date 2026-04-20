# OpenTofu: Infrastructure as Code from Zero to Production

> **Audience:** You can write code, you've deployed things, but you've never written IaC.
> **Cloud provider:** AWS (everything here uses the AWS provider).
> **Time to read:** ~30 minutes. Time to run every example: ~2 hours.

---

## 1. Why IaC Exists (The "I Clicked My Way to Production" Problem)

Think of it like this: you'd never deploy your application by SSHing into a server and typing commands from memory. You have a build pipeline, a Dockerfile, a `package.json`. Your **code** is reproducible. But your **infrastructure** — the VPC, the database, the S3 bucket, the IAM role — that's still a bunch of clicks in the AWS Console that only Dave remembers, and Dave is on vacation.

**The pain without IaC:**

| Problem | What happens |
|---|---|
| **No reproducibility** | "Works in dev" because someone hand-configured it differently in staging |
| **No audit trail** | Who added that public S3 bucket? When? Why? Nobody knows |
| **No code review** | Infra changes bypass PR review entirely |
| **No rollback** | You clicked "modify" and now the database is gone. Good luck |
| **Snowflake environments** | Dev, staging, and prod have drifted so far apart they're different planets |

**IaC fixes all of this.** You describe your infrastructure in files, commit those files to git, review them in PRs, and a tool reads those files and makes reality match what you wrote. If you delete a resource from the file and apply — the tool deletes it from AWS. If you change a property — the tool updates it. Your infrastructure is now **code**: versioned, reviewable, reproducible.

---

## 2. What Is OpenTofu?

**OpenTofu** is an open-source Infrastructure-as-Code tool that lets you define cloud resources (AWS, GCP, Azure, etc.) in declarative configuration files, then creates, updates, and destroys those resources to match your config. It's a fork of HashiCorp's Terraform, created in 2023 after Terraform switched from an open-source license to the Business Source License (BSL). OpenTofu is maintained by the Linux Foundation, uses the same HCL language, is compatible with nearly all Terraform providers and modules, and is a drop-in replacement — if you've seen Terraform docs, you already know OpenTofu. Every `terraform` command has a `tofu` equivalent.

---

## 3. Mental Model: The Four Concepts You Need

Before you touch any code, internalize these four things. Once this clicks, everything else is just details.

```
  You write CONFIG files          OpenTofu talks to PROVIDERS        Resources exist in the CLOUD
  (what you want)                 (plugins that know AWS APIs)       (what actually exists)
       │                                │                                  │
       ▼                                ▼                                  ▼
  ┌──────────┐    tofu plan     ┌──────────────┐                  ┌──────────────┐
  │ .tf files │ ──────────────► │   OpenTofu    │ ◄─── compares ──│  STATE file  │
  │           │                 │   Engine      │ ──── reads ────►│ (.tfstate)   │
  └──────────┘                  └──────────────┘                  └──────────────┘
                                       │                                  ▲
                                  tofu apply                              │
                                       │          records what it did     │
                                       ▼               into state         │
                                ┌──────────────┐ ─────────────────────────┘
                                │  AWS / Cloud │
                                └──────────────┘
```

### 3.1 Providers

A **provider** is a plugin that knows how to talk to a specific API. The `aws` provider knows how to call the AWS SDK. The `google` provider knows GCP. You declare which providers you need, OpenTofu downloads them during `init`.

Think of it like a **database driver** — your app doesn't talk raw TCP to Postgres; it uses `pg` or JDBC. Similarly, OpenTofu doesn't talk raw HTTP to AWS; it uses the `hashicorp/aws` provider.

### 3.2 Resources

A **resource** is a single piece of infrastructure: an S3 bucket, an EC2 instance, an IAM role. You declare resources in `.tf` files. Each resource has a **type** (`aws_s3_bucket`) and a **local name** you choose (`my_bucket`). Together they form a unique address: `aws_s3_bucket.my_bucket`.

### 3.3 State

**State** is OpenTofu's memory. It's a JSON file (`.tfstate`) that maps every resource in your config to a real-world cloud resource. When you say "I want an S3 bucket named `my-logs`," and OpenTofu creates it, it writes down: "resource `aws_s3_bucket.my_bucket` = real bucket `arn:aws:s3:::my-logs`."

Why does this matter? Because **next time you run `tofu plan`**, OpenTofu compares your config to the state to figure out what changed. Without state, it would have no idea what it already created.

> **This is the part that confuses EVERYONE at first.** The state file is not your config. It's not the cloud. It's the **bridge between the two**. If your state says a bucket exists but someone deleted it in the console, OpenTofu will try to recreate it. If your state is lost, OpenTofu thinks nothing exists and will try to create duplicates.

### 3.4 The Plan/Apply Loop

This is your daily workflow:

```
  ┌──────────────────────────────────────────────────┐
  │                                                    │
  │   1. Edit .tf files                                │
  │        │                                           │
  │        ▼                                           │
  │   2. tofu plan         (dry run — shows changes)   │
  │        │                                           │
  │        ▼                                           │
  │   3. READ THE PLAN     (seriously, read it)        │
  │        │                                           │
  │        ▼                                           │
  │   4. tofu apply        (executes the changes)      │
  │        │                                           │
  │        ▼                                           │
  │   5. Commit .tf files to git                       │
  │                                                    │
  └──────────────────────────────────────────────────┘
```

`plan` is your safety net. It shows you exactly what will be **created**, **modified**, or **destroyed** — before anything happens. Treat it like a diff before a merge.

---

## 4. HCL Syntax in 3 Minutes

HCL (HashiCorp Configuration Language) is not a programming language — it's a **configuration language** with just enough features to keep you sane. If you can read JSON, you can read HCL.

```hcl
# --- BLOCKS are the building units ---
# Pattern: block_type "label1" "label2" { ... }

resource "aws_s3_bucket" "my_bucket" {    # type = aws_s3_bucket, name = my_bucket
  bucket = "my-unique-bucket-name-12345"  # argument = value
  
  tags = {                                # maps use { key = value } syntax
    Environment = "dev"
    Team        = "platform"
  }
}

# --- REFERENCES point to other resources ---
resource "aws_s3_bucket_versioning" "my_bucket" {
  bucket = aws_s3_bucket.my_bucket.id     # <type>.<name>.<attribute>
  versioning_configuration {
    status = "Enabled"
  }
}

# --- VARIABLES are inputs ---
variable "environment" {
  type    = string
  default = "dev"
}

# --- STRING INTERPOLATION ---
bucket = "${var.project}-${var.environment}-logs"

# --- LOCALS are computed constants ---
locals {
  bucket_prefix = "${var.project}-${var.environment}"
}

# --- OUTPUTS expose values after apply ---
output "bucket_arn" {
  value = aws_s3_bucket.my_bucket.arn
}
```

**That's it.** Blocks, arguments, references, variables, locals, outputs. Everything else is just more block types.

---

## 5. Machine Setup

### 5.1 Install OpenTofu

| OS | Command |
|---|---|
| **macOS (Homebrew)** | `brew install opentofu` |
| **Linux (snap)** | `snap install --classic opentofu` |
| **Linux (apt — Debian/Ubuntu)** | See [opentofu.org/docs/intro/install](https://opentofu.org/docs/intro/install/) |
| **Windows (Chocolatey)** | `choco install opentofu` |

Verify:

```bash
tofu --version
# OpenTofu v1.9.x
```

### 5.2 AWS Credentials

OpenTofu uses the **same credential chain** as the AWS CLI. If `aws s3 ls` works, OpenTofu will work.

```bash
# Option 1: Install AWS CLI and configure
brew install awscli
aws configure
# Enter: Access Key ID, Secret Access Key, region (us-east-1), output (json)

# Option 2: Environment variables (great for CI)
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="wJalr..."
export AWS_DEFAULT_REGION="us-east-1"
```

> **Security note:** Never hardcode credentials in `.tf` files. Use `aws configure`, environment variables, or IAM roles. Add `*.tfvars` to `.gitignore` if your tfvars contain secrets (they shouldn't, but people do it).

### 5.3 Editor Setup

Install the "HashiCorp HCL" extension in VS Code (or IntelliJ's HCL plugin). It gives you syntax highlighting, auto-formatting, and autocomplete. Works perfectly with OpenTofu since the syntax is identical.

---

## 6. First Config: Hello, OpenTofu

Create a project folder and one file:

```bash
mkdir learn-tofu && cd learn-tofu
touch main.tf
```

```hcl
# main.tf — the simplest possible config

# Tell OpenTofu: "I need to talk to AWS"
terraform {                            # yes, the block is still called "terraform" — historical naming
  required_providers {
    aws = {
      source  = "hashicorp/aws"        # download the AWS provider from the registry
      version = "~> 5.0"               # any 5.x version (the ~> means "compatible with")
    }
  }
  required_version = ">= 1.6.0"       # minimum OpenTofu version
}

# Configure the AWS provider
provider "aws" {
  region = "us-east-1"                 # which AWS region to create resources in
}
```

Now run your first commands:

```bash
# Step 1: Download the AWS provider plugin
tofu init

# You'll see:
# Initializing provider plugins...
# - Installing hashicorp/aws v5.x.x...
# OpenTofu has been successfully initialized!

# Step 2: See what OpenTofu would do (nothing — we have no resources yet)
tofu plan

# You'll see:
# No changes. Your infrastructure matches the configuration.
```

**What just happened:**
- `tofu init` created a `.terraform/` directory with the downloaded AWS provider binary
- `tofu init` created a `.terraform.lock.hcl` file pinning the exact provider version (commit this to git!)
- `tofu plan` compared your config (no resources) to state (no state yet) and found nothing to do

---

## 7. Adding Your First Resource: An S3 Bucket

Add this to `main.tf` below the provider block:

```hcl
# Our first real resource!
resource "aws_s3_bucket" "my_first_bucket" {
  bucket = "my-tofu-learning-bucket-12345"   # must be globally unique across ALL of AWS
  
  tags = {
    Name        = "My First Tofu Bucket"
    ManagedBy   = "opentofu"
  }
}
```

### 7.1 Plan It

```bash
tofu plan
```

```
OpenTofu will perform the following actions:

  # aws_s3_bucket.my_first_bucket will be created
  + resource "aws_s3_bucket" "my_first_bucket" {
      + arn                         = (known after apply)
      + bucket                      = "my-tofu-learning-bucket-12345"
      + id                          = (known after apply)
      + region                      = (known after apply)
      + tags                        = {
          + "ManagedBy" = "opentofu"
          + "Name"      = "My First Tofu Bucket"
        }
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

The `+` means **create**. Every attribute marked `(known after apply)` is something AWS will generate (like the ARN).

### 7.2 Apply It

```bash
tofu apply
# Type "yes" when prompted
```

OpenTofu calls the AWS API, creates the bucket, and writes the result to `terraform.tfstate`. Your bucket now exists.

### 7.3 Modify It

Change the tag:

```hcl
  tags = {
    Name        = "My Renamed Bucket"       # changed
    ManagedBy   = "opentofu"
  }
```

```bash
tofu plan
```

```
  # aws_s3_bucket.my_first_bucket will be updated in-place
  ~ resource "aws_s3_bucket" "my_first_bucket" {
      ~ tags   = {
          ~ "Name"      = "My First Tofu Bucket" -> "My Renamed Bucket"
        }
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

The `~` means **update in-place**. Safe. Apply it.

### 7.4 Destroy It

```bash
tofu destroy
# Type "yes" when prompted
```

Gone. The bucket is deleted from AWS, and the state file is updated to reflect that nothing exists.

---

## 8. The Five Daily Commands

| Command | What it does | When to use |
|---|---|---|
| `tofu init` | Downloads providers, initializes backend | First time, or after adding/changing providers |
| `tofu plan` | Dry run — shows what would change | **Every time** before apply. Read it carefully |
| `tofu apply` | Executes the plan — creates/updates/destroys resources | After you've reviewed the plan |
| `tofu destroy` | Tears down **everything** managed by this config | Cleaning up dev/test environments |
| `tofu fmt` | Auto-formats `.tf` files (consistent indentation, alignment) | Before every commit. Add to pre-commit hook |

**Bonus commands you'll use weekly:**

| Command | What it does |
|---|---|
| `tofu validate` | Checks syntax without talking to AWS |
| `tofu output` | Shows output values from last apply |
| `tofu state list` | Lists all resources in state |
| `tofu state show <addr>` | Shows details of one resource in state |

---

## 9. Variables and Outputs

### 9.1 Variables (Inputs)

Variables make your config reusable. Instead of hardcoding `"us-east-1"` everywhere, parameterize it:

```hcl
# variables.tf

variable "aws_region" {
  description = "AWS region for all resources"   # always write a description — future you will thank you
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  # no default — forces the caller to provide a value
}

variable "enable_versioning" {
  description = "Enable S3 bucket versioning"
  type        = bool
  default     = true
}

variable "allowed_origins" {
  description = "CORS allowed origins for the uploads bucket"
  type        = list(string)
  default     = ["https://myapp.com"]
}

variable "extra_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
```

**How values are provided (in order of precedence, highest first):**

| Method | Example | Use case |
|---|---|---|
| CLI flag | `tofu apply -var="environment=prod"` | One-offs, CI overrides |
| `.tfvars` file | `environment = "dev"` in `terraform.tfvars` | Per-environment defaults |
| Environment var | `export TF_VAR_environment=dev` | CI/CD pipelines |
| `default` in variable block | `default = "us-east-1"` | Sensible defaults |
| Interactive prompt | OpenTofu asks you at runtime | Don't rely on this |

### 9.2 Outputs (Exports)

Outputs expose values after `apply`. Essential for passing data between configs or just seeing what was created:

```hcl
# outputs.tf

output "bucket_name" {
  description = "Name of the created S3 bucket"
  value       = aws_s3_bucket.my_first_bucket.id
}

output "bucket_arn" {
  description = "ARN of the created S3 bucket"
  value       = aws_s3_bucket.my_first_bucket.arn
}

# After apply, see all outputs:
# tofu output
# tofu output bucket_arn    ← just one value
```

---

## 10. Realistic File Layout

Forget having one giant `main.tf`. Here's how real projects are organized:

```
my-project/
├── providers.tf          # provider config, required_providers, backend
├── variables.tf          # all input variables
├── outputs.tf            # all outputs
├── main.tf               # primary resources (or split further — see below)
├── s3.tf                 # S3-specific resources (for larger projects)
├── rds.tf                # database resources
├── iam.tf                # IAM roles and policies
├── terraform.tfvars      # actual values for this environment
├── terraform.tfvars.example  # checked into git — shows required vars without real values
├── .terraform.lock.hcl   # provider version lock — COMMIT THIS
├── .gitignore            # ignore .terraform/, *.tfstate, *.tfvars (if secrets)
└── README.md
```

**`.gitignore` for every OpenTofu project:**

```gitignore
# Local provider plugins
.terraform/

# State files — NEVER commit these (they contain secrets and resource IDs)
*.tfstate
*.tfstate.*

# Crash logs
crash.log
crash.*.log

# Override files (used for local dev overrides)
override.tf
override.tf.json
*_override.tf
*_override.tf.json

# If your .tfvars contain secrets, ignore them too
# terraform.tfvars
```

> **This is where it gets really interesting!** The file names don't matter to OpenTofu — it reads **all `.tf` files** in the directory and merges them. You could name them `pizza.tf` and `taco.tf` and it would work fine. The naming convention is for **humans**.

---

## 11. Concrete Example: Two Production-Grade S3 Buckets

Let's build something real. Two S3 buckets:
- **Logs bucket** — stores application and access logs. Private, encrypted, lifecycle rules to delete old logs.
- **Uploads bucket** — stores user-uploaded files. Private, encrypted, versioned, CORS-enabled for browser uploads.

We'll use the **modern split-resource pattern** — where each concern (versioning, encryption, public access block) is a separate resource instead of inline arguments. This is the current AWS provider best practice (the old inline style is deprecated and will be removed).

```
s3-example/
├── providers.tf
├── variables.tf
├── outputs.tf
├── main.tf
├── terraform.tfvars
└── .terraform.lock.hcl
```

### `providers.tf`

```hcl
# providers.tf — provider configuration and version constraints

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # For a real project, you'd add a backend block here for remote state:
  # backend "s3" {
  #   bucket = "my-terraform-state-bucket"
  #   key    = "s3-example/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "aws" {
  region = var.aws_region

  # default_tags applies these tags to EVERY resource automatically.
  # This is incredibly powerful — no more forgetting to tag something.
  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "opentofu"
    }
  }
}
```

### `variables.tf`

```hcl
# variables.tf — all inputs in one place

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name — used in resource naming and tags"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "log_retention_days" {
  description = "Days to keep log files before deleting"
  type        = number
  default     = 90
}

variable "log_archive_days" {
  description = "Days before transitioning logs to cheaper storage (Glacier)"
  type        = number
  default     = 30
}

variable "uploads_cors_origins" {
  description = "Allowed origins for CORS on the uploads bucket"
  type        = list(string)
  default     = ["*"]    # override this in prod!
}
```

### `terraform.tfvars`

```hcl
# terraform.tfvars — actual values for this deployment
# Tip: create one of these per environment (dev.tfvars, prod.tfvars)
#      and apply with: tofu apply -var-file="prod.tfvars"

project_name       = "myapp"
environment        = "dev"
aws_region         = "us-east-1"
log_retention_days = 90
log_archive_days   = 30
uploads_cors_origins = ["http://localhost:3000", "https://myapp.dev"]
```

### `main.tf`

```hcl
# main.tf — two production-grade S3 buckets
# Pattern: each concern (versioning, encryption, access) is a SEPARATE resource.
# Why? The AWS provider deprecated inline config for S3. Separate resources are
# also easier to read, diff, and conditionally apply.

locals {
  # Computed prefix used for bucket names — keeps naming consistent
  prefix = "${var.project_name}-${var.environment}"
}

# ===========================================================================
# LOGS BUCKET
# Purpose: application logs, access logs, audit trails
# ===========================================================================

resource "aws_s3_bucket" "logs" {
  # Bucket names must be globally unique. The prefix + "logs" + account info helps.
  bucket = "${local.prefix}-logs"

  # force_destroy = true lets you `tofu destroy` even if the bucket has objects.
  # ONLY use this in dev. In prod, you WANT the safety net of "can't delete non-empty bucket."
  force_destroy = var.environment != "prod"

  tags = {
    Purpose = "application-logs"
  }
}

# Block ALL public access — logs should never be public
resource "aws_s3_bucket_public_access_block" "logs" {
  bucket = aws_s3_bucket.logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Encrypt at rest with AWS-managed keys (SSE-S3)
# Why SSE-S3 and not SSE-KMS? KMS adds cost per API call and complexity.
# SSE-S3 is free and sufficient for most use cases.
resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"    # SSE-S3: free, automatic, no key management
    }
    bucket_key_enabled = true     # reduces KMS costs if you switch to KMS later
  }
}

# Lifecycle rules — automatically manage log retention
# Without this, storage costs grow forever. Ask me how I know.
resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    id     = "archive-then-delete"
    status = "Enabled"

    # Move to Glacier after N days (cheap cold storage — pennies per GB)
    transition {
      days          = var.log_archive_days
      storage_class = "GLACIER"
    }

    # Delete after N days — logs older than this have no value
    expiration {
      days = var.log_retention_days
    }
  }

  rule {
    id     = "cleanup-incomplete-uploads"
    status = "Enabled"

    # Abort multipart uploads that were never completed (leaked storage)
    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# No versioning on logs — they're append-only, no need to keep old versions
resource "aws_s3_bucket_versioning" "logs" {
  bucket = aws_s3_bucket.logs.id

  versioning_configuration {
    status = "Suspended"
  }
}


# ===========================================================================
# UPLOADS BUCKET
# Purpose: user-uploaded files (avatars, documents, etc.)
# ===========================================================================

resource "aws_s3_bucket" "uploads" {
  bucket        = "${local.prefix}-uploads"
  force_destroy = var.environment != "prod"

  tags = {
    Purpose = "user-uploads"
  }
}

# Block public access — serve files through CloudFront or signed URLs, not public S3
resource "aws_s3_bucket_public_access_block" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Encrypt uploads at rest
resource "aws_s3_bucket_server_side_encryption_configuration" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

# Version user uploads — lets users recover overwritten or deleted files
resource "aws_s3_bucket_versioning" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  versioning_configuration {
    status = "Enabled"
  }
}

# CORS — required for browser-based direct uploads (presigned URLs)
# Without CORS, the browser blocks the PUT request to S3.
resource "aws_s3_bucket_cors_configuration" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST"]     # PUT for direct upload, GET for retrieval
    allowed_origins = var.uploads_cors_origins    # lock this down in prod!
    expose_headers  = ["ETag"]                   # needed for multipart upload completion
    max_age_seconds = 3600                       # browser caches CORS preflight for 1 hour
  }
}

# Lifecycle: clean up incomplete multipart uploads and old versions
resource "aws_s3_bucket_lifecycle_configuration" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  rule {
    id     = "cleanup-old-versions"
    status = "Enabled"

    # Delete non-current versions after 30 days
    # (versioning keeps every old version forever by default — this caps it)
    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }

  rule {
    id     = "cleanup-incomplete-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 3
    }
  }
}
```

### `outputs.tf`

```hcl
# outputs.tf — expose bucket info for other configs or scripts

output "logs_bucket_name" {
  description = "Name of the logs S3 bucket"
  value       = aws_s3_bucket.logs.id
}

output "logs_bucket_arn" {
  description = "ARN of the logs S3 bucket"
  value       = aws_s3_bucket.logs.arn
}

output "uploads_bucket_name" {
  description = "Name of the uploads S3 bucket"
  value       = aws_s3_bucket.uploads.id
}

output "uploads_bucket_arn" {
  description = "ARN of the uploads S3 bucket"
  value       = aws_s3_bucket.uploads.arn
}

output "uploads_bucket_domain" {
  description = "Regional domain name of the uploads bucket (for CloudFront)"
  value       = aws_s3_bucket.uploads.bucket_regional_domain_name
}
```

### Running It

```bash
cd s3-example
tofu init          # download the AWS provider
tofu plan          # review — you should see 12 resources to create
tofu apply         # type "yes" — buckets are live
tofu output        # see the bucket names and ARNs

# When you're done experimenting:
tofu destroy       # tears everything down
```

---

## 12. Concrete Example: Aurora PostgreSQL Serverless v2

### Can I Run This Locally?

Honest answer: **No.** Aurora is an AWS-proprietary service. There's no local emulator that faithfully reproduces Aurora's behavior. Here are your options:

| Option | Fidelity | Cost | Best for |
|---|---|---|---|
| **Docker Postgres** | Standard Postgres only — no Aurora-specific features | Free | Application development, query testing |
| **LocalStack Pro** | Simulates Aurora API but not the engine | ~$35/mo | Testing IaC, API integration tests |
| **Cheap dev cluster on AWS** | Real Aurora — the actual thing | ~$0.06/hr (~$44/mo) | Integration tests, realistic performance testing |
| **`tofu plan` only** | Validates config syntax and provider logic, creates nothing | Free | Validating IaC changes in CI |

**Recommendation:** Use Docker Postgres for daily development (your app talks standard Postgres anyway), and use `tofu plan` in CI to validate your IaC. Spin up a real Aurora cluster only when you need to test Aurora-specific behavior.

### Local Postgres Substitute

```yaml
# docker-compose.yml — local stand-in for Aurora PostgreSQL
# Your app connects to this the same way it would connect to Aurora.
# Connection string: postgresql://myapp:localdev123@localhost:5432/myapp

services:
  postgres:
    image: postgres:16                    # match your Aurora Postgres version
    container_name: local-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: myapp                  # same as your Aurora database name
      POSTGRES_USER: myapp               # same as your Aurora master username
      POSTGRES_PASSWORD: localdev123      # obviously not production
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data   # persist data across restarts
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myapp"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
```

```bash
docker compose up -d
# Your app connects to: postgresql://myapp:localdev123@localhost:5432/myapp
```

### The Real Aurora Cluster

File layout:

```
aurora-example/
├── providers.tf
├── variables.tf
├── outputs.tf
├── main.tf
├── terraform.tfvars
└── docker-compose.yml    # local Postgres substitute
```

### `providers.tf`

```hcl
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "opentofu"
    }
  }
}
```

### `variables.tf`

```hcl
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name — used in naming"
  type        = string
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
}

variable "db_name" {
  description = "Name of the initial database"
  type        = string
  default     = "myapp"
}

variable "db_master_username" {
  description = "Master username for the database"
  type        = string
  default     = "myapp"
}

variable "min_acu" {
  description = "Minimum Aurora Capacity Units (0.5 = smallest, cheapest)"
  type        = number
  default     = 0.5     # ~$0.06/hr — the minimum for Serverless v2
}

variable "max_acu" {
  description = "Maximum Aurora Capacity Units (scales up under load)"
  type        = number
  default     = 2.0     # cap it low for dev — costs add up fast
}
```

### `main.tf`

```hcl
# main.tf — Aurora PostgreSQL Serverless v2 cluster

locals {
  prefix = "${var.project_name}-${var.environment}"
}

# ===========================================================================
# NETWORKING — use the default VPC for simplicity
# In production, you'd use a dedicated VPC with private subnets.
# ===========================================================================

# Look up the default VPC (every AWS account has one)
data "aws_vpc" "default" {
  default = true
}

# Look up all subnets in the default VPC
data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# Security group — controls who can connect to the database
resource "aws_security_group" "aurora" {
  name_prefix = "${local.prefix}-aurora-"      # name_prefix lets AWS add a unique suffix
  vpc_id      = data.aws_vpc.default.id
  description = "Allow PostgreSQL access to Aurora cluster"

  # Allow inbound Postgres traffic from within the VPC
  ingress {
    description = "PostgreSQL from VPC"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default.cidr_block]    # only from within the VPC
  }

  # Allow all outbound (Aurora needs to reach AWS APIs for metrics, etc.)
  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Lifecycle trick: create the new SG before destroying the old one during changes.
  # Without this, there's a brief period where the DB has no security group.
  lifecycle {
    create_before_destroy = true
  }
}

# DB subnet group — tells Aurora which subnets it can use
resource "aws_db_subnet_group" "aurora" {
  name       = "${local.prefix}-aurora"
  subnet_ids = data.aws_subnets.default.ids

  tags = {
    Name = "${local.prefix}-aurora-subnet-group"
  }
}

# ===========================================================================
# PASSWORD — generate a random password and store it in Secrets Manager
# ===========================================================================

# Generate a random password — never hardcode database passwords
resource "random_password" "db_master" {
  length           = 32
  special          = true
  override_special = "!#$%^&*()-_=+[]{}|:,.<>?"   # exclude chars that break connection strings
}

# Store the password in AWS Secrets Manager
# Why Secrets Manager? Your app can fetch the password at runtime without it
# ever touching config files, env vars, or logs.
resource "aws_secretsmanager_secret" "db_master" {
  name                    = "${local.prefix}/aurora/master-password"
  description             = "Master password for Aurora PostgreSQL cluster"
  recovery_window_in_days = var.environment == "prod" ? 30 : 0    # 0 = no recovery window in dev (instant delete)
}

resource "aws_secretsmanager_secret_version" "db_master" {
  secret_id = aws_secretsmanager_secret.db_master.id

  # Store as JSON so your app can parse it predictably
  secret_string = jsonencode({
    username = var.db_master_username
    password = random_password.db_master.result
    engine   = "postgres"
    port     = 5432
    dbname   = var.db_name
  })
}

# ===========================================================================
# AURORA CLUSTER — the main event
# ===========================================================================

resource "aws_rds_cluster" "main" {
  cluster_identifier = "${local.prefix}-aurora"

  # Engine config
  engine         = "aurora-postgresql"
  engine_mode    = "provisioned"                 # "provisioned" is required for Serverless v2
  engine_version = "16.4"                        # PostgreSQL 16 — check AWS docs for latest

  # Database config
  database_name   = var.db_name
  master_username = var.db_master_username
  master_password = random_password.db_master.result

  # Networking
  db_subnet_group_name   = aws_db_subnet_group.aurora.name
  vpc_security_group_ids = [aws_security_group.aurora.id]

  # Serverless v2 scaling — this is the magic
  # Aurora scales compute up and down automatically based on load.
  # min = 0.5 ACU ≈ 1 GB RAM, ~$0.06/hr
  # max = 2.0 ACU ≈ 4 GB RAM, ~$0.24/hr (scales only when needed)
  serverlessv2_scaling_configuration {
    min_capacity = var.min_acu
    max_capacity = var.max_acu
  }

  # Storage encryption — always on, no reason not to
  storage_encrypted = true

  # ⚠️ DEV ONLY settings — change these for production!
  skip_final_snapshot = true                     # prod: false (take a snapshot before deletion)
  deletion_protection = false                    # prod: true  (prevents accidental `tofu destroy`)

  # Apply changes immediately in dev. In prod, you'd want to schedule a maintenance window.
  apply_immediately = var.environment != "prod"

  tags = {
    Name = "${local.prefix}-aurora-cluster"
  }
}

# Aurora needs at least one instance to serve traffic.
# The instance type "db.serverless" means "let Serverless v2 manage the compute."
resource "aws_rds_cluster_instance" "main" {
  identifier         = "${local.prefix}-aurora-instance-1"
  cluster_identifier = aws_rds_cluster.main.id

  instance_class = "db.serverless"               # this is the Serverless v2 magic — not a fixed size
  engine         = aws_rds_cluster.main.engine
  engine_version = aws_rds_cluster.main.engine_version

  # Performance Insights — free tier includes 7 days of retention
  # Gives you slow query analysis, wait event breakdowns, etc.
  performance_insights_enabled          = true
  performance_insights_retention_period = 7      # days (free tier)

  tags = {
    Name = "${local.prefix}-aurora-instance-1"
  }
}
```

### `outputs.tf`

```hcl
output "cluster_endpoint" {
  description = "Writer endpoint — use this for read-write connections"
  value       = aws_rds_cluster.main.endpoint
}

output "cluster_reader_endpoint" {
  description = "Reader endpoint — use this for read-only connections (load balanced)"
  value       = aws_rds_cluster.main.reader_endpoint
}

output "cluster_port" {
  description = "Database port"
  value       = aws_rds_cluster.main.port
}

output "database_name" {
  description = "Name of the initial database"
  value       = aws_rds_cluster.main.database_name
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret containing credentials"
  value       = aws_secretsmanager_secret.db_master.arn
}

# Connection string for convenience (without password — fetch from Secrets Manager)
output "connection_hint" {
  description = "How to connect (fetch password from Secrets Manager)"
  value       = "psql -h ${aws_rds_cluster.main.endpoint} -p ${aws_rds_cluster.main.port} -U ${var.db_master_username} -d ${var.db_name}"
}
```

### `terraform.tfvars`

```hcl
project_name       = "myapp"
environment        = "dev"
aws_region         = "us-east-1"
db_name            = "myapp"
db_master_username = "myapp"
min_acu            = 0.5      # minimum cost: ~$0.06/hr
max_acu            = 2.0      # cap for dev — increase for prod
```

### Running It

```bash
cd aurora-example
tofu init
tofu plan          # review — you'll see ~7 resources
tofu apply         # takes 5-10 minutes (Aurora cluster creation is slow)
tofu output        # grab the endpoint

# Connect:
# 1. Get the password from Secrets Manager:
aws secretsmanager get-secret-value \
  --secret-id myapp-dev/aurora/master-password \
  --query SecretString --output text | jq -r .password

# 2. Connect with psql:
psql -h $(tofu output -raw cluster_endpoint) \
     -p 5432 -U myapp -d myapp

# When done:
tofu destroy       # tears everything down — takes a few minutes
```

### Cost

> **Serverless v2 at minimum (0.5 ACU):** ~$0.06/hour = ~$1.44/day = ~$44/month
>
> It scales to zero **billing-wise** only if you stop the cluster. There's no true scale-to-zero like Lambda. At 0.5 ACU, it's as cheap as Aurora gets while running.

### What's Missing for Production

This example is intentionally simplified for learning. Here's what you'd add for a production deployment:

| Missing piece | Why it matters |
|---|---|
| **Remote state backend** | State stored locally = single point of failure. Use S3 + DynamoDB for locking |
| **Private subnets** | Default VPC subnets are public. Databases should be in private subnets with no internet access |
| **Multi-AZ read replicas** | Add 1-2 more `aws_rds_cluster_instance` resources across AZs for failover |
| **`deletion_protection = true`** | Prevents accidental `tofu destroy` from nuking your production database |
| **`skip_final_snapshot = false`** | Takes an automatic backup before any deletion |
| **Automated backups + retention** | `backup_retention_period = 7` (or more) on the cluster |
| **Enhanced monitoring** | `monitoring_interval = 60`, `monitoring_role_arn` — OS-level metrics |
| **CloudWatch alarms** | Alert on high CPU, low memory, connection count, replication lag |
| **IAM database authentication** | Replace password auth with IAM roles for your application |
| **Custom parameter group** | Tune Postgres settings (`shared_buffers`, `max_connections`, etc.) |
| **KMS customer-managed key** | For compliance, use your own KMS key instead of the AWS default |
| **VPC endpoints** | Secrets Manager access without going through the internet |

---

## 13. What to Learn Next (In Order)

This is the order I'd recommend. Each topic builds on the previous one:

| # | Topic | Why now |
|---|---|---|
| 1 | **Remote state (S3 + DynamoDB)** | You cannot collaborate or run CI/CD with local state. This is a hard prerequisite for everything else |
| 2 | **Modules** | Once you have 2+ environments (dev/staging/prod), you'll copy-paste configs. Modules eliminate that. Think of them like functions for infrastructure |
| 3 | **Data sources** | Look up existing resources (VPCs, AMIs, account IDs) instead of hardcoding. You've already used these in the Aurora example |
| 4 | **`count` and `for_each`** | Create multiple similar resources without copy-paste. `for_each` over a map is the most powerful pattern in OpenTofu |
| 5 | **Workspaces** | Manage multiple environments (dev/staging/prod) from the same config. Simpler alternative to directory-per-env |
| 6 | **CI/CD (GitHub Actions)** | Automate `plan` on PRs, `apply` on merge to main. This is when IaC truly pays off — every infra change is reviewed |

---

## 14. Hard-Won Lessons

These are the things that will save you hours of pain. I'm putting them last because they'll make more sense now that you've seen real configs.

### Read Every Plan. Every. Single. Time.

```
  # aws_rds_cluster.main must be replaced
  -/+ resource "aws_rds_cluster" "main" {
```

That `-/+` means **destroy and recreate**. For an S3 bucket, that's annoying. For a production database, that's a catastrophe. The plan tells you. Read it.

### Commit the Lock File

`.terraform.lock.hcl` pins exact provider versions — like `package-lock.json` for npm. If you don't commit it, every team member might get different provider versions, leading to "works on my machine" but for infrastructure.

```bash
git add .terraform.lock.hcl    # always
git add *.tf                    # your config
# NEVER: git add terraform.tfstate
```

### State Drift Is Real

Someone logs into the AWS Console and changes a security group rule. Now the cloud doesn't match your state, and your state doesn't match your config. This is called **drift**.

```bash
tofu plan    # will show "unexpected changes" — this is drift
tofu apply   # will fix it by reverting to what your config says
```

**Best practice:** Don't let people make manual changes. If they must, run `tofu plan` frequently to catch drift early.

### Renames Destroy and Recreate

```hcl
# Before
resource "aws_s3_bucket" "logs" { ... }

# After — you "renamed" it
resource "aws_s3_bucket" "application_logs" { ... }
```

OpenTofu sees this as: **delete** `aws_s3_bucket.logs` and **create** `aws_s3_bucket.application_logs`. Your bucket gets destroyed. The fix:

```bash
# Tell OpenTofu: "the resource moved, don't recreate it"
tofu state mv aws_s3_bucket.logs aws_s3_bucket.application_logs
```

Or in OpenTofu 1.7+, use the `moved` block in your config:

```hcl
moved {
  from = aws_s3_bucket.logs
  to   = aws_s3_bucket.application_logs
}
```

### `force_destroy` Is a Footgun

```hcl
resource "aws_s3_bucket" "uploads" {
  bucket        = "my-uploads"
  force_destroy = true           # ← deletes ALL objects when you destroy the bucket
}
```

In dev, this is convenient. In prod, this deletes every user-uploaded file when someone runs `tofu destroy`. Use variable-based toggling:

```hcl
force_destroy = var.environment != "prod"
```

### Don't Store State Locally for Team Projects

Local state means:
- Only you can run `apply` (the state file is on your laptop)
- No locking — two people can `apply` at the same time and corrupt state
- If your disk dies, you lose track of what exists in the cloud

Set up an S3 backend with DynamoDB locking as soon as you have a second person or a CI pipeline.

### Blast Radius: Keep Configs Small

One giant config that manages your entire AWS account is one bad `apply` away from disaster. Split by concern:

```
infra/
├── networking/       # VPC, subnets, route tables
├── database/         # RDS, ElastiCache
├── storage/          # S3 buckets
├── compute/          # ECS, Lambda, EC2
└── monitoring/       # CloudWatch, alerts
```

Each directory is an independent OpenTofu config with its own state. A bad change to storage can't accidentally break networking.

### The Import Trap

Have existing AWS resources you want to manage with OpenTofu? You can import them:

```bash
tofu import aws_s3_bucket.existing my-existing-bucket-name
```

But `import` only adds the resource to **state**. You still need to write the matching `.tf` config by hand. If your config doesn't match the real resource, the next `apply` will modify it to match your config. Always `plan` after importing.

---

**You now have everything you need to start managing real infrastructure with OpenTofu.** The learning curve is front-loaded — once the mental model clicks (config + state + cloud = the plan/apply loop), everything else is just learning new resource types.

Go build something. Break it. Read the plan. Fix it. That's how you learn.
