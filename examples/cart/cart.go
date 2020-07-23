// Code generated by esmaq, DO NOT EDIT.
package main

import (
	"context"
	"errors"
	esmaq "github.com/stevenferrer/esmaq"
)

// State is the state type
type State esmaq.StateType

// String implements Stringer for State
func (s State) String() string {
	return string(s)
}

// List of state types
const (
	StateNew        State = "new"
	StateFinalizing State = "finalizing"
	StateSubmitted  State = "submitted"
	StateCancelled  State = "cancelled"
)

// Event is the event type
type Event esmaq.EventType

// String implements Stringer for Event
func (e Event) String() string {
	return string(e)
}

// List of event types
const (
	EventCheckout Event = "checkout"
	EventSubmit   Event = "submit"
	EventModify   Event = "modify"
	EventCancel   Event = "cancel"
)

// ctxKey is a context key
type ctxKey int

// List of context keys
const (
	fromKey ctxKey = iota
	toKey
)

// Cart is a state machine
type Cart struct {
	core      *esmaq.Core
	callbacks *Callbacks
}

// Callbacks defines the state machine callbacks
type Callbacks struct {
	Checkout func(ctx context.Context, cartID int64) (err error)
	Submit   func(ctx context.Context, cartID int64) (orderID int64, err error)
	Modify   func(ctx context.Context, cartID int64) (err error)
	Cancel   func(ctx context.Context, cartID int64) (err error)
}

// Actions defines the state machine actions
type Actions struct {
	New        esmaq.Actions
	Finalizing esmaq.Actions
	Submitted  esmaq.Actions
	Cancelled  esmaq.Actions
}

// Checkout is a transition method for EventCheckout
func (sm *Cart) Checkout(ctx context.Context, cartID int64) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" is not set in context")
	}

	fromst, err := sm.core.GetState(castst(from))
	if err != nil {
		return err
	}

	tost, err := sm.core.Transition(castst(from), castet(EventCheckout))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateFinalizing)

	if sm.callbacks != nil && sm.callbacks.Checkout != nil {
		err = sm.callbacks.Checkout(ctx, cartID)
		if err != nil {
			return err
		}
	}

	if fromst.Actions.OnExit != nil {
		err = fromst.Actions.OnExit(ctx)
		if err != nil {
			return err
		}
	}

	if tost.Actions.OnEnter != nil {
		err = tost.Actions.OnEnter(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Submit is a transition method for EventSubmit
func (sm *Cart) Submit(ctx context.Context, cartID int64) (orderID int64, err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return 0, errors.New("\"from\" is not set in context")
	}

	fromst, err := sm.core.GetState(castst(from))
	if err != nil {
		return 0, err
	}

	tost, err := sm.core.Transition(castst(from), castet(EventSubmit))
	if err != nil {
		return 0, err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateSubmitted)

	if sm.callbacks != nil && sm.callbacks.Submit != nil {
		orderID, err = sm.callbacks.Submit(ctx, cartID)
		if err != nil {
			return 0, err
		}
	}

	if fromst.Actions.OnExit != nil {
		err = fromst.Actions.OnExit(ctx)
		if err != nil {
			return 0, err
		}
	}

	if tost.Actions.OnEnter != nil {
		err = tost.Actions.OnEnter(ctx)
		if err != nil {
			return 0, err
		}
	}

	return orderID, nil
}

// Modify is a transition method for EventModify
func (sm *Cart) Modify(ctx context.Context, cartID int64) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" is not set in context")
	}

	fromst, err := sm.core.GetState(castst(from))
	if err != nil {
		return err
	}

	tost, err := sm.core.Transition(castst(from), castet(EventModify))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateNew)

	if sm.callbacks != nil && sm.callbacks.Modify != nil {
		err = sm.callbacks.Modify(ctx, cartID)
		if err != nil {
			return err
		}
	}

	if fromst.Actions.OnExit != nil {
		err = fromst.Actions.OnExit(ctx)
		if err != nil {
			return err
		}
	}

	if tost.Actions.OnEnter != nil {
		err = tost.Actions.OnEnter(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Cancel is a transition method for EventCancel
func (sm *Cart) Cancel(ctx context.Context, cartID int64) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" is not set in context")
	}

	fromst, err := sm.core.GetState(castst(from))
	if err != nil {
		return err
	}

	tost, err := sm.core.Transition(castst(from), castet(EventCancel))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateCancelled)

	if sm.callbacks != nil && sm.callbacks.Cancel != nil {
		err = sm.callbacks.Cancel(ctx, cartID)
		if err != nil {
			return err
		}
	}

	if fromst.Actions.OnExit != nil {
		err = fromst.Actions.OnExit(ctx)
		if err != nil {
			return err
		}
	}

	if tost.Actions.OnEnter != nil {
		err = tost.Actions.OnEnter(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// CtxWtFrom injects `from` state to context
func CtxWtFrom(ctx context.Context, from State) context.Context {
	return context.WithValue(ctx, fromKey, from)
}

// ctxWtTo injects 'to' state to context
func ctxWtTo(ctx context.Context, to State) context.Context {
	return context.WithValue(ctx, toKey, to)
}

// fromCtx retrieves 'from' state from context
func fromCtx(ctx context.Context) (State, bool) {
	from, ok := ctx.Value(fromKey).(State)
	return from, ok
}

// ToCtx retrieves 'to' state from context
func ToCtx(ctx context.Context) (State, bool) {
	to, ok := ctx.Value(toKey).(State)
	return to, ok
}

// NewCart is a factory for state machine Cart
func NewCart(callbacks *Callbacks, actions *Actions) *Cart {
	return &Cart{
		callbacks: callbacks,
		core: esmaq.NewCore([]esmaq.StateConfig{
			{
				From:    castst(StateNew),
				Actions: actions.New,
				Transitions: []esmaq.TransitionConfig{
					{
						Event: castet(EventCheckout),
						To:    castst(StateFinalizing),
					},
				},
			},
			{
				From:    castst(StateFinalizing),
				Actions: actions.Finalizing,
				Transitions: []esmaq.TransitionConfig{
					{
						Event: castet(EventSubmit),
						To:    castst(StateSubmitted),
					},
					{
						Event: castet(EventModify),
						To:    castst(StateNew),
					},
					{
						Event: castet(EventCancel),
						To:    castst(StateCancelled),
					},
				},
			},
			{
				From:        castst(StateSubmitted),
				Actions:     actions.Submitted,
				Transitions: []esmaq.TransitionConfig{},
			},
			{
				From:        castst(StateCancelled),
				Actions:     actions.Cancelled,
				Transitions: []esmaq.TransitionConfig{},
			},
		}),
	}
}

// castst casts State to esmaq.StateType
func castst(s State) esmaq.StateType {
	return esmaq.StateType(s)
}

// castet casts Event to esmaq.EventType
func castet(e Event) esmaq.EventType {
	return esmaq.EventType(e)
}