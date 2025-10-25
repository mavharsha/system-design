# Architecture Notes

## Purpose
Create concise architectural guidance notes with practical recommendations.

## Prompt

You are providing architectural guidance in note form. Please create focused architecture notes based on the requirements provided.

**Analyze and advise on:**

1. **Requirements Analysis**
   - Clarify functional requirements
   - Identify non-functional requirements (scalability, performance, security)
   - Understand constraints (budget, timeline, team size)
   - Determine success criteria

2. **Architecture Style Recommendation**
   - Monolithic vs Microservices vs Serverless
   - Layered vs Hexagonal vs Clean Architecture
   - Event-driven vs Request-response
   - Justify your recommendation with pros/cons

3. **Technology Stack**
   - Programming languages and frameworks
   - Database choices (SQL, NoSQL, hybrid)
   - Message queues and event streaming
   - Caching solutions
   - Cloud provider recommendations
   - Justify each choice based on requirements

4. **Component Design**
   - Major system components and their responsibilities
   - Component interactions and communication patterns
   - API design approach (REST, GraphQL, gRPC)
   - Service boundaries

5. **Data Architecture**
   - Data modeling approach
   - Database per service vs shared database
   - Data consistency strategy
   - Data migration and versioning
   - Backup and disaster recovery

6. **Scalability Strategy**
   - Horizontal scaling approach
   - Load balancing strategy
   - Caching layers
   - Database scaling (read replicas, sharding)
   - CDN usage for static assets

7. **Reliability & Resilience**
   - Fault tolerance mechanisms
   - Circuit breakers and retry logic
   - Rate limiting and throttling
   - Health checks and monitoring
   - Graceful degradation

8. **Security Architecture**
   - Authentication and authorization strategy
   - API security (OAuth, JWT, API keys)
   - Data encryption approach
   - Security boundaries
   - Compliance requirements (GDPR, HIPAA, etc.)

9. **DevOps & Infrastructure**
   - CI/CD pipeline recommendations
   - Infrastructure as Code approach
   - Container orchestration (Kubernetes, ECS)
   - Monitoring and observability stack
   - Log aggregation strategy

10. **Migration Strategy** (if applicable)
    - Phased migration approach
    - Strangler pattern implementation
    - Risk mitigation
    - Rollback plans

**Provide notes in this format:**

1. **Recommended Approach**
   - Architecture style and why
   - Key technologies (with brief justification)
   - Main components (3-5 bullet points)

2. **Key Decisions & Trade-offs**
   - Top 3-4 critical decisions
   - Pros/cons for each
   - Why this choice fits the requirements

3. **Implementation Priorities**
   - Phase 1: MVP (what to build first)
   - Phase 2: Scale (what comes next)
   - Phase 3: Optimize (future improvements)

4. **Risks & Mitigation**
   - Top 3 risks
   - Quick mitigation strategy for each

5. **Quick Wins**
   - 2-3 things that provide immediate value
   - Low-hanging fruit

**Notes Style:**
- Be practical and actionable
- Focus on critical decisions
- Keep it concise (aim for 10-15 bullet points)
- Highlight trade-offs clearly
- Prioritize by impact

Create notes that give clear direction without overwhelming detail.

