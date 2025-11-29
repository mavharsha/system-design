# Angular Interview Questions

## Basic Level
### Fundamentals
- **What is Angular?**
  - A TypeScript-based open-source web application framework led by the Angular Team at Google.
  - Used for building single-page client applications using HTML and TypeScript.
- **Differentiate between Angular and AngularJS.**
  - **Architecture:** Angular is component-based; AngularJS is controller-based (MVC).
  - **Language:** Angular uses TypeScript; AngularJS uses JavaScript.
  - **Mobile Support:** Angular supports mobile development; AngularJS does not.
  - **Performance:** Angular is generally faster due to AOT compilation and better change detection.
- **What are the key building blocks of an Angular application?**
  - Modules
  - Components
  - Templates
  - Metadata
  - Data Binding
  - Directives
  - Services
  - Dependency Injection

### Components & Directives
- **What is a Component?**
  - A class with the `@Component` decorator that controls a patch of screen called a view.
- **What is a Directive?**
  - Classes that add behavior to elements in the application.
  - **Types:**
    - Components (directives with a template)
    - Structural directives (change the DOM layout, e.g., `*ngIf`, `*ngFor`)
    - Attribute directives (change the appearance/behavior of an element, e.g., `ngClass`, `ngStyle`)
- **What are Lifecycle Hooks?**
  - Methods that hook into key events in a component's lifecycle (creation, change detection, destruction).
  - Examples: `ngOnInit`, `ngOnChanges`, `ngOnDestroy`.

### Data Binding
- **Explain Data Binding types.**
  - **Interpolation `{{ }}`:** Component to View (one-way).
  - **Property Binding `[ ]`:** Component to View (one-way).
  - **Event Binding `( )`:** View to Component (one-way).
  - **Two-way Binding `[( )]`:** Combines property and event binding (e.g., `[(ngModel)]`).

## Intermediate Level
### Dependency Injection & Services
- **What is Dependency Injection (DI)?**
  - A design pattern in which a class asks for dependencies from external sources rather than creating them.
- **How is a service provided?**
  - Using the `providedIn: 'root'` syntax in the `@Injectable` decorator (preferred).
  - Or in the `providers` array of a module or component.
- **What is the difference between `@Optional`, `@Self`, `@SkipSelf`, and `@Host`?**
  - Resolution modifiers that control how the DI system looks for a dependency.

### Routing
- **What are Route Guards?**
  - Interfaces that can tell the router whether to allow navigation to or from a route.
  - Types: `CanActivate`, `CanActivateChild`, `CanDeactivate`, `CanLoad`, `Resolve`.
- **What is Lazy Loading?**
  - Loading NgModules (or standalone components) on demand rather than at startup to improve initial load time.

### Forms
- **Difference between Template-driven and Reactive Forms.**
  - **Template-driven:** Asynchronous, template-heavy, logic in template, uses `ngModel`. Good for simple forms.
  - **Reactive:** Synchronous, code-heavy, logic in component class, uses `FormControl`/`FormGroup`. Better for complex, dynamic forms and testing.

## Advanced Level
### Change Detection
- **What is Change Detection?**
  - The mechanism by which Angular synchronizes the state of the application UI with the state of the data.
- **Difference between Default and OnPush strategies.**
  - **Default:** Checks every component in the tree whenever any event occurs.
  - **OnPush:** Only checks the component if:
    - An `@Input` reference changes (immutability matters).
    - An event originates from the component or its children.
    - An Observable using the async pipe emits a value.
    - Change detection is manually triggered.
- **What is Zone.js?**
  - A library that monkey-patches browser asynchronous APIs (like `setTimeout`, DOM events) to notify Angular when to run change detection.

### Performance
- **What is AOT (Ahead-of-Time) Compilation?**
  - Compiles HTML templates and TypeScript code into JavaScript code before the browser downloads and runs it.
  - Benefits: Faster rendering, smaller download size, template errors detected at build time.
- **What is JIT (Just-in-Time) Compilation?**
  - Compiles the application in the browser at runtime.
- **What are Pure Pipes?**
  - Pipes that are executed only when a pure change to the input value is detected (primitive value change or object reference change).
  - Default behavior for pipes.

### Modern Angular
- **What are Standalone Components?**
  - Components, directives, or pipes that are not part of any NgModule.
  - Simplified architecture (no `NgModule` required).
  - Introduced in Angular 14 (preview), stable in v15.
- **What are Signals?**
  - A reactive primitive for managing state and notifying consumers of changes.
  - Granular updates without relying on Zone.js.
  - **Writable Signals:** Can be updated directly (`set`, `update`).
  - **Computed Signals:** Derive values from other signals.
  - **Effects:** Run side effects when signal values change.
- **Observable vs Signal?**
  - **Observable:** Push-based stream of values over time (RxJS). Great for events and complex async flows.
  - **Signal:** Holds a current value, pull-based (mostly), tracks dependencies automatically. Great for synchronous state and template rendering.

