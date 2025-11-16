# Large-scale, Highly-Available Software System

## Goals
- Scalable
- Resilient
- Observable
- Secure
- Cost-effective

## Client experience layer

### Web clients
#### Frameworks
- SPA: React, Angular, Vue, Svelte
- SSR/SSG: Next.js, Nuxt.js, Remix, Astro
- Micro-frontends and Web Components
- Progressive Web Apps and offline-first

#### Performance and scaling
- Bundling and tree shaking
- Code splitting and lazy loading
- Image optimization and responsive images
- Browser caching and CDN for static assets

#### Reliability and UX
- Graceful degradation and skeleton loading
- Service workers and offline support
- Feature flags and experimentation
- Accessibility - a11y and localization - i18n
- Client-side telemetry and error boundaries

### Mobile clients
#### Technologies
- Native: Swift, Kotlin
- Cross-platform: React Native, Flutter
- Hybrid: Ionic, Capacitor

#### Offline and performance
- Local storage - SQLite, Realm, Room, Core Data
- Background sync and delta updates
- App thinning and on-demand features

#### Reliability
- Retry with exponential backoff
- Staged rollouts and feature flags
- Crash reporting and analytics

### Desktop clients
#### Technologies
- Electron, Tauri
- Native desktop frameworks

#### Operations
- Auto-updates and version pinning
- Local caching and offline mode

## Edge, network and CDN layer

### DNS and traffic entry
- DNS providers: Route53, Cloudflare, Google, Azure
- Anycast routing and GeoDNS
- Health checks and failover
- Multiple DNS providers and TTL tuning

### CDN and edge
- CDN providers: Cloudflare, Akamai, Fastly, CloudFront
- Edge compute: Workers, Lambda at Edge
#### Caching strategies
- Cache-Control and ETags
- Stale-while-revalidate
- Cache invalidation and purging
- Multi-CDN and origin shielding

## Load balancing and traffic distribution

### Load balancers
- Software: NGINX, HAProxy, Envoy
- Cloud: ALB, NLB, GCP and Azure load balancers
- Hardware appliances when needed

### Algorithms
- Round robin and weighted round robin
- Least connections and least response time
- IP or consistent hashing
- Random and weighted random

### Resilience and scale
- Active-active vs active-passive
- Health checks and connection draining
- Horizontal scaling and autoscaling of LBs
- Global load balancing - GSLB

## API gateway and edge services

### Gateways and proxies
- Kong, Apigee, Tyk, WSO2
- AWS API Gateway, Azure APIM, GCP Endpoints
- Spring Cloud Gateway, Zuul, Envoy-based gateways

### Security capabilities
- OAuth2 and OpenID Connect
- JWT validation and API keys
- mTLS and integration with IdPs

### Traffic management
- Rate limiting and quotas
- Spike arrest and throttling
- Circuit breakers, retries and timeouts
- Protocol translation: REST and gRPC

### API productization
- Developer portals and API catalog
- Plans and usage tiers
- Versioning and deprecation policies

## Application and service layer

### Architectural styles
- Monolith and modular monolith
- Microservices and domain-driven design
- Event-driven architecture
- Serverless and functions
- Hexagonal and clean architecture

### Microservices
- Bounded contexts and aggregates
#### Service mesh patterns
- Sidecars and traffic policies
- mTLS and observability within mesh
- Data ownership per service

### Serverless
- AWS Lambda, Azure Functions, Cloud Functions
- Container-based serverless: Fargate, Cloud Run
- Scale to zero and pay per request

### Deployment and high availability
- Blue-green deployments
- Canary releases and progressive delivery
- Rolling updates
- Feature flags and kill switches
- Graceful shutdown and health probes

### Tech stacks
- JVM: Java, Kotlin, Scala
- .NET: C#, F#
- Node.js, Deno, Bun
- Python: Django, FastAPI, Flask
- Go, Rust, Elixir, Ruby, PHP

## Communication and integration

### Synchronous APIs
#### REST APIs
- HTTP/1.1, HTTP/2 and HTTP/3
- OpenAPI and schema first
- API versioning strategies

#### GraphQL
- Single endpoint and schema
- Federation and stitched schemas
- Subscriptions for realtime

#### gRPC
- Strongly typed proto contracts
- Unary and streaming RPCs

### Realtime communication
- WebSockets and Socket.io
- Server-sent events
- SignalR and similar frameworks
- Scaling with pub-sub backends and sticky sessions

