# Java Interview Questions Mind Map

## Java Basics

### Language Fundamentals
- **Variables and Data Types**
  - What is the difference between primitive types and wrapper classes?
  - Explain autoboxing and unboxing with examples.
  - What are the default values of primitive types in Java?
  - Why is String immutable in Java?
  - What is the difference between String, StringBuffer, and StringBuilder?
  
- **Operators**
  - Explain the difference between == and equals() method.
  - What is the difference between & and && operators?
  - How does the ternary operator work?
  - Explain bitwise operators and their use cases.
  
- **Control Flow**
  - What is the difference between break and continue statements?
  - Can you use switch with String in Java? What about enums?
  - Explain labeled break and continue statements.
  - What are the different types of loops in Java?
  
- **Methods**
  - What is method overloading? Can you overload methods by return type?
  - What is method overriding? What are the rules?
  - Explain varargs (variable arguments) in Java.
  - What is the difference between method overloading and overriding?
  
- **Classes and Objects**
  - What is the difference between a class and an object?
  - Explain the object creation process in Java.
  - What happens when you create an object using new keyword?
  - Can you create an object without using new keyword?
  
- **Access Modifiers**
  - Explain all four access modifiers in Java.
  - What is the default access modifier?
  - Can a private method be overridden?
  - What is the difference between protected and package-private?
  
- **Static Keyword**
  - What is the static keyword? When is it used?
  - Can you override a static method?
  - What is a static block? When is it executed?
  - Explain static variables vs instance variables.
  - Can a static method access non-static variables?
  
- **Final Keyword**
  - What does the final keyword mean for classes, methods, and variables?
  - Can a final method be overridden?
  - What is the difference between final, finally, and finalize?
  - Can you reassign a final variable?

### Object-Oriented Programming
- **Encapsulation**
  - What is encapsulation? How is it achieved in Java?
  - Explain the concept of data hiding.
  - Why are getters and setters important?
  - What is the benefit of making fields private?
  
- **Inheritance**
  - What is inheritance? How is it implemented in Java?
  - Explain the super keyword and its usage.
  - What is the difference between super() and this()?
  - Can you inherit from multiple classes in Java?
  - What is the Object class? What methods does it provide?
  
- **Polymorphism**
  - What is polymorphism? Explain compile-time and runtime polymorphism.
  - What is method overriding? What are the rules?
  - Explain dynamic method dispatch.
  - Can you override a static method? Why or why not?
  - What is the difference between method overloading and overriding?
  
- **Abstraction**
  - What is abstraction? How is it achieved in Java?
  - What is the difference between abstract class and interface?
  - Can an abstract class have a constructor?
  - Can you create an instance of an abstract class?
  - What are the differences between abstract class and interface (Java 8+)?
  
- **Composition vs Inheritance**
  - What is composition? When should you use it over inheritance?
  - Explain "favor composition over inheritance" principle.
  - What are the advantages of composition?
  
- **SOLID Principles**
  - Explain each of the SOLID principles.
  - Give an example of Single Responsibility Principle.
  - What is the Open/Closed Principle?
  - Explain Liskov Substitution Principle with an example.
  - What is Interface Segregation Principle?
  - Explain Dependency Inversion Principle.
  
- **Design Patterns**
  - Explain Singleton pattern. How do you implement it thread-safely?
  - What is Factory pattern? When is it used?
  - Explain Builder pattern and its benefits.
  - What is Strategy pattern? Give an example.
  - Explain Observer pattern.
  - What is the difference between Factory and Abstract Factory patterns?

### Exception Handling
- **Exception Types**
  - What is the difference between checked and unchecked exceptions?
  - What is the difference between Error and Exception?
  - Explain the exception hierarchy in Java.
  - When should you use checked vs unchecked exceptions?
  
- **Exception Handling Mechanisms**
  - Explain try-catch-finally blocks.
  - Can you have multiple catch blocks? What is the order?
  - What is try-with-resources? When should you use it?
  - Can you have a try block without catch or finally?
  - What happens if both catch and finally throw exceptions?
  
- **Custom Exceptions**
  - How do you create a custom exception?
  - When should you create custom exceptions?
  - Should custom exceptions be checked or unchecked?
  
- **Exception Propagation**
  - How does exception propagation work?
  - What happens if an exception is not caught?
  - Explain the throws keyword.
  - What is the difference between throw and throws?
  
