# Java: From Basics to Pro

## Java Basics

### Language Fundamentals
- Variables and data types: primitives and objects
- Operators: arithmetic, logical, bitwise, ternary
- Control flow: if-else, switch, loops, break/continue
- Methods: overloading, overriding, varargs
- Classes and objects: instantiation, constructors
- Access modifiers: public, private, protected, package-private
- Static keyword: static methods, variables, blocks
- Final keyword: final classes, methods, variables

### Object-Oriented Programming
- Encapsulation: data hiding, getters/setters
- Inheritance: extends, super, method overriding
- Polymorphism: compile-time and runtime
- Abstraction: abstract classes and interfaces
- Composition vs inheritance
- SOLID principles
- Design patterns: Singleton, Factory, Builder, Strategy

### Exception Handling
- Checked vs unchecked exceptions
- Try-catch-finally blocks
- Try-with-resources
- Custom exceptions
- Exception propagation
- Best practices: when to catch, when to throw

### Memory Management
- Stack vs heap memory
- Garbage collection: GC algorithms, GC tuning
- Memory leaks: common causes and prevention
- JVM memory areas: heap, method area, stack
- OutOfMemoryError and StackOverflowError

### Generics
- Generic classes and methods
- Type parameters and wildcards
- Bounded type parameters
- Type erasure
- Generic collections

### Annotations and Reflection
- Built-in annotations: @Override, @Deprecated, @SuppressWarnings
- Custom annotations
- Reflection API: Class, Method, Field
- Use cases and performance considerations

## Collections Framework

### Core Interfaces
- Collection interface hierarchy
- List: ordered, allows duplicates
- Set: no duplicates, unordered
- Map: key-value pairs
- Queue: FIFO/LIFO operations
- Deque: double-ended queue

### List Implementations
- ArrayList: dynamic array, fast random access
- LinkedList: doubly-linked list, fast insertions/deletions
- Vector: thread-safe, synchronized
- Stack: LIFO structure

### Set Implementations
- HashSet: hash table, O(1) average
- LinkedHashSet: maintains insertion order
- TreeSet: sorted, red-black tree
- EnumSet: optimized for enums

### Map Implementations
- HashMap: hash table, O(1) average
- LinkedHashMap: maintains insertion/access order
- TreeMap: sorted by keys, red-black tree
- Hashtable: thread-safe, synchronized
- ConcurrentHashMap: thread-safe, concurrent

### Queue Implementations
- PriorityQueue: priority-based ordering
- ArrayDeque: resizable array deque
- BlockingQueue: thread-safe queues
- ConcurrentLinkedQueue: lock-free concurrent queue

### Collection Utilities
- Collections class: sorting, searching, synchronization
- Arrays class: array operations
- Comparator and Comparable interfaces
- Stream operations on collections

## Concurrency and Parallelism

### Threading Basics
- Thread creation: extending Thread, implementing Runnable
- Thread lifecycle: NEW, RUNNABLE, BLOCKED, WAITING, TERMINATED
- Thread priorities and scheduling
- Daemon threads
- Thread groups

### Synchronization
- Synchronized keyword: methods and blocks
- Intrinsic locks and monitor concept
- Volatile keyword: visibility guarantees
- Atomic classes: AtomicInteger, AtomicReference
- Lock interface: ReentrantLock, ReadWriteLock
- Condition objects for thread coordination

### Thread Communication
- wait(), notify(), notifyAll()
- Producer-consumer patterns
- BlockingQueue implementations
- CountDownLatch: one-time synchronization
- CyclicBarrier: reusable synchronization
- Semaphore: resource access control
- Exchanger: thread data exchange

### Executor Framework
- Executor and ExecutorService interfaces
- ThreadPoolExecutor: configurable thread pools
- ScheduledExecutorService: delayed and periodic tasks
- Executors utility class: factory methods
- ForkJoinPool: work-stealing for parallel tasks
- CompletableFuture: asynchronous programming

### Concurrent Collections
- ConcurrentHashMap: thread-safe map
- CopyOnWriteArrayList: thread-safe list
- BlockingQueue implementations
- ConcurrentLinkedQueue: lock-free queue
- Thread-safe collections vs synchronized wrappers

### Advanced Concurrency
- ThreadLocal: thread-local variables
- Phaser: advanced synchronization barrier
- StampedLock: optimistic locking
- Lock-free algorithms and data structures
- Memory model: happens-before relationships

## Asynchronous Programming

### CompletableFuture
- Creating CompletableFuture: supplyAsync, runAsync
- Chaining operations: thenApply, thenCompose, thenCombine
- Exception handling: handle, exceptionally
- Combining futures: allOf, anyOf
- Timeouts and cancellation

### Reactive Programming
- Reactive Streams specification
- Project Reactor: Mono and Flux
- RxJava: Observable, Single, Maybe
- Backpressure handling
- Operators: map, filter, flatMap, reduce

### Asynchronous I/O
- NIO: Non-blocking I/O
- Channels: FileChannel, SocketChannel, ServerSocketChannel
- Buffers: ByteBuffer, CharBuffer
- Selectors: multiplexing I/O operations
- AsynchronousFileChannel: async file operations

### Future and Callable
- Future interface: get(), isDone(), cancel()
- Callable vs Runnable
- FutureTask implementation
- Limitations and CompletableFuture advantages

## Spring Framework

### Core Concepts
- Dependency Injection (DI): constructor, setter, field injection
- Inversion of Control (IoC) container
- ApplicationContext vs BeanFactory
- Bean lifecycle: initialization, destruction
- Bean scopes: singleton, prototype, request, session
- @Component, @Service, @Repository, @Controller

