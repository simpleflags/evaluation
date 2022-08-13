package evaluation

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
)

const (
	oneHundred = 100
)

var (
	// ErrQueryProviderMissing ...
	ErrQueryProviderMissing = errors.New("query field is missing in evaluator")
)

type visitor struct {
	variables []string
}

func (v *visitor) Enter(_ *ast.Node) {}
func (v *visitor) Exit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		if n.Value != "target" {
			v.variables = append(v.variables, n.Value)
		}
	}
}

// DataProvider provides methods for segment and flag retrieval
type DataProvider interface {
	GetVariable(key string) (Variable, error)
	GetConfiguration(key string) (Configuration, error)
}

// Evaluator engine evaluates flag from provided query
type Evaluator struct {
	provider DataProvider
}

// NewEvaluator constructs evaluator with query instance
func NewEvaluator(provider DataProvider) (*Evaluator, error) {
	if provider == nil {
		return &Evaluator{}, ErrQueryProviderMissing
	}
	return &Evaluator{
		provider: provider,
	}, nil
}

func (e *Evaluator) evaluateExpression(expression string, target Target) (bool, error) {
	if expression == "" { // return true if expression is empty it means to deliver to any
		return true, nil
	}

	values := make(map[string]interface{})

	targetMap := make(map[string]interface{})
	for key, value := range target {
		targetMap[key] = value
	}

	variables, err := Variables(expression)
	if err != nil {
		return false, err
	}

	for _, v := range variables {
		variable, err := e.provider.GetVariable(v)
		if err != nil {
			return false, err
		}
		values[v] = variable.Value
	}

	values["target"] = targetMap
	out, err := expr.Eval(expression, values)
	if err != nil {
		return false, err
	}
	result, ok := out.(bool)
	if !ok {
		return false, errors.New("type assertion error")
	}
	return result, nil
}

func (e *Evaluator) evaluateRules(rules []Rule, target Target) (interface{}, error) {
	if len(rules) == 0 {
		return nil, errors.New("no rules or target specified")
	}

	for _, rule := range rules {
		ok, err := e.evaluateExpression(rule.Expression, target)
		if err != nil {
			return nil, err
		}
		// rule ok?
		if ok {
			return rule.Value, nil
		}
	}
	return nil, nil
}

func (e *Evaluator) evaluateFlag(fc *Configuration, target Target) (interface{}, error) {
	if fc.On {
		v, err := e.evaluateRules(fc.Rules, target)
		if err != nil {
			return nil, err
		}
		// need to check if rule is satisfied
		if v != nil {
			return v, nil
		}
	}
	return fc.OffValue, nil
}

func (e *Evaluator) checkPreRequisite(parent *Configuration, target Target) error {
	if e.provider == nil {
		log.Errorf(ErrQueryProviderMissing.Error())
		return ErrQueryProviderMissing
	}
	prerequisites := parent.Prerequisites
	if prerequisites != nil {
		log.Infof(
			"Checking pre requisites %v of parent feature %v",
			prerequisites,
			parent.Identifier)
		for _, pre := range prerequisites {
			prereqFeature := pre.Identifier
			prereqFeatureConfig, err := e.provider.GetConfiguration(prereqFeature)
			if err != nil {
				log.Errorf(
					"Could not retrieve the pre requisite details of feature flag : %v", prereqFeature)
				return nil
			}

			prereqEvaluatedVariation, err := e.evaluateFlag(&prereqFeatureConfig, target)
			if err != nil {
				return err
			}

			// Compare if the pre requisite variation is a possible valid value of
			// the pre requisite FF
			log.Infof(
				"Pre requisite flag %v should have the value %v",
				prereqFeatureConfig.Identifier,
				pre.Value)
			if pre.Value != prereqEvaluatedVariation {
				return errors.New("prerequisites rule not satisfied")
			}
			// no need to check this error because it is recursion so err should never happen
			if err = e.checkPreRequisite(&prereqFeatureConfig, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Evaluator) Evaluate(key string, target Target) Evaluation {
	if e.provider == nil {
		log.Errorf(ErrQueryProviderMissing.Error())
		return Evaluation{
			err: ErrQueryProviderMissing,
		}
	}
	flag, err := e.provider.GetConfiguration(key)
	if err != nil {
		return Evaluation{
			err: err,
		}
	}

	evaluation := Evaluation{
		Project:     flag.Project,
		Environment: flag.Environment,
		Identifier:  flag.Identifier,
	}

	if flag.Prerequisites != nil {
		err = e.checkPreRequisite(&flag, target)
		if err != nil {
			return Evaluation{
				err: err,
			}
		}
	}

	val, err := e.evaluateFlag(&flag, target)
	if err != nil {
		evaluation.err = err
	} else {
		evaluation.Value = val
	}

	return evaluation
}

func Variables(expression string) ([]string, error) {
	tree, err := parser.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression, err: %w", err)
	}

	visitor := &visitor{}
	ast.Walk(&tree.Node, visitor)
	return visitor.variables, nil
}