- **Best Practices**
  - What are the best practices for exception handling?
  - Should you catch Exception or specific exceptions?
  - Is it good practice to catch and ignore exceptions?
  - How do you handle exceptions in a multi-threaded environment?

### Memory Management
- **Memory Areas**
  - Explain the difference between stack and heap memory.
  - What are the different memory areas in JVM?
  - What is stored in the method area?
  - Explain the stack memory and what it contains.
  
- **Garbage Collection**
  - What is garbage collection? How does it work?
  - Explain different GC algorithms (Mark-Sweep, Mark-Compact, Generational).
  - What is the difference between Minor GC and Major GC?
  - Explain G1, ZGC, and Shenandoah garbage collectors.
  - How do you tune garbage collection?
  - What is the difference between System.gc() and Runtime.gc()?
  
- **Memory Leaks**
  - What is a memory leak in Java?
  - What are common causes of memory leaks?
  - How do you detect memory leaks?
  - How do you prevent memory leaks?
  
- **Memory Errors**
  - What is OutOfMemoryError? What are its types?
  - What is StackOverflowError? How is it caused?
  - How do you handle OutOfMemoryError?
  - How do you increase heap size in Java?

### Generics
- **Generic Basics**
  - What are generics? Why were they introduced?
  - What is type erasure? How does it work?
  - What is the difference between List and List<Object>?
  - Can you use primitive types with generics?
  
- **Wildcards**
  - What are wildcards in generics?
  - Explain ? extends T (upper bounded wildcard).
  - Explain ? super T (lower bounded wildcard).
  - What is the difference between ? and ? extends Object?
  - Explain PECS (Producer Extends, Consumer Super).
  
- **Bounded Type Parameters**
  - What are bounded type parameters?
  - What is the difference between <T extends Number> and <? extends Number>?
  - Can you have multiple bounds in generics?
  
- **Generic Collections**
  - Why should you use generic collections?
  - What happens if you use raw types?
  - Can you create an array of generic types?

### Annotations and Reflection
- **Built-in Annotations**
  - What are annotations? How are they used?
  - Explain @Override, @Deprecated, @SuppressWarnings.
  - What is @FunctionalInterface?
  - Explain @SafeVarargs and @Target.
  
- **Custom Annotations**
  - How do you create a custom annotation?
  - What is @Retention? Explain different retention policies.
  - What is @Target? What are element types?
  
- **Reflection API**
  - What is reflection? When is it used?
  - How do you get the Class object?
  - How do you invoke a method using reflection?
  - What are the performance implications of reflection?
  - When should you use reflection?

## Collections Framework

### Core Interfaces
- **Collection Hierarchy**
  - Explain the Collection interface hierarchy.
  - What is the difference between Collection and Collections?
  - What are the main interfaces in the Collections framework?
  
- **List Interface**
  - What is a List? What are its characteristics?
  - What is the difference between List and Set?
  - Explain the List interface methods.
  
- **Set Interface**
  - What is a Set? What are its characteristics?
  - How does Set ensure uniqueness?
  - What is the difference between Set and List?
  
- **Map Interface**
  - What is a Map? Is it part of Collection interface?
  - Explain the Map interface hierarchy.
  - What is the difference between Map and Collection?
  
- **Queue Interface**
  - What is a Queue? Explain FIFO and LIFO.
  - What is the difference between Queue and Deque?
  - When would you use a Queue?

### List Implementations
- **ArrayList**
  - How does ArrayList work internally?
  - What is the default capacity of ArrayList?
  - What is the time complexity of ArrayList operations?
  - When should you use ArrayList vs LinkedList?
  - How does ArrayList resize?
  
- **LinkedList**
  - How does LinkedList work internally?
  - What is the time complexity of LinkedList operations?
  - When should you use LinkedList vs ArrayList?
  - Is LinkedList faster than ArrayList for insertions?
  
- **Vector**
  - What is Vector? How is it different from ArrayList?
  - Is Vector thread-safe?
  - Should you use Vector in modern Java?
  
- **Stack**
  - What is Stack? How does it work?
  - What operations does Stack support?
  - Is Stack thread-safe?

### Set Implementations
- **HashSet**
  - How does HashSet work internally?
  - What is the time complexity of HashSet operations?
  - How does HashSet ensure uniqueness?
  - What happens if you add duplicate elements?
  - What is the relationship between HashSet and HashMap?
  
