package storage

import (
	"fmt"
)

type QueryNode interface {
	Evaluate(entity Entity) (bool, error)
	ToString() string
}

type BinaryOp struct {
	Left, Right QueryNode
	Operator    string // AND OR
}

func (c *BinaryOp) Evaluate(entity Entity) (bool, error) {
	lv, le := c.Left.Evaluate(entity)
	if le != nil {
		return false, le
	}

	rv, re := c.Right.Evaluate(entity)
	if re != nil {
		return false, re
	}

	switch c.Operator {
	case "OR":
		return lv || rv, nil
	case "AND":
		return lv && rv, nil
	}

	return false, fmt.Errorf("no operator %s", c.Operator)
}

func (c *BinaryOp) ToString() string {
	return fmt.Sprintf("(%s) %s (%s)", c.Left.ToString(), c.Operator, c.Right.ToString())
}

type Condition struct {
	Field    string
	Operator string // = > <  etc.
	Value    interface{}
}

func (c *Condition) Evaluate(entity Entity) (bool, error) {
	v, ok := entity[c.Field]
	if !ok {
		return false, fmt.Errorf("no field %s found in entity(%s)", c.Field, entity["id"])
	}

	switch c.Operator {
	case "=":
		return v == c.Value, nil
	}

	return false, fmt.Errorf("no operator %s", c.Operator)
}

func (c *Condition) ToString() string {
	return fmt.Sprintf("%s %s '%s'", c.Field, c.Operator, c.Value)
}

type QueryBuilder struct {
	query QueryNode
}

func NewQuery(query QueryNode) QueryBuilder {
	return QueryBuilder{query: query}
}

func (b QueryBuilder) And(q QueryNode) QueryNode {
	return &BinaryOp{Left: b.query, Operator: "AND", Right: q}
}

func (b QueryBuilder) Or(q QueryNode) QueryNode {
	return &BinaryOp{Left: b.query, Operator: "OR", Right: q}
}
