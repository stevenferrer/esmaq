package gen

import (
	"fmt"
	"io"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"

	"github.com/sf9v/esmaq"
)

const pkgPath = "github.com/sf9v/esmaq"

// Schema is the state machine schema
type Schema struct {
	// Name is the state machine name
	Name,
	// Pkg is package name
	Pkg string
	// States is the states config
	States []State
}

// State is the state configuration
type State struct {
	From        esmaq.StateType
	Transitions []Transition
}

// Transition is the transition configuration
type Transition struct {
	To       esmaq.StateType
	Event    esmaq.EventType
	Callback Callback
}

// Callback is a callback function signature
type Callback struct {
	Ins  []Param
	Outs []Param
}

// Param is a callback parameter
type Param struct {
	ID string
	V  interface{}
}

// Generate generates the state machine
func Generate(schema Schema, out io.Writer) error {
	// default name
	fsmName := "StateMachine"
	if len(schema.Name) > 0 {
		fsmName = schema.Name
	}

	// camelize name
	fsmName = toCamel(fsmName)

	//default package name
	pkg := "main"
	if len(schema.Pkg) > 0 {
		pkg = schema.Pkg
	}

	f := jen.NewFile(pkg)
	f.PackageComment("Code generated by esmaq, DO NOT EDIT.")

	rcvr := "sm"
	rcvrType := "*" + fsmName

	states := []State{}
	states = append(states, schema.States...)

	// helper method for checking if
	// state is already in collection
	isIn := func(s esmaq.StateType) bool {
		for _, state := range states {
			if s == state.From {
				return true
			}
		}
		return false
	}

	for _, stateCfg := range schema.States {
		for _, trsnCfg := range stateCfg.Transitions {
			if isIn(trsnCfg.To) {
				continue
			}
			states = append(states, State{From: trsnCfg.To})
		}
	}

	// state types
	f.Comment("State is the state type").Line().
		Type().Id("State").String().Line()
	f.Comment("String implements Stringer for State").Line().
		Func().Params(jen.Id("s").Id("State")).Id("String").
		Params().Params(jen.String()).
		Block(jen.Return(jen.String().Call(jen.Id("s"))))
	f.Comment("List of state types").Line().
		Const().DefsFunc(func(g *jen.Group) {
		for _, state := range states {
			s := string(state.From)
			g.Id(stName(state.From)).Id("State").Op("=").Lit(s)
		}
	}).Line()

	// event types
	f.Comment("Event is the event type").Line().
		Type().Id("Event").String()
	f.Comment("String implements Stringer for Event").Line().
		Func().Params(jen.Id("e").Id("Event")).Id("String").
		Params().Params(jen.String()).
		Block(jen.Return(jen.String().Call(jen.Id("e"))))
	f.Comment("List of event types").Line().
		Const().DefsFunc(func(g *jen.Group) {
		for _, state := range states {
			for _, trsn := range state.Transitions {
				e := string(trsn.Event)
				g.Id(etName(trsn.Event)).Id("Event").Op("=").Lit(e)
			}
		}
	})

	f.Comment("ctxKey is a context key").Line().
		Type().Id("ctxKey").Int()
	f.Comment("List of context keys").Line().
		Const().DefsFunc(func(g *jen.Group) {
		g.Id("fromKey").Id("ctxKey").Op("=").Id("iota")
		g.Id("toKey")
	})

	// callback function arguments
	cbFnArgs := []jen.Code{}
	// transition methods
	methods := []jen.Code{}
	for _, state := range states {
		for _, trsn := range state.Transitions {
			fnName := toCamel(string(trsn.Event))

			// input args
			ins := []jen.Code{jen.Id("ctx").Qual("context", "Context")}
			// input arg identifiers
			inIDs := []jen.Code{jen.Id("ctx")}
			for _, in := range trsn.Callback.Ins {
				ins = append(ins, getParamC(in.ID, in.V))
				inIDs = append(inIDs, jen.Id(in.ID))
			}

			// output args
			outs := []jen.Code{}
			// output arg identifiers
			outIDs := []jen.Code{}
			// return args when error happened
			errRets := []jen.Code{}

			for _, out := range trsn.Callback.Outs {
				outs = append(outs, getParamC(out.ID, out.V))
				outIDs = append(outIDs, jen.Id(out.ID))
				errRets = append(errRets, getZeroValC(out.V))
			}

			// return args when no error happened
			okRets := append(cloneC(outIDs), jen.Nil())

			// add err in as last arg
			outs = append(outs, jen.Id("err").Error())
			outIDs = append(outIDs, jen.Id("err"))

			cbName := fnName
			cbFnArgs = append(cbFnArgs, jen.Id(cbName).Func().
				Params(ins...).Params(outs...))

			// transition methods
			comment := fmt.Sprintf("%s is a transition method for %s", fnName, etName(trsn.Event))
			method := jen.Comment(comment).Line().
				Func().Params(jen.Id(rcvr).Id(rcvrType)).Id(fnName).
				Params(ins...).Params(outs...).
				BlockFunc(func(g *jen.Group) {
					g.List(jen.Id("from"), jen.Id("ok")).
						Op(":=").Id("fromCtx").Call(jen.Id("ctx"))
					g.If(jen.Op("!").Id("ok")).
						Block(jen.Return(append(cloneC(errRets), jen.Qual("errors", "New").
							Call(jen.Lit(`"from" is not set in context`)))...)).Line()

					g.List(jen.Id("fromst"), jen.Id("err")).Op(":=").Id(rcvr).
						Dot("core").Dot("GetState").
						Call(jen.Id("castst").Call(jen.Id("from")))
					g.If(jen.Err().Op("!=").Nil()).
						Block(jen.Return(append(cloneC(errRets), jen.Id("err"))...)).Line()

					g.List(jen.Id("tost"), jen.Id("err")).
						Op(":=").Id(rcvr).Dot("core").Dot("Transition").
						Call(jen.Id("castst").Call(jen.Id("from")),
							jen.Id("castet").Call(jen.Id(etName(trsn.Event))))
					g.If(jen.Err().Op("!=").Nil()).
						Block(jen.Return(append(cloneC(errRets), jen.Id("err"))...)).Line()

					g.Comment(`inject "to" in context`)
					g.Id("ctx").Op("=").Id("ctxWtTo").
						Call(jen.Id("ctx"), jen.Id(stName(trsn.To))).Line()

					g.If(jen.Id(rcvr).Dot("callbacks").Op("!=").Nil()).Op("&&").
						Id(rcvr).Dot("callbacks").Dot(cbName).Op("!=").Nil().
						BlockFunc(func(g *jen.Group) {
							g.List(outIDs...).Op("=").Id(rcvr).Dot("callbacks").
								Dot(cbName).Call(inIDs...)
							g.If(jen.Err().Op("!=").Nil()).
								Block(jen.Return(append(cloneC(errRets), jen.Id("err"))...))
						}).Line()

					g.If(jen.Id("fromst").
						Dot("Actions").
						Dot("OnExit").
						Op("!=").Nil()).
						BlockFunc(func(g *jen.Group) {
							g.Err().Op("=").Id("fromst").Dot("Actions").
								Dot("OnExit").Call(jen.Id("ctx"))
							g.If(jen.Err().Op("!=").Nil()).
								Block(jen.Return(append(cloneC(errRets), jen.Id("err"))...))
						}).
						Line()

					g.If(jen.Id("tost").Dot("Actions").Dot("OnEnter").Op("!=").Nil()).
						Block(jen.Err().Op("=").Id("tost").Dot("Actions").Dot("OnEnter").
							Call(jen.Id("ctx")), jen.If(jen.Err().Op("!=").Nil()).
							Block(jen.Return(append(cloneC(errRets), jen.Id("err"))...))).Line()

					g.Return(okRets...)
				})

			methods = append(methods, method)
		}
	}

	// state machine type definition
	f.Comment(fmt.Sprintf("%s is a state machine", fsmName)).Line().
		Type().Id(fsmName).Struct(
		jen.Id("core").Op("*").Qual(pkgPath, "Core"),
		jen.Id("callbacks").Op("*").Id("Callbacks"),
	).Line()

	// callback and actions type definition
	f.Comment("Callbacks defines the state machine callbacks").Line().
		Type().Id("Callbacks").Struct(cbFnArgs...).Line()
	f.Comment("Actions defines the state machine actions").Line().
		Type().Id("Actions").StructFunc(func(g *jen.Group) {
		for _, state := range states {
			g.Id(toCamel(string(state.From))).Qual(pkgPath, "Actions")
		}
	})

	// transition methods
	for _, m := range methods {
		f.Add(m).Line()
	}

	// context helpers
	f.Comment("CtxWtFrom injects `from` state to context").Line().
		Func().Id("CtxWtFrom").
		Params(jen.Id("ctx").Qual("context", "Context"),
			jen.Id("from").Id("State"),
		).Params(jen.Qual("context", "Context")).
		Block(jen.Return(jen.Qual("context", "WithValue").
			Call(jen.Id("ctx"), jen.Id("fromKey"), jen.Id("from")))).Line()

	f.Comment("ctxWtTo injects 'to' state to context").Line().
		Func().Id("ctxWtTo").
		Params(jen.Id("ctx").Qual("context", "Context"),
			jen.Id("to").Id("State")).
		Params(jen.Qual("context", "Context")).
		Block(jen.Return(jen.Qual("context", "WithValue").
			Call(jen.Id("ctx"), jen.Id("toKey"), jen.Id("to")))).
		Line()

	f.Comment("fromCtx retrieves 'from' state from context").Line().
		Func().Id("fromCtx").
		Params(jen.Id("ctx").Qual("context", "Context")).
		Params(jen.Id("State"), jen.Bool()).
		Block(jen.List(jen.Id("from"), jen.Id("ok")).Op(":=").Id("ctx").
			Dot("Value").Call(jen.Id("fromKey")).Assert(jen.Id("State")),
			jen.Return(jen.Id("from"), jen.Id("ok"))).Line()

	f.Comment("ToCtx retrieves 'to' state from context").Line().
		Func().Id("ToCtx").
		Params(jen.Id("ctx").Qual("context", "Context")).
		Params(jen.Id("State"), jen.Bool()).
		Block(jen.List(jen.Id("to"), jen.Id("ok")).Op(":=").Id("ctx").
			Dot("Value").Call(jen.Id("toKey")).Assert(jen.Id("State")),
			jen.Return(jen.Id("to"), jen.Id("ok"))).Line()

	// state machine factory
	factory := "New" + toCamel(fsmName)
	f.Comment(fmt.Sprintf("%s is a factory for state machine %s",
		factory, fsmName))
	f.Func().Id(factory).Params(jen.Id("callbacks").Op("*").Id("Callbacks"),
		jen.Id("actions").Op("*").Id("Actions")).
		Params(jen.Id("*" + fsmName)).Block(jen.Return(jen.Op("&").Id(fsmName).
		Block(jen.Id("callbacks").Op(":").Id("callbacks").Op(","),
			jen.Id("core").Op(":").Qual(pkgPath, "NewCore").Params(
				jen.Op("[]").Qual(pkgPath, "StateConfig").
					BlockFunc(func(g *jen.Group) {
						for _, state := range states {
							g.BlockFunc(func(g *jen.Group) {
								g.Id("From").Op(":").Id("castst").
									Call(jen.Id(stName(state.From))).Op(",")
								g.Id("Actions").Op(":").Id("actions").
									Dot(toCamel(string(state.From))).Op(",")
								g.Id("Transitions").Op(":").Op("[]").
									Qual(pkgPath, "TransitionConfig").
									BlockFunc(func(g *jen.Group) {
										for _, trsn := range state.Transitions {
											g.BlockFunc(func(g *jen.Group) {
												g.Id("Event").Op(":").Id("castet").
													Call(jen.Id(etName(trsn.Event))).Op(",")
												g.Id("To").Op(":").Id("castst").
													Call(jen.Id(stName(trsn.To))).Op(",")
											}).Op(",")
										}
									}).Op(",")
							}).Op(",")
						}
					})).
				Op(",")))).Line()

	// cast helpers
	f.Comment("castst casts State to esmaq.StateType").Line().
		Func().Id("castst").Params(jen.Id("s").Id("State")).
		Params(jen.Qual(pkgPath, "StateType")).
		Block(jen.Return(jen.Qual(pkgPath, "StateType").
			Call(jen.Id("s")))).Line()
	f.Comment("castet casts Event to esmaq.EventType").Line().
		Func().Id("castet").Params(jen.Id("e").Id("Event")).
		Params(jen.Qual(pkgPath, "EventType")).
		Block(jen.Return(jen.Qual(pkgPath, "EventType").
			Call(jen.Id("e"))))

	return f.Render(out)
}

func toCamel(s string) string {
	return strcase.ToCamel(s)
}

func cloneC(c1 []jen.Code) []jen.Code {
	c2 := []jen.Code{}
	return append(c2, c1...)
}

func stName(s esmaq.StateType) string {
	return toCamel("state_" + string(s))
}

func etName(e esmaq.EventType) string {
	return toCamel("event_" + string(e))
}