- **LinkedHashSet**
  - What is LinkedHashSet? How is it different from HashSet?
  - When should you use LinkedHashSet?
  - Does LinkedHashSet maintain insertion order?
  
- **TreeSet**
  - How does TreeSet work internally?
  - What is the time complexity of TreeSet operations?
  - How does TreeSet maintain order?
  - Can you use TreeSet with objects that don't implement Comparable?
  
- **EnumSet**
  - What is EnumSet? When is it used?
  - What are the advantages of EnumSet?

### Map Implementations
- **HashMap**
  - How does HashMap work internally?
  - Explain the hash function and bucket structure.
  - What is the time complexity of HashMap operations?
  - What happens when HashMap is resized?
  - Explain the difference between HashMap in Java 7 and Java 8.
  - What is the load factor? What is its default value?
  - How does HashMap handle collisions?
  - Can you use a custom object as a key in HashMap?
  - What happens if you don't override equals() and hashCode()?
  
- **LinkedHashMap**
  - What is LinkedHashMap? How is it different from HashMap?
  - Does LinkedHashMap maintain insertion order?
  - What is access order in LinkedHashMap?
  
- **TreeMap**
  - How does TreeMap work internally?
  - What is the time complexity of TreeMap operations?
  - How does TreeMap maintain order?
  - Can you use TreeMap with a custom comparator?
  
- **Hashtable**
  - What is Hashtable? How is it different from HashMap?
  - Is Hashtable thread-safe?
  - Should you use Hashtable in modern Java?
  - What is the difference between Hashtable and HashMap?
  
- **ConcurrentHashMap**
  - What is ConcurrentHashMap? How is it different from HashMap?
  - How does ConcurrentHashMap achieve thread-safety?
  - Explain the internal structure of ConcurrentHashMap.
  - What is the difference between ConcurrentHashMap and synchronized HashMap?

### Queue Implementations
- **PriorityQueue**
  - What is PriorityQueue? How does it work?
  - How does PriorityQueue maintain order?
  - What is the time complexity of PriorityQueue operations?
  
- **ArrayDeque**
  - What is ArrayDeque? When is it used?
  - What is the difference between ArrayDeque and LinkedList?
  - Is ArrayDeque thread-safe?
  
- **BlockingQueue**
  - What is BlockingQueue? When is it used?
  - Explain different BlockingQueue implementations.
  - What is the difference between take() and poll()?

### Collection Utilities
- **Collections Class**
  - What utility methods does the Collections class provide?
  - How do you sort a collection?
  - How do you make a collection thread-safe?
  - Explain binarySearch() in Collections.
  
- **Arrays Class**
  - What utility methods does the Arrays class provide?
  - How do you sort an array?
  - What is the difference between Arrays.sort() and Collections.sort()?
  
- **Comparator and Comparable**
  - What is the difference between Comparator and Comparable?
  - When should you use Comparator vs Comparable?
  - How do you sort objects using Comparator?
  - Can you have multiple comparators for the same class?

## Concurrency and Parallelism

### Threading Basics
- **Thread Creation**
  - What are the different ways to create a thread?
  - What is the difference between extending Thread and implementing Runnable?
  - What is the difference between Runnable and Callable?
  - Can you start a thread twice?
  - What happens if you call run() instead of start()?
  
- **Thread Lifecycle**
  - Explain the thread lifecycle states.
  - What is the difference between BLOCKED and WAITING states?
  - How do you check the state of a thread?
  - What causes a thread to enter BLOCKED state?
  
- **Thread Priorities**
  - What are thread priorities? How do they work?
  - Can you guarantee thread execution order using priorities?
  - What is the default priority of a thread?
  
- **Daemon Threads**
  - What is a daemon thread?
  - What is the difference between daemon and user threads?
  - When should you use daemon threads?

### Synchronization
- **Synchronized Keyword**
  - What is the synchronized keyword? How does it work?
  - What is the difference between synchronized method and synchronized block?
  - What is an intrinsic lock or monitor lock?
  - Can a static method be synchronized?
  - What is the difference between synchronized(this) and synchronized(MyClass.class)?
  - Can you synchronize a constructor?
  
- **Volatile Keyword**
  - What is the volatile keyword? How does it work?
  - What is the difference between volatile and synchronized?
  - Does volatile guarantee thread-safety?
  - Explain visibility guarantees of volatile.
  - When should you use volatile?
  
