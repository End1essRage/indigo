package storage

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type QueryNode interface {
	Evaluate(entity Entity) (bool, error)
	ToString() string
	Bson() bson.M
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

func (c *BinaryOp) Bson() bson.M {
	switch c.Operator {
	case "OR":
		return bson.M{"$or": bson.A{c.Left.Bson(), c.Right.Bson()}}
	case "AND":
		return bson.M{"$and": bson.A{c.Left.Bson(), c.Right.Bson()}}
	default:
		return bson.M{}
	}
}

type Condition struct {
	Field    string
	Operator string // = > < >= <= !=
	Value    interface{}
}

func (c *Condition) Evaluate(entity Entity) (bool, error) {
	v, ok := entity[c.Field]
	if !ok {
		return false, fmt.Errorf("no field %s found in entity", c.Field)
	}

	switch c.Operator {
	case "=", "==":
		return v == c.Value, nil
	case "!=":
		return v != c.Value, nil
	case ">":
		return compareValues(v, c.Value) > 0, nil
	case "<":
		return compareValues(v, c.Value) < 0, nil
	case ">=":
		return compareValues(v, c.Value) >= 0, nil
	case "<=":
		return compareValues(v, c.Value) <= 0, nil
	}

	return false, fmt.Errorf("unsupported operator %s", c.Operator)
}

// Вспомогательная функция для сравнения значений
func compareValues(a, b interface{}) int {
	switch aVal := a.(type) {
	case int:
		switch bVal := b.(type) {
		case int:
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		case float64:
			if float64(aVal) < bVal {
				return -1
			} else if float64(aVal) > bVal {
				return 1
			}
			return 0
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		case int:
			if aVal < float64(bVal) {
				return -1
			} else if aVal > float64(bVal) {
				return 1
			}
			return 0
		}
	case string:
		if bVal, ok := b.(string); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	}
	return 0
}

func (c *Condition) Bson() bson.M {
	switch c.Operator {
	case "=", "==":
		return bson.M{c.Field: bson.M{"$eq": c.Value}}
	case "!=":
		return bson.M{c.Field: bson.M{"$ne": c.Value}}
	case ">":
		return bson.M{c.Field: bson.M{"$gt": c.Value}}
	case "<":
		return bson.M{c.Field: bson.M{"$lt": c.Value}}
	case ">=":
		return bson.M{c.Field: bson.M{"$gte": c.Value}}
	case "<=":
		return bson.M{c.Field: bson.M{"$lte": c.Value}}
	default:
		return bson.M{}
	}
}

func (c *Condition) ToString() string {
	return fmt.Sprintf("%s %s %v", c.Field, c.Operator, c.Value)
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