### Asynchronous messaging
#### Message queues
- RabbitMQ, ActiveMQ, SQS, Service Bus

#### Event streams
- Kafka, Pulsar, Kinesis, Event Hubs
- NATS and other brokers

#### Patterns
- Pub-sub and fan-out/fan-in
- Request-reply via queues
- Dead-letter queues
- Outbox pattern and change data capture
- Saga pattern for distributed transactions

## Data management layer

### Relational databases
- PostgreSQL, MySQL, SQL Server, Oracle
- NewSQL and distributed SQL: Spanner, CockroachDB, TiDB, Aurora
- Schema design and normalization
- Transactions and isolation levels
#### Scaling
- Read replicas and connection pooling
- Sharding and federation
- Automatic failover and multi-region replication

### NoSQL and specialized stores
- Document: MongoDB, CouchDB, Cosmos
- Key value: Redis, Memcached, DynamoDB, etcd
- Column family: Cassandra, HBase, Bigtable
- Graph: Neo4j, ArangoDB, Neptune
- Time series: InfluxDB, TimescaleDB, Prometheus
- Search: Elasticsearch, OpenSearch, Solr, Algolia

### Caching
- Technologies: Redis, Memcached, Varnish, Hazelcast
#### Patterns
- Cache aside
- Write-through and write-behind
- Near cache and L1/L2 caches
- Cache stampede prevention and jitter

### Analytics and warehousing
- Data warehouse: Snowflake, BigQuery, Redshift, Synapse
- Data lake and lakehouse architectures
- Batch and streaming ETL/ELT
- BI tools: Looker, Tableau, Power BI, Superset

## Storage, backup and disaster recovery

### Storage types
- Object storage: S3, GCS, Azure Blob
- Block storage: EBS and managed disks
- File storage: EFS, Azure Files, NFS

### Durability and lifecycle
- Lifecycle policies: hot, warm, cold, archive
- Versioning and object locks
- Cross-region and cross-account replication

### Backup and DR
- Incremental and snapshot backups
- Immutable backups and WORM
- RPO and RTO definitions
#### DR strategies
- Pilot light and warm standby
- Active active multi-region
- Runbooks and regular DR drills

## Streaming, events and workflows

### Stream processing
- Kafka Streams, Flink, Spark Streaming
- Kinesis Data Analytics and similar
- Windowing and watermarking
- Exactly-once or effectively-once semantics

### Workflow orchestration
- Airflow, Prefect, Dagster
- Temporal, Cadence, Step Functions
- Business workflows and human approvals
- Failure handling and retries with backoff

## Security, privacy and compliance

### Identity and access management
#### Identity providers
- Okta, Auth0, Keycloak, Cognito, Azure AD

#### Protocols
- OAuth2 and OpenID Connect
- SAML2 and SCIM

#### Authentication
- MFA and WebAuthn

#### Authorization
- RBAC and ABAC
- Policy engines like OPA

### Network and application security
- WAFs and application firewalls
- IDS and IPS
- DDoS protection and rate limiting
- Zero trust network access
- Secure SDLC and threat modeling

### Secrets and cryptography
- Secrets managers and Vault
- KMS and HSM integration
- TLS and mTLS
- Encryption at rest and in transit
- Key rotation and certificate management

### Compliance and privacy
- GDPR and CCPA
- HIPAA and PCI DSS
- SOC 2 and ISO standards
- Data classification and handling rules
- Data retention and right to be forgotten
- Audit logs and evidence collection

## Observability, reliability and SRE

### Metrics and monitoring
- Prometheus and Grafana
- DataDog, New Relic, Dynatrace
- Cloud monitoring: CloudWatch, Azure Monitor
- SLI, SLO and SLA
- Error budgets and capacity planning

### Logging
- Centralized logging with ELK or OpenSearch
- Splunk and similar platforms
- Structured logs and correlation IDs
- Log retention and sampling

### Tracing
- Distributed tracing with OpenTelemetry
- Jaeger and Zipkin
- End-to-end request traces

### Incident management and resilience
- Alerting and paging
- PagerDuty, Opsgenie and others
- Runbooks and playbooks
- On-call rotations and escalation policies
- Blameless postmortems
- Chaos engineering and game days

## Infrastructure, platform and orchestration

### Compute and virtualization
- Bare metal and VMs
- Containers and Docker
- Serverless runtimes

