# Gravatar System Design

## 1. Quick Summary

**What it does:** Gravatar (Globally Recognized Avatar) provides a centralized avatar service that allows users to upload and manage profile images that can be accessed across multiple websites using their email hash.

**Primary use case:** Websites/applications can display consistent user avatars by querying Gravatar using the user's email address, eliminating the need to store and manage avatars locally.

**Key features:**
- Multi-image upload per user
- Single active avatar selection
- Global CDN distribution
- Email-based lookup using MD5 hash
- Fast, cached image delivery

---

## 2. System Requirements

### Functional Requirements
- Users can upload multiple images to their profile
- Users can mark one image as their active avatar
- Images are accessible via email hash (MD5)
- Support multiple image sizes
- High availability and low latency globally

### Non-Functional Requirements
- **Read-heavy:** 99.9% reads, 0.1% writes
- **Response Time:** < 100ms globally
- **Storage:** Billions of images, petabytes of data
- **Availability:** 99.99% uptime
- **Scalability:** Handle millions of requests per second

---

## 3. Database Schema

### Users Table
```
users
├── user_id (UUID, PK)
├── email (VARCHAR, UNIQUE)
├── email_hash (CHAR(32), UNIQUE) -- MD5 of lowercase email
├── active_image_id (UUID, FK)
├── created_at (TIMESTAMP)
├── updated_at (TIMESTAMP)
└── account_status (ENUM: active, suspended, deleted)
```

### Images Table
```
images
├── image_id (UUID, PK)
├── user_id (UUID, FK)
├── original_url (VARCHAR) -- S3/Cloud storage path
├── cdn_url (VARCHAR) -- CDN endpoint
├── file_size (BIGINT) -- in bytes
├── format (VARCHAR) -- jpg, png, webp
├── width (INT)
├── height (INT)
├── is_active (BOOLEAN)
├── uploaded_at (TIMESTAMP)
└── status (ENUM: processing, active, deleted)
```

### Image Variants Table
```
image_variants
├── variant_id (UUID, PK)
├── image_id (UUID, FK)
├── size (INT) -- 80, 200, 400 pixels
├── url (VARCHAR) -- CDN URL for this size
├── format (VARCHAR)
└── created_at (TIMESTAMP)
```

---

## 4. Database Indexes

### Critical Indexes for Performance

**Users Table:**
- `PRIMARY KEY (user_id)` -- Clustered index
- `UNIQUE INDEX idx_email_hash (email_hash)` -- **Most critical** for lookups
- `INDEX idx_email (email)` -- For user management
- `INDEX idx_active_image (active_image_id)` -- Join optimization

**Images Table:**
- `PRIMARY KEY (image_id)` -- Clustered index
- `INDEX idx_user_active (user_id, is_active)` -- Find user's active image
- `INDEX idx_user_uploaded (user_id, uploaded_at DESC)` -- List user's images
- `INDEX idx_status (status)` -- Cleanup/maintenance queries

**Image Variants Table:**
- `PRIMARY KEY (variant_id)` -- Clustered index
- `UNIQUE INDEX idx_image_size (image_id, size)` -- Prevent duplicate variants
- `INDEX idx_image_id (image_id)` -- Fetch all variants for an image

### Index Strategy Reasoning
- **email_hash index** is the hot path for 99.9% of traffic
- Composite indexes optimize common query patterns
- Covering indexes reduce disk I/O
- Consider partitioning users table by email_hash ranges for horizontal scaling

---

## 5. API Server Design

### API Endpoints

#### Read Operations (Public)
```
GET /avatar/{email_hash}
  Query Params:
    - size: 80 | 200 | 400 (default: 80)
    - default: 404 | mp | identicon | monsterid | robohash
  Response: Image binary or redirect to CDN
  Cache: CDN + Browser (max-age: 86400)
```

```
GET /avatar/{email_hash}.json
  Response: 
    {
      "entry": [{
        "hash": "abc123...",
        "thumbnailUrl": "https://cdn.../80",
        "profileUrl": "https://gravatar.com/abc123"
      }]
    }
  Cache: 3600 seconds
```

#### Write Operations (Authenticated)
```
POST /api/v1/images
  Headers: Authorization: Bearer {token}
  Body: multipart/form-data (image file)
  Response: { image_id, upload_status, processing_eta }
```

```
PUT /api/v1/users/me/active-image
  Headers: Authorization: Bearer {token}
  Body: { image_id }
  Response: { success, active_image_id, cdn_urls }
```

