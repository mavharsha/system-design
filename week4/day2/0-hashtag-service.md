## Social networks

#### Hashtag service (instagram)

1. Define tradeoffs



Questions
0. On the fly or precomputation
- On the fly would be expensive as we would have to go through all the posts
- Precomputation better.
    - But with a little bit of staleness

1. What constitutes better UX?
    Better UX (optimizing for this) (Need to render it as soon as possible)
    - As it needs to render as soon as possible, choose to have one api call rather than multiple.
    - For example: GET /tags/sunset
    - Sample response: 
```javascript
{
    tag: "sunset",
    totalPosts: 1000000,
    topPosts: [
        // top 100 posts
    ]
}
```
2. Where would you store the data? SQL/NOSQL
    - Frequent updates (totalPosts)
    - Support of partial updates (totalPosts)


3. Read path & paginations + CDN?

- Hashtag service returns
- Caches to CDN
- UI has the JSON with 100 posts
    - UI lazy loads the posts that are not visible on viewport

4. Write path

- User creates a post
- POST service receives the post
- POST service emits an event to kafka (userId as partition) (ordering of post to be based on userId)
- HashTag service Consumer group listens to these posts
    - Consumer group listens to 100 posts
    - keeps an internal map of the hashtag to count counter
    - Then updates to total posts.

Optimization (in the earlier approach, because the posts listened to are based on userId, the cardinality of hashTags that the consumer group would get would be high)
- HashTag service publishes these posts to another topic in kafka where the partition is based on Hashtag
    - If a post caption has 8 hash tags, this would result in publishing 8 events
    - Auxiliary writes (8)
- New consumer group would listen to events. Fetch a batch of 100 posts based on hashtags.
    - keeps an internal map of the hashtag to count counter
    - Then updates to total posts.
- Use mutex to make sure the map is synchronized. 
- Use threads to make sure infra is utilized.
    - Why wait when one is writing to the db?