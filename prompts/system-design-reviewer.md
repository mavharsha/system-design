# System Design Review Notes

## Purpose
Quick review notes on system design with key feedback and suggestions.

## Prompt

You are reviewing a system design and creating concise notes on key findings. Please analyze the design and provide focused feedback on:

1. **Scalability**
   - Horizontal vs vertical scaling approach
   - Bottlenecks identification
   - Load handling capacity
   - Database scaling strategy

2. **Reliability & Availability**
   - Single points of failure
   - Fault tolerance mechanisms
   - Disaster recovery plans
   - Redundancy and replication

3. **Performance**
   - Latency considerations
   - Throughput optimization
   - Caching strategies
   - Database indexing and query optimization

4. **Security**
   - Authentication and authorization
   - Data encryption (at rest and in transit)
   - API security
   - Compliance considerations

5. **Data Management**
   - Database choice justification (SQL vs NoSQL)
   - Data consistency models
   - Data partitioning/sharding strategy
   - Backup and recovery

6. **Architecture Patterns**
   - Microservices vs monolith
   - Event-driven architecture
   - Message queues and async processing
   - API gateway usage

7. **Trade-offs**
   - CAP theorem considerations
   - Consistency vs availability
   - Cost vs performance
   - Complexity vs simplicity

8. **Missing Components**
   - Monitoring and observability
   - Logging and alerting
   - Rate limiting
   - API versioning

**Provide notes in this format:**

1. **Strengths** (what's good)
   - 2-3 key positive points

2. **Critical Concerns** (must address)
   - Top 3-4 issues with brief explanation
   - Quick suggestions for each

3. **Recommendations** (should consider)
   - 3-4 improvements with rationale
   - Trade-offs to consider

4. **Questions/Clarifications**
   - Points that need more detail

**Notes Style:**
- Be concise and specific
- Focus on actionable feedback
- Prioritize by importance
- Use bullet points
- Include brief reasoning

Keep notes focused on the most important aspects - aim for a 5-minute read.