- **Atomic Classes**
  - What are atomic classes? Give examples.
  - How do atomic classes achieve thread-safety?
  - What is the difference between AtomicInteger and volatile int?
  - Explain compareAndSet (CAS) operation.
  
- **Lock Interface**
  - What is the Lock interface? How is it different from synchronized?
  - What is ReentrantLock? How does it work?
  - What is ReadWriteLock? When is it used?
  - What is the difference between Lock and synchronized?
  - Explain tryLock() and lockInterruptibly().

### Thread Communication
- **wait(), notify(), notifyAll()**
  - Explain wait(), notify(), and notifyAll() methods.
  - Why must wait() be called in a synchronized block?
  - What is the difference between notify() and notifyAll()?
  - What happens if you call wait() without holding the lock?
  - Explain the producer-consumer problem using wait/notify.
  
- **BlockingQueue**
  - How does BlockingQueue help in thread communication?
  - What are the different BlockingQueue implementations?
  - Explain put() and take() methods.
  
- **CountDownLatch**
  - What is CountDownLatch? How does it work?
  - When would you use CountDownLatch?
  - Can CountDownLatch be reused?
  
- **CyclicBarrier**
  - What is CyclicBarrier? How does it work?
  - What is the difference between CountDownLatch and CyclicBarrier?
  - When would you use CyclicBarrier?
  
- **Semaphore**
  - What is Semaphore? How does it work?
  - Explain the concept of permits in Semaphore.
  - When would you use Semaphore?
  - What is the difference between Semaphore and Lock?

### Executor Framework
- **Executor and ExecutorService**
  - What is the Executor framework? Why was it introduced?
  - What is the difference between Executor and ExecutorService?
  - How do you shut down an ExecutorService?
  - What is the difference between shutdown() and shutdownNow()?
  
- **ThreadPoolExecutor**
  - What is ThreadPoolExecutor? How does it work?
  - Explain corePoolSize, maximumPoolSize, and queue capacity.
  - What are the different types of thread pools?
  - How do you create a custom thread pool?
  
- **ScheduledExecutorService**
  - What is ScheduledExecutorService?
  - How do you schedule tasks to run periodically?
  - What is the difference between scheduleAtFixedRate and scheduleWithFixedDelay?
  
- **ForkJoinPool**
  - What is ForkJoinPool? When is it used?
  - Explain the work-stealing algorithm.
  - What is the difference between ForkJoinPool and ThreadPoolExecutor?
  
- **CompletableFuture**
  - What is CompletableFuture? How is it different from Future?
  - How do you chain CompletableFuture operations?
  - Explain thenApply(), thenCompose(), and thenCombine().
  - How do you handle exceptions in CompletableFuture?

### Concurrent Collections
- **ConcurrentHashMap**
  - How does ConcurrentHashMap achieve thread-safety?
  - What is the difference between ConcurrentHashMap and synchronized HashMap?
  - Explain the internal structure of ConcurrentHashMap (Java 8).
  - What is the concurrency level in ConcurrentHashMap?
  
- **CopyOnWriteArrayList**
  - What is CopyOnWriteArrayList? How does it work?
  - When should you use CopyOnWriteArrayList?
  - What are the performance implications?
  
- **Thread-safe Collections**
  - What are the different thread-safe collections?
  - What is the difference between concurrent collections and synchronized wrappers?
  - When should you use concurrent collections vs synchronized wrappers?

### Advanced Concurrency
- **ThreadLocal**
  - What is ThreadLocal? How does it work?
  - When would you use ThreadLocal?
  - What are the memory leak concerns with ThreadLocal?
  - How do you clean up ThreadLocal variables?
  
- **Memory Model**
  - What is the Java Memory Model?
  - Explain happens-before relationships.
  - What is visibility in the context of concurrency?
  - Explain the concept of reordering.
  
- **Lock-free Algorithms**
  - What are lock-free algorithms?
  - How do they achieve thread-safety without locks?
  - Give examples of lock-free data structures.

## Asynchronous Programming

### CompletableFuture
- **Creating CompletableFuture**
  - How do you create a CompletableFuture?
  - What is the difference between supplyAsync() and runAsync()?
  - How do you specify an executor for CompletableFuture?
  
- **Chaining Operations**
  - How do you chain CompletableFuture operations?
  - What is the difference between thenApply() and thenCompose()?
  - Explain thenCombine() for combining futures.
  - What is the difference between thenApply() and thenAccept()?
  
