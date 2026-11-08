# Technical Specification: go-fsm

## 1. Domain & Purpose
`go-fsm` is a thread-safe, generic Finite State Machine (FSM) engine for Go. It orchestrates stateful business lifecycles — such as e-commerce checkout funnels, order status workflows, payment gateway state progressions, and user account activations — while strictly enforcing valid transitions and guard conditions.

## 2. Architecture & Design Principles
- **Generic State & Event Types**: Both state `S` and event `E` are generic comparable types (supporting custom string enums, integers, or custom types) with zero reflection.
- **Conditional Guard Predicates**: Transitions can declare boolean guard closures (`func() bool`) to dynamically permit or reject state changes.
- **Hierarchical Lifecycle Hooks**: Supports granular state entry (`OnEnter`), state exit (`OnExit`), and global transition listeners (`OnTransition`).
- **Thread Safety**: Complete synchronization via `sync.RWMutex`.
- **Zero Heavy Dependencies**: Pure Go standard library (`sync`, `fmt`).

## 3. API Surface
- `type FSM[S comparable, E comparable] struct`
- `func New[S comparable, E comparable](initial S) *FSM[S, E]`
- `func (f *FSM[S, E]) AddTransition(from S, event E, to S, guard func() bool) *FSM[S, E]`
- `func (f *FSM[S, E]) OnEnter(state S, hook EnterHook[S, E]) *FSM[S, E]`
- `func (f *FSM[S, E]) OnExit(state S, hook ExitHook[S, E]) *FSM[S, E]`
- `func (f *FSM[S, E]) OnTransition(hook TransitionHook[S, E]) *FSM[S, E]`
- `func (f *FSM[S, E]) Current() S`
- `func (f *FSM[S, E]) Can(event E) bool`
- `func (f *FSM[S, E]) Trigger(event E) error`
- `func (f *FSM[S, E]) AvailableEvents() []E`

## 4. Error Handling
- `ErrInvalidTransition`: Transition not defined for current state and event.
- `ErrGuardFailed`: Transition guard condition returned false.

## 5. Verification & Testing
- Unit tests verify order lifecycle state progression (`Pending -> Paid -> Shipped -> Delivered`), illegal jump rejections, guard predicate evaluation, and hook invocation ordering (`Exit -> Mutate -> Enter -> Transition`).
