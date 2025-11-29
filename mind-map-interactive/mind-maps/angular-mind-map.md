# Angular

## Core Concepts
### Architecture
- Modules (NgModule)
  - Declarations
  - Imports
  - Exports
  - Providers
  - Bootstrap
- Standalone Components (v14+)
  - No NgModule required
  - Imports directly in component
  - Bootstrapped via bootstrapApplication
- Components
- Templates
- Metadata
- Data Binding
- Directives
- Services
- Dependency Injection

### Components & Templates
- **Lifecycle Hooks**
  - ngOnChanges
  - ngOnInit
  - ngDoCheck
  - ngAfterContentInit
  - ngAfterContentChecked
  - ngAfterViewInit
  - ngAfterViewChecked
  - ngOnDestroy
- **Data Binding**
  - Interpolation {{ }}
  - Property Binding [ ]
  - Event Binding ( )
  - Two-way Binding [( )]
- **Directives**
  - Structural (*ngIf, *ngFor, *ngSwitch)
  - Attribute (ngClass, ngStyle)
  - Custom Directives
- **Pipes**
  - Built-in (Date, UpperCase, LowerCase, Currency, etc.)
  - Pure vs Impure
  - Custom Pipes
- **Custom Components**
  - @Input() / @Output()
  - Content Projection (ng-content)
  - ViewChild / ContentChild
  - HostBinding / HostListener
  - Styles encapsulation (Emulated, ShadowDom, None)

### Signals (v16+)
- **Writable Signals**
  - signal()
  - set()
  - update()
  - mutate() (deprecated/removed in favor of update)
- **Computed Signals**
  - computed()
  - Memoized
  - Auto-dependency tracking
- **Effects**
  - effect()
  - Side effects (logging, manual DOM sync)
  - Injection context required
- **RxJS Interop**
  - toSignal()
  - toObservable()

### Dependency Injection
- Providers (useClass, useValue, useFactory, useExisting)
- Hierarchical Injection
- Resolution Modifiers (@Optional, @SkipSelf, @Self, @Host)
- providedIn: 'root'

## Routing & Navigation
- RouterModule
- Routes Configuration
- RouterOutlet
- RouterLink
- ActiveRoute
- **Guards**
  - CanActivate
  - CanActivateChild
  - CanDeactivate
  - CanLoad
  - Resolve
- Lazy Loading

## Forms
### Template-Driven
- ngModel
- NgForm
- Simple validation

### Reactive Forms
- FormControl
- FormGroup
- FormArray
- FormBuilder
- Custom Validators
- Async Validators

## RxJS & Observables
- Observable vs Promise
- Subscription
- **Operators**
  - Creation (of, from)
  - Transformation (map, switchMap, mergeMap, concatMap)
  - Filtering (filter, take, debounceTime)
  - Combination (combineLatest, merge, forkJoin)
- Subjects (Subject, BehaviorSubject, ReplaySubject, AsyncSubject)

## State Management
- Services with Observables/Subjects
- **Signals** (Modern Angular)
  - Writable Signals
  - Computed Signals
  - Effects
- NgRx (Redux pattern)
  - Store
  - Actions
  - Reducers
  - Selectors
  - Effects

## Advanced Topics
- Change Detection
  - Default
  - OnPush
  - Zone.js
- Dynamic Components
- Content Projection (ng-content)
- ViewChild / ContentChild
- HostBinding / HostListener
- Standalone Components
- Server-Side Rendering (SSR) / Angular Universal

## Interview Questions
### Basics
- What is Angular?
- Differences between Angular and AngularJS?
- What are Lifecycle Hooks?
- Explain the Digest Cycle (legacy) vs Change Detection.
- What is a Directive?
- What is AOT vs JIT compilation?

### Intermediate
- What is the difference between Constructor and ngOnInit?
- How does Dependency Injection work in Angular?
- Difference between Observables and Promises?
- What are pure and impure pipes?
- Explain Lazy Loading.
- What is the use of codelyzer?

### Advanced
- How does Change Detection work?
- What is Zone.js?
- Explain NgRx flow.
- Difference between mergeMap, switchMap, concatMap, exhaustMap.
- How to optimize Angular application performance?
- What are Standalone Components?
- How do Signals differ from Observables?