### Kubernetes and orchestration
- Kubernetes clusters: EKS, AKS, GKE, OpenShift
- Node pools and autoscaling
- Ingress and service discovery
- Network policies and security
- Operators and custom resources

### Infrastructure as code
- Terraform, Pulumi, CloudFormation, ARM
- Configuration management: Ansible, Puppet, Chef
- Policy as code with OPA and Conftest

### GitOps
- Argo CD and Flux
- Git as source of truth
- Declarative environment management

### Service mesh
- Istio, Linkerd, Consul
- Traffic shaping and canaries
- Mesh level mTLS and telemetry

## Development workflow, CI/CD and quality

### Source control and branching
- GitHub, GitLab, Bitbucket
- Trunk based development
- GitFlow where needed
- Code reviews and pull requests

### CI/CD pipelines
- Jenkins, GitHub Actions, GitLab CI, CircleCI
- Build, test, lint and security stages
- Environment promotion: dev, stage, prod
- Progressive delivery and canary analysis

### Testing strategy
- Unit and integration tests
- Contract tests for services
- End to end and UI tests
- Performance and load tests
- Chaos and resilience tests

### Artifacts and environments
- Artifact repositories: Artifactory, Nexus
- Container registries: ECR, ACR, GCR, Docker Hub
- Environment parity and config management
- Ephemeral preview environments

## Data science, ML and personalization

### ML platform
- Feature stores and model registry
- Training pipelines and orchestration
- Offline and online feature computation

### Model serving and inference
- Online real-time inference services
- Batch scoring jobs
- A/B testing of models and shadow deployments

### ML observability
- Monitoring model performance
- Data drift and concept drift detection
- Explainability and fairness checks

## Multi region, global scale and edge

### Topologies
- Single region high availability
- Multi region active active
- Hub and spoke vs mesh

### Data and consistency
- Multi primary databases
- Replication lag and read your writes
- CAP theorem tradeoffs
- Conflict resolution strategies

### Global traffic management
- Latency based routing and GeoDNS
- Edge caching and edge compute
- Regional failover and failback

## Cost optimization and FinOps

### Cloud cost management
- Tagging and cost allocation
- Dashboards and budgets
- Anomaly detection and alerts

### Optimization techniques
- Right sizing compute resources
- Reserved, savings plans and spot instances
- Autoscaling and scale to zero
- Storage tiering and lifecycle rules
- Minimizing data egress and duplication

### Architecture choices
- Serverless vs always on
- Heavy use of caching
- Efficient data layouts and partitioning

## Governance, risk and architecture

### Architecture governance
- Architecture review boards
- Reference architectures and patterns
- Technology radar and approved stacks

### Risk management
- Risk register and mitigation plans
- Business continuity planning
- Third party and vendor risk

### Documentation and diagrams
- C4 model and views for systems
- Sequence and deployment diagrams
- Docs as code and living documentation

## Developer experience and internal platforms

### Developer tooling
- Internal CLIs and templates
- Preconfigured pipelines and blueprints
- Good IDE and editor support

### Internal developer platform
- Self service infrastructure
- Golden paths and paved roads
- Service catalog and scorecards

### Onboarding and knowledge
- Onboarding checklists and guides
- Architecture decision records
- Wikis, runbooks and design docs

## People, process and culture

### Team structures and ownership
- Product teams and platform teams
- Clear service ownership and on-call
- Conway's law aware design

### Processes
- Agile, Scrum, Kanban
- Discovery and delivery pipelines
- Change management and CAB where needed

### Culture
- Psychological safety
- Blamelessness and learning
- Continuous improvement and retrospectives

## Anti single point of failure principles

### Redundancy at every layer
- Multiple instances, zones and regions
- Redundant dependencies and providers
- Avoid hidden SPOFs like auth and CI

### Graceful degradation
- Feature flags for fallbacks
- Read only modes
- Rate limiting and load shedding

### Automated failover and self healing
- Health checks and restarts
- Autoscaling and replacement
- Automated DNS and LB failover

### Data durability and recoverability
- Multiple replicas and backups
- Cross region replication
- Regular restore tests

### Testing for resilience
- Chaos experiments
- Load and stress tests
- DR drills and game days

### Monitoring and feedback loops
- Logs, metrics and traces
- Actionable alerts
- Capacity and trend analysis

### Documentation and readiness
- Runbooks and playbooks
- Up to date diagrams and ownership maps
- Defined RPO and RTO and business impact