```
DELETE /api/v1/images/{image_id}
  Headers: Authorization: Bearer {token}
  Response: { success }
```

### API Server Architecture

**Load Balancer Layer**
- Geographic DNS routing
- Health checks every 5 seconds
- SSL termination
- Rate limiting: 100 req/sec per IP

**API Gateway**
- Authentication/Authorization
- Request validation
- Rate limiting per user
- API versioning
- Request logging and metrics

**Application Servers (Stateless)**
- Auto-scaling based on CPU/memory
- Connection pooling to database
- Redis caching layer
- Async image processing queue
- Regional deployment (US-East, US-West, EU, Asia)

**Caching Strategy**
- **L1 Cache:** In-memory LRU (hot email hashes)
- **L2 Cache:** Redis cluster (distributed cache)
- **L3 Cache:** CDN edge caches (global distribution)
- Cache key: `avatar:{email_hash}:{size}`
- TTL: 24 hours, invalidate on update

---

## 6. CDN Architecture

### CDN Configuration

**Multi-Tier CDN Strategy**
```
User Request
    ↓
Edge PoP (200+ locations globally)
    ↓ (cache miss)
Regional PoP (10-20 locations)
    ↓ (cache miss)
Origin Shield (2-3 locations)
    ↓ (cache miss)
Origin Server (API + S3)
```

**Edge PoP Configuration**
- **Cache TTL:** 30 days for images (immutable URLs)
- **Negative Cache:** 5 minutes for 404s
- **Stale-While-Revalidate:** Serve stale for 1 hour while refreshing
- **Compression:** Gzip, Brotli enabled
- **HTTP/2:** Enabled for multiplexing

**Cache Keys**
- Include: email_hash, size, format
- Exclude: authentication headers, tracking params
- Normalize: lowercase, sort query params

**Image Optimization at Edge**
- **WebP conversion:** Serve WebP to supporting browsers
- **Responsive images:** Auto-select size based on device
- **Lazy load hints:** Add cache headers for optimal loading
- **Image quality:** Adaptive compression (80-95% based on network)

### CDN Purge Strategy
- **On active image change:** Purge all sizes for that email_hash
- **Purge method:** Use CDN API with tag-based invalidation
- **Propagation time:** 30-60 seconds globally
- **Fallback:** If purge fails, TTL expires in 24h

### Origin Protection
- **Origin Shield:** Single-region cache before origin
- **Rate Limiting:** 10k req/sec per origin server
- **DDoS Protection:** CDN-level filtering
- **Origin Authentication:** Signed requests from CDN

---

## 7. Architecture Workflow

### Upload Flow
1. User uploads image via web/API
2. API server validates (size, format, auth)
3. Store original in S3/Cloud Storage
4. Queue async job for processing
5. Worker generates variants (80px, 200px, 400px)
6. Upload variants to CDN origin storage
7. Update database with URLs
8. Return success to user

### Image Request Flow (Cache Hit)
```
User → DNS → Nearest Edge PoP → Return Cached Image
Time: ~10-50ms
```

### Image Request Flow (Cache Miss)
```
User → Edge PoP (miss) 
    → Regional PoP (miss)
    → Origin Shield (miss)
    → API Server → Redis (miss)
    → Database → Fetch record
    → Return URL → CDN caches at each layer
    → Return to User
Time: ~200-500ms (first request only)
```

### Active Image Update Flow
1. User selects new active image
2. API server updates user.active_image_id
3. Trigger CDN purge for email_hash
4. Invalidate Redis cache
5. Next request fetches new image
6. All layers re-cache

---

## 8. Key Design Decisions

### Why Email Hash (MD5)?
- **Privacy:** Don't expose actual emails in URLs
- **Fixed length:** Consistent URL structure
- **Standard:** MD5 is fast, collision probability negligible for this use case
- **Compatibility:** Widely supported across platforms

### Why Multi-Image Storage?
- Users can pre-upload images before switching
- A/B testing different avatars
- Quick rollback if new avatar has issues
- Business model: premium features (more images)

### Storage vs. CDN vs. Database
- **Original images:** Cloud storage (S3) - durable, cheap
- **Processed variants:** CDN origin - high availability
- **Metadata only:** Database - fast lookups, small footprint
- **Never store images in database:** Blob storage is expensive, slow