- **Exception Handling**
  - How do you handle exceptions in CompletableFuture?
  - Explain handle() and exceptionally() methods.
  - What happens if an exception is not handled?
  
- **Combining Futures**
  - How do you combine multiple CompletableFutures?
  - Explain allOf() and anyOf().
  - How do you get results from allOf()?

### Reactive Programming
- **Reactive Streams**
  - What is reactive programming?
  - Explain the Reactive Streams specification.
  - What are the main components of reactive streams?
  
- **Project Reactor**
  - What is Project Reactor?
  - Explain Mono and Flux.
  - What is the difference between Mono and Flux?
  - How do you handle backpressure in Reactor?
  
- **RxJava**
  - What is RxJava?
  - Explain Observable, Single, and Maybe.
  - What are the main operators in RxJava?

### Asynchronous I/O
- **NIO**
  - What is NIO? How is it different from traditional I/O?
  - Explain Channels, Buffers, and Selectors.
  - What are the advantages of NIO?
  - When should you use NIO?

## Spring Framework

### Core Concepts
- **Dependency Injection**
  - What is Dependency Injection (DI)?
  - What are the different types of dependency injection?
  - What is the difference between constructor, setter, and field injection?
  - When should you use constructor injection vs setter injection?
  - What is Inversion of Control (IoC)?
  
- **Spring Container**
  - What is the Spring IoC container?
  - What is the difference between ApplicationContext and BeanFactory?
  - Explain the bean lifecycle in Spring.
  - What are bean scopes? Explain each scope.
  
- **Stereotype Annotations**
  - What is @Component? How is it different from @Service, @Repository, @Controller?
  - When should you use @Service vs @Component?
  - What is the difference between @Controller and @RestController?

### Configuration
- **Configuration Styles**
  - What are the different ways to configure Spring?
  - What is the difference between XML and annotation-based configuration?
  - Explain Java-based configuration with @Configuration.
  - When should you use @Bean annotation?
  
- **Property Injection**
  - How do you inject properties in Spring?
  - Explain @Value annotation.
  - What is @PropertySource?
  - How do you use external configuration files?
  
- **Profiles**
  - What are Spring profiles? How do you use them?
  - How do you activate a profile?
  - Explain @Profile annotation.
  
- **Conditional Beans**
  - What is @Conditional? When is it used?
  - How do you create custom conditions?

### AOP (Aspect-Oriented Programming)
- **AOP Concepts**
  - What is AOP? Why is it used?
  - Explain cross-cutting concerns.
  - What are join points, pointcuts, and advice?
  - What is the difference between @Before, @After, and @Around?
  
- **Proxies**
  - How does Spring AOP work internally?
  - What is the difference between JDK dynamic proxies and CGLIB?
  - When does Spring use JDK proxies vs CGLIB?

### Spring MVC
- **MVC Architecture**
  - Explain the Spring MVC architecture.
  - What is DispatcherServlet? How does it work?
  - Explain the request flow in Spring MVC.
  
- **Request Mapping**
  - What is @RequestMapping? How does it work?
  - What is the difference between @GetMapping and @PostMapping?
  - How do you handle path variables and request parameters?
  - Explain @RequestParam vs @PathVariable.
  
- **Response Handling**
  - How do you return data from a controller?
  - What is ResponseEntity?
  - Explain @ResponseBody and @RestController.
  - How do you handle different content types?

### Spring Boot
- **Auto-configuration**
  - What is Spring Boot auto-configuration?
  - How does Spring Boot auto-configuration work?
  - How do you exclude auto-configuration?
  - Explain @SpringBootApplication annotation.
  
- **Starters**
  - What are Spring Boot starters?
  - How do starters work?
  - Give examples of common starters.
  
- **Configuration Properties**
  - What is @ConfigurationProperties?
  - How do you bind properties to Java objects?
  - What is the difference between @Value and @ConfigurationProperties?
  
- **Actuator**
  - What is Spring Boot Actuator?
  - What endpoints does Actuator provide?
  - How do you customize Actuator endpoints?
  - How do you secure Actuator endpoints?

### Spring Security
- **Authentication**
  - How does Spring Security work?
  - What is UserDetailsService?
  - How do you implement custom authentication?
  - Explain password encoding in Spring Security.
  - What is JWT? How do you implement JWT in Spring Security?
  
- **Authorization**
  - What is the difference between authentication and authorization?
  - How do you implement method-level security?
  - Explain @PreAuthorize and @PostAuthorize.
  - How do you configure URL-based security?
  
