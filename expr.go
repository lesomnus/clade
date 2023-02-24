package clade

import (
	"fmt"
	"strings"

	"github.com/lesomnus/boolal"
	"github.com/lesomnus/pl"
	"gopkg.in/yaml.v3"
)

type Pipeline pl.Pl

func (p *Pipeline) UnmarshalYAML(node *yaml.Node) error {
	expr := ""
	if err := node.Decode(&expr); err != nil {
		return err
	}

	var (
		pipeline *pl.Pl
		err      error
	)
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		pipeline, err = pl.ParseString(expr)
	} else {
		fn, _ := pl.NewFn("pass", expr)
		pipeline = pl.NewPl(fn)
	}

	if err != nil {
		return fmt.Errorf("parse pipeline: %w", err)
	}

	*p = Pipeline(*pipeline)

	return nil
}

type BoolAlgebra boolal.Expr

func (a *BoolAlgebra) UnmarshalYAML(node *yaml.Node) error {
	expr := ""
	if err := node.Decode(&expr); err != nil {
		return err
	}

	algebra, err := boolal.ParseString(expr)
	if err != nil {
		return fmt.Errorf("parse boolean algebra: %w", err)
	}

	*a = *(*BoolAlgebra)(algebra)

	return nil
}