### Configuration
- XML-based configuration
- Java-based configuration: @Configuration, @Bean
- Annotation-based configuration: @ComponentScan
- Property injection: @Value, @PropertySource
- Profile-based configuration: @Profile
- Conditional beans: @Conditional

### AOP (Aspect-Oriented Programming)
- Cross-cutting concerns
- Join points, pointcuts, advice
- @Aspect, @Before, @After, @Around
- Proxies: JDK dynamic proxies, CGLIB
- Transaction management with AOP

### Spring MVC
- DispatcherServlet architecture
- @Controller and @RestController
- Request mapping: @RequestMapping, @GetMapping, @PostMapping
- Path variables and request parameters
- Model and ModelAndView
- View resolvers: JSP, Thymeleaf, FreeMarker
- Exception handling: @ControllerAdvice, @ExceptionHandler

### Spring Boot

#### Core Features
- Auto-configuration: convention over configuration
- Starter dependencies: spring-boot-starter-*
- Embedded servers: Tomcat, Jetty, Undertow
- Actuator: production-ready features
- DevTools: hot reload, live reload

#### Configuration
- application.properties and application.yml
- Profile-specific configuration
- Externalized configuration
- @ConfigurationProperties: type-safe configuration
- Environment variables and command-line arguments

#### Web Development
- RESTful APIs: @RestController, ResponseEntity
- Request/Response handling
- Content negotiation
- Error handling: @ControllerAdvice, ErrorController
- Static resources and webjars

#### Testing
- @SpringBootTest: integration testing
- @WebMvcTest: web layer testing
- @DataJpaTest: JPA layer testing
- MockMvc: testing web endpoints
- TestContainers: integration testing with containers

#### Production Features
- Actuator endpoints: health, metrics, info
- Custom health indicators
- Metrics: Micrometer integration
- Logging: Logback, Log4j2 configuration
- Monitoring and management

### Spring Security

#### Authentication
- Authentication mechanisms: form, basic, OAuth2
- UserDetailsService: custom user loading
- Password encoding: BCrypt, Argon2
- JWT (JSON Web Tokens): creation and validation
- OAuth2: authorization server, resource server
- SAML: enterprise SSO

#### Authorization
- Method security: @PreAuthorize, @PostAuthorize
- URL-based security: HttpSecurity configuration
- Role-based access control (RBAC)
- Permission-based access control
- Custom security expressions

#### Advanced Features
- CSRF protection
- CORS configuration
- Session management
- Remember-me functionality
- Security headers configuration
- Custom filters and authentication providers

### Spring Data

#### Spring Data JPA
- Repository interfaces: CrudRepository, JpaRepository
- Query methods: findBy, countBy, deleteBy
- @Query: custom JPQL and native queries
- Specifications: dynamic queries
- Projections: interface and DTO projections
- Auditing: @CreatedDate, @LastModifiedDate

#### Spring Data REST
- RESTful repositories
- HAL format
- Custom endpoints
- Security integration

#### Spring Data MongoDB
- MongoDB repositories
- Document mapping
- Aggregation pipelines
- GridFS support

#### Spring Data Redis
- Redis repositories
- Cache abstraction
- Pub/Sub messaging
- Connection pooling

### JPA (Java Persistence API)

#### Entity Management
- @Entity, @Table, @Id, @GeneratedValue
- Entity relationships: @OneToMany, @ManyToOne, @ManyToMany
- Cascade operations: CascadeType
- Fetch strategies: EAGER vs LAZY
- Entity lifecycle: persist, merge, remove, detach

#### Querying
- JPQL (Java Persistence Query Language)
- Criteria API: type-safe queries
- Native SQL queries
- Named queries: @NamedQuery, @NamedNativeQuery
- Query hints and optimization

#### Advanced Features
- Inheritance strategies: SINGLE_TABLE, JOINED, TABLE_PER_CLASS
- Embedded objects: @Embeddable, @Embedded
- Composite keys: @EmbeddedId
- Versioning: @Version for optimistic locking
- Listeners: @PrePersist, @PostLoad, @EntityListeners

#### Performance
- N+1 query problem and solutions
- Batch operations: @BatchSize
- Second-level cache: EhCache, Hazelcast
- Connection pooling: HikariCP, C3P0
- Query optimization techniques

## Advanced Topics

### JVM Internals
- Class loading: classloaders, class loading process
- Bytecode: JVM instruction set
- JIT compilation: HotSpot optimizations
- GC algorithms: G1, ZGC, Shenandoah
- JVM tuning: heap size, GC parameters
- JVM monitoring: jstat, jmap, jstack

### Performance Optimization
- Profiling tools: JProfiler, VisualVM, async-profiler
- Memory profiling: heap dumps, thread dumps
- CPU profiling: sampling, instrumentation
- Benchmarking: JMH (Java Microbenchmark Harness)
- Optimization techniques: caching, pooling, lazy loading

### Build Tools
- Maven: POM, dependencies, plugins, lifecycle
- Gradle: build scripts, tasks, plugins
- Dependency management
- Multi-module projects
- Build optimization

### Testing
- JUnit 5: @Test, @BeforeEach, @AfterEach
- TestNG: advanced features
- Mockito: mocking and stubbing
- AssertJ: fluent assertions
- Integration testing strategies
- Test coverage: JaCoCo

### Design Patterns
- Creational: Singleton, Factory, Builder, Prototype
- Structural: Adapter, Decorator, Facade, Proxy
- Behavioral: Observer, Strategy, Template Method, Command
- Enterprise patterns: Repository, Service Layer, DTO

### Best Practices
- Code style: naming conventions, formatting
- Documentation: JavaDoc, README
- Error handling strategies
- Logging: SLF4J, Logback, Log4j2
- Security: input validation, SQL injection prevention
- Code review guidelines