- **OAuth2**
  - What is OAuth2? How does it work?
  - Explain OAuth2 roles (client, resource server, authorization server).
  - How do you implement OAuth2 in Spring Security?

### Spring Data
- **Spring Data JPA**
  - What is Spring Data JPA?
  - What is a repository? Explain CrudRepository and JpaRepository.
  - How do query methods work in Spring Data?
  - Explain @Query annotation.
  - What is the difference between JPQL and native queries?
  - How do you implement custom repository methods?
  
- **JPA Entity Management**
  - Explain @Entity, @Table, @Id, @GeneratedValue.
  - What are entity relationships? Explain @OneToMany, @ManyToOne, @ManyToMany.
  - What is the difference between EAGER and LAZY fetching?
  - Explain cascade operations.
  - What is the N+1 query problem? How do you solve it?

## Advanced Topics

### JVM Internals
- **Class Loading**
  - How does class loading work in Java?
  - What are classloaders? Explain different types.
  - What is the class loading process?
  - Explain delegation model in class loading.
  
- **Bytecode**
  - What is bytecode?
  - How does JVM execute bytecode?
  - What is JIT compilation?
  
- **Garbage Collection**
  - Explain different GC algorithms in detail.
  - What is the difference between G1, ZGC, and Shenandoah?
  - How do you tune GC for your application?
  - What are GC logs? How do you analyze them?
  
- **JVM Tuning**
  - How do you tune JVM heap size?
  - What are the important JVM parameters?
  - How do you monitor JVM performance?
  - Explain jstat, jmap, jstack tools.

### Performance Optimization
- **Profiling**
  - What is profiling? Why is it important?
  - Explain different profiling techniques.
  - What tools do you use for profiling Java applications?
  - How do you analyze heap dumps and thread dumps?
  
- **Optimization Techniques**
  - What are common performance bottlenecks in Java?
  - How do you optimize database queries?
  - Explain caching strategies.
  - What is connection pooling? How does it help?

### Build Tools
- **Maven**
  - What is Maven? How does it work?
  - Explain POM (Project Object Model).
  - What is the Maven lifecycle?
  - How do you manage dependencies in Maven?
  - Explain Maven plugins.
  
- **Gradle**
  - What is Gradle? How is it different from Maven?
  - Explain Gradle build scripts.
  - What are Gradle tasks?
  - How do you manage dependencies in Gradle?

### Testing
- **JUnit**
  - What is JUnit? How do you write tests?
  - Explain @Test, @BeforeEach, @AfterEach annotations.
  - What is the difference between @BeforeAll and @BeforeEach?
  - How do you test exceptions in JUnit?
  - Explain parameterized tests.
  
- **Mockito**
  - What is Mockito? How does it work?
  - How do you create mocks and stubs?
  - Explain @Mock and @InjectMocks.
  - What is the difference between mock() and spy()?
  - How do you verify method calls?
  
- **Integration Testing**
  - What is integration testing?
  - How do you write integration tests in Spring?
  - Explain @SpringBootTest.
  - What is TestContainers?

### Design Patterns
- **Creational Patterns**
  - Explain Singleton pattern. How do you implement it thread-safely?
  - What is Factory pattern? Give an example.
  - Explain Builder pattern and its benefits.
  - What is Prototype pattern?
  
- **Structural Patterns**
  - Explain Adapter pattern.
  - What is Decorator pattern?
  - Explain Facade pattern.
  - What is Proxy pattern? How is it used in Spring?
  
- **Behavioral Patterns**
  - Explain Observer pattern.
  - What is Strategy pattern?
  - Explain Template Method pattern.
  - What is Command pattern?

### Best Practices
- **Code Quality**
  - What are Java coding conventions?
  - Explain naming conventions in Java.
  - What is JavaDoc? How do you write good documentation?
  - What are code review best practices?
  
- **Error Handling**
  - What are best practices for exception handling?
  - How do you design exception hierarchies?
  - When should you use checked vs unchecked exceptions?
  
- **Logging**
  - What logging frameworks are available in Java?
  - Explain SLF4J and its benefits.
  - How do you configure logging levels?
  - What are logging best practices?
  
- **Security**
  - What are common security vulnerabilities in Java applications?
  - How do you prevent SQL injection?
  - Explain input validation.
  - What are best practices for handling sensitive data?

