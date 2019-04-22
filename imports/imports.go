package imports

import (
	"fmt"

	. "github.com/philandstuff/dhall-golang/ast"
	"github.com/philandstuff/dhall-golang/parser"
)

func ResolveStringAsExpr(name, content string) (Expr, error) {
	expr, err := parser.Parse(name, []byte(content))
	return expr.(Expr), err
}

func Load(e Expr, ancestors ...Resolvable) (Expr, error) {
	switch e := e.(type) {
	case Embed:
		i := Import(e)
		here := i.ImportHashed
		for _, ancestor := range ancestors {
			if ancestor == here {
				return nil, fmt.Errorf("Detected import cycle in %s", ancestor)
			}
		}
		imports := append(ancestors, here)
		content, err := i.Resolve()
		if err != nil {
			return nil, err
		}
		if i.ImportMode == RawText {
			return TextLit{Suffix: content}, nil
		} else {
			// dynamicExpr may contain more imports
			dynamicExpr, err := ResolveStringAsExpr(i.Name(), content)
			if err != nil {
				return nil, err
			}

			// recursively load any more imports
			expr, err := Load(dynamicExpr, imports...)
			if err != nil {
				return nil, err
			}

			// ensure that expr typechecks in empty context
			_, err = expr.TypeWith(EmptyContext())
			if err != nil {
				return nil, err
			}
			return expr, nil
		}
	case *LambdaExpr:
		resolvedType, err := Load(e.Type)
		if err != nil {
			return nil, err
		}
		resolvedBody, err := Load(e.Body)
		if err != nil {
			return nil, err
		}
		return &LambdaExpr{
			Label: e.Label,
			Type:  resolvedType,
			Body:  resolvedBody,
		}, nil
	case *Pi:
		resolvedType, err := Load(e.Type)
		if err != nil {
			return nil, err
		}
		resolvedBody, err := Load(e.Body)
		if err != nil {
			return nil, err
		}
		return &Pi{
			Label: e.Label,
			Type:  resolvedType,
			Body:  resolvedBody,
		}, nil
	case *App:
		resolvedFn, err := Load(e.Fn)
		if err != nil {
			return nil, err
		}
		resolvedArg, err := Load(e.Arg)
		if err != nil {
			return nil, err
		}
		return &App{
			Fn:  resolvedFn,
			Arg: resolvedArg,
		}, nil
	case Let:
		newBindings := make([]Binding, len(e.Bindings))
		for i, binding := range e.Bindings {
			var err error
			newBindings[i].Variable = binding.Variable
			if binding.Annotation != nil {
				newBindings[i].Annotation, err = Load(binding.Annotation)
				if err != nil {
					return nil, err
				}
			}
			newBindings[i].Value, err = Load(binding.Value)
			if err != nil {
				return nil, err
			}
		}
		resolvedBody, err := Load(e.Body)
		if err != nil {
			return nil, err
		}
		return Let{Bindings: newBindings, Body: resolvedBody}, nil
	case Annot:
		resolvedExpr, err := Load(e.Expr)
		if err != nil {
			return nil, err
		}
		resolvedAnnotation, err := Load(e.Annotation)
		if err != nil {
			return nil, err
		}
		return Annot{Expr: resolvedExpr, Annotation: resolvedAnnotation}, nil
	case TextLit:
		newTextLit := TextLit{make(Chunks, len(e.Chunks)), e.Suffix}
		for i, chunk := range e.Chunks {
			newTextLit.Chunks[i].Prefix = chunk.Prefix
			resolvedExpr, err := Load(chunk.Expr)
			if err != nil {
				return nil, err
			}
			newTextLit.Chunks[i].Expr = resolvedExpr
		}
		return newTextLit, nil
	case BoolIf:
		resolvedCond, err := Load(e.Cond)
		if err != nil {
			return nil, err
		}
		resolvedT, err := Load(e.T)
		if err != nil {
			return nil, err
		}
		resolvedF, err := Load(e.F)
		if err != nil {
			return nil, err
		}
		return BoolIf{
			Cond: resolvedCond,
			T:    resolvedT,
			F:    resolvedF,
		}, nil
	case NaturalPlus:
		resolvedL, err := Load(e.L)
		if err != nil {
			return nil, err
		}
		resolvedR, err := Load(e.R)
		if err != nil {
			return nil, err
		}
		return NaturalPlus{L: resolvedL, R: resolvedR}, nil
	case NaturalTimes:
		resolvedL, err := Load(e.L)
		if err != nil {
			return nil, err
		}
		resolvedR, err := Load(e.R)
		if err != nil {
			return nil, err
		}
		return NaturalTimes{L: resolvedL, R: resolvedR}, nil
	case EmptyList:
		resolvedType, err := Load(e.Type)
		if err != nil {
			return nil, err
		}
		return EmptyList{Type: resolvedType}, nil
	case NonEmptyList:
		newList := make([]Expr, len(e))
		for i, item := range e {
			var err error
			newList[i], err = Load(item)
			if err != nil {
				return nil, err
			}
		}
		return NonEmptyList(newList), nil
	case Record:
		newRecord := make(map[string]Expr, len(e))
		for k, v := range e {
			var err error
			newRecord[k], err = Load(v)
			if err != nil {
				return nil, err
			}
		}
		return Record(newRecord), nil
	case RecordLit:
		newRecord := make(map[string]Expr, len(e))
		for k, v := range e {
			var err error
			newRecord[k], err = Load(v)
			if err != nil {
				return nil, err
			}
		}
		return RecordLit(newRecord), nil
	case Field:
		newRecord, err := Load(e.Record)
		if err != nil {
			return nil, err
		}
		return Field{Record: newRecord, FieldName: e.FieldName}, nil
	default:
		// Const, NaturalLit, etc
		return e, nil
	}
}