### Read-Heavy Optimization
- **Aggressive caching:** 3-tier (API, Redis, CDN)
- **CDN-first architecture:** 99.9% of requests never hit origin
- **Immutable URLs:** Cache forever by using versioned URLs
- **Async writes:** Non-blocking upload processing

### Handling Default Images
- **Generate at edge:** If user has no avatar, CDN generates default
- **Options:** Geometric patterns, robots, monsters, initials
- **Cache defaults:** Even 404s/defaults are cached
- **Personalization:** Hash-based generation (consistent per email)

---

## 9. Scalability Considerations

### Horizontal Scaling
- **API Servers:** Stateless, auto-scale based on traffic
- **Database:** Read replicas (10-20), write primary (1-2)
- **Redis:** Cluster mode with 20-50 nodes
- **CDN:** Scales automatically (provider managed)

### Data Partitioning
- **Users table:** Shard by email_hash range (0-9, a-f)
- **Images table:** Shard by user_id or co-locate with users
- **Avoid cross-shard joins:** Denormalize if needed

### Cost Optimization
- **CDN bandwidth:** Biggest cost, optimize with compression
- **Storage:** Use intelligent tiering (S3 Glacier for old/unused images)
- **Database:** Read replicas only for analytics, not user-facing
- **Image processing:** Spot instances for workers

---

## 10. Quick Reference

### Key Metrics to Monitor
- **CDN hit ratio:** Target > 99%
- **API response time:** p99 < 100ms
- **Image upload time:** p99 < 5 seconds
- **Database query time:** p99 < 10ms
- **Cache hit ratio (Redis):** Target > 95%

### Database Query Patterns
**Most frequent:**
```
SELECT cdn_url FROM images 
WHERE user_id = (SELECT user_id FROM users WHERE email_hash = ?) 
AND is_active = true
```

**Optimized with join:**
```
SELECT i.cdn_url, iv.url as variant_url, iv.size
FROM users u
JOIN images i ON u.active_image_id = i.image_id
JOIN image_variants iv ON i.image_id = iv.image_id
WHERE u.email_hash = ?
```

### Configuration Values
- **Max image size:** 5MB original
- **Allowed formats:** JPG, PNG, GIF, WebP
- **Variant sizes:** 80px, 200px, 400px (square)
- **Max images per user:** 10 (free), 50 (premium)
- **CDN TTL:** 30 days (images), 24 hours (metadata)
- **Rate limits:** 100 req/min (upload), unlimited (read)

### CDN Configuration Checklist
- [ ] Enable Gzip/Brotli compression
- [ ] Configure cache headers (max-age, stale-while-revalidate)
- [ ] Set up origin shield in primary region
- [ ] Enable HTTP/2 and HTTP/3
- [ ] Configure tag-based purging
- [ ] Set up monitoring and alerts
- [ ] Enable DDoS protection
- [ ] Configure CORS for browser access

---

## 11. Common Issues & Solutions

### Issue: CDN cache poisoning
**Solution:** Use signed URLs for authenticated requests, separate cache keys for different user states

### Issue: Slow image uploads
**Solution:** Use resumable uploads, multipart upload for large files, process asynchronously

### Issue: Database hotspot on email_hash
**Solution:** Add in-memory cache (Redis), use database read replicas, consider sharding

### Issue: CDN costs too high
**Solution:** Implement smart compression, limit variant sizes, use tiered storage

### Issue: Stale images after update
**Solution:** Implement proper cache invalidation, use versioned URLs, reduce TTL for user-specific queries

### Issue: DDoS attacks on upload endpoint
**Solution:** Rate limiting per IP and per user, CAPTCHA for new accounts, authentication required

---

## 12. Related Topics

### Technologies & Patterns
- **Image Processing:** ImageMagick, Sharp, libvips
- **CDN Providers:** CloudFlare, Fastly, AWS CloudFront, Akamai
- **Object Storage:** AWS S3, Google Cloud Storage, Azure Blob
- **Caching:** Redis Cluster, Memcached
- **Message Queue:** RabbitMQ, AWS SQS, Kafka (for async processing)

### Similar System Designs
- **Image upload service:** Generic file upload architecture
- **CDN optimization:** Content delivery patterns
- **URL shortener:** Hash-based lookups and redirection
- **OAuth provider:** Token-based authentication at scale

### Further Reading
- Consistent Hashing for CDN routing
- Image optimization techniques
- Database sharding strategies
- Rate limiting algorithms (Token bucket, Leaky bucket)
- CAP theorem and eventual consistency
