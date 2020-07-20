// Code generated by esmaq, DO NOT EDIT.
package matter

import (
	"context"
	"errors"
	esmaq "github.com/stevenferrer/esmaq"
)

type State esmaq.State

const (
	StateSolid  State = "solid"
	StateLiquid State = "liquid"
	StateGas    State = "gas"
)

type Event esmaq.Event

const (
	EventMelt     Event = "melt"
	EventFreeze   Event = "freeze"
	EventVaporize Event = "vaporize"
	EventCondense Event = "condense"
)

type Matter struct {
	core          *esmaq.Core
	eventHandlers *EventHandlers
}

type EventHandlers struct {
	Melt     *MeltEventHandlers
	Freeze   *FreezeEventHandlers
	Vaporize *VaporizeEventHandlers
	Condense *CondenseEventHandlers
}

type MeltEventHandlers struct {
	OnTransition func(ctx context.Context) (err error)
	OnEnter      func(context.Context) error
}

type FreezeEventHandlers struct {
	OnTransition func(ctx context.Context) (err error)
	OnEnter      func(context.Context) error
}

type VaporizeEventHandlers struct {
	OnTransition func(ctx context.Context) (err error)
	OnEnter      func(context.Context) error
}

type CondenseEventHandlers struct {
	OnTransition func(ctx context.Context) (err error)
	OnEnter      func(context.Context) error
}

func (sm *Matter) Melt(ctx context.Context) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" state not set in context")
	}

	// see transition is allowed
	err = sm.core.CanTransition(esmaq.Event(EventMelt), esmaq.State(from))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateLiquid)

	err = sm.eventHandlers.Melt.OnTransition(ctx)
	if err != nil {
		return err
	}

	err = sm.eventHandlers.Melt.OnEnter(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (sm *Matter) Freeze(ctx context.Context) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" state not set in context")
	}

	// see transition is allowed
	err = sm.core.CanTransition(esmaq.Event(EventFreeze), esmaq.State(from))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateSolid)

	err = sm.eventHandlers.Freeze.OnTransition(ctx)
	if err != nil {
		return err
	}

	err = sm.eventHandlers.Freeze.OnEnter(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (sm *Matter) Vaporize(ctx context.Context) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" state not set in context")
	}

	// see transition is allowed
	err = sm.core.CanTransition(esmaq.Event(EventVaporize), esmaq.State(from))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateGas)

	err = sm.eventHandlers.Vaporize.OnTransition(ctx)
	if err != nil {
		return err
	}

	err = sm.eventHandlers.Vaporize.OnEnter(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (sm *Matter) Condense(ctx context.Context) (err error) {
	from, ok := fromCtx(ctx)
	if !ok {
		return errors.New("\"from\" state not set in context")
	}

	// see transition is allowed
	err = sm.core.CanTransition(esmaq.Event(EventCondense), esmaq.State(from))
	if err != nil {
		return err
	}

	// inject "to" in context
	ctx = ctxWtTo(ctx, StateLiquid)

	err = sm.eventHandlers.Condense.OnTransition(ctx)
	if err != nil {
		return err
	}

	err = sm.eventHandlers.Condense.OnEnter(ctx)
	if err != nil {
		return err
	}

	return nil
}

type ctxKey int

const (
	fromKey ctxKey = iota
	toKey
)

func CtxWtFrom(ctx context.Context, from State) context.Context {
	return context.WithValue(ctx, fromKey, from)
}

func ctxWtTo(ctx context.Context, to State) context.Context {
	return context.WithValue(ctx, toKey, to)
}

func fromCtx(ctx context.Context) (State, bool) {
	from, ok := ctx.Value(fromKey).(State)
	return from, ok
}

func ToCtx(ctx context.Context) (State, bool) {
	to, ok := ctx.Value(toKey).(State)
	return to, ok
}

func NewMatter(eventHandlers *EventHandlers) *Matter {
	stateConfigs := []esmaq.StateConfig{
		{
			From: esmaq.State(StateSolid),
			Transitions: []esmaq.Transitions{
				{
					Event: esmaq.Event(EventMelt),
					To:    esmaq.State(StateLiquid),
				},
			},
		},
		{
			From: esmaq.State(StateLiquid),
			Transitions: []esmaq.Transitions{
				{
					Event: esmaq.Event(EventFreeze),
					To:    esmaq.State(StateSolid),
				},
				{
					Event: esmaq.Event(EventVaporize),
					To:    esmaq.State(StateGas),
				},
			},
		},
		{
			From: esmaq.State(StateGas),
			Transitions: []esmaq.Transitions{
				{
					Event: esmaq.Event(EventCondense),
					To:    esmaq.State(StateLiquid),
				},
			},
		},
	}

	matter := &Matter{
		core:          esmaq.NewCore(stateConfigs),
		eventHandlers: eventHandlers,
	}

	return matter
}