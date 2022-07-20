package evaluation

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

type Variable struct {
	Identifier string      `json:"identifier"`
	Value      interface{} `json:"value"`
} // @name Var

type Rule struct {
	Expression string      `json:"expression"`
	Value      interface{} `json:"value"`
} // @name Rule

type RolloutItem struct {
	Value  interface{} `json:"__value__"`
	Weight int         `json:"__weight__"`
} // @name RolloutItem

type Prerequisite struct {
	Identifier string      `json:"identifier" validate:"required"`
	Value      interface{} `json:"value" validate:"required"`
} // @name Prerequisite

type Configuration struct {
	Project       string         `json:"project"`
	Environment   string         `json:"environment"`
	Identifier    string         `json:"identifier"`
	Deprecated    bool           `json:"deprecated"`
	On            bool           `json:"on"`
	OnValue       interface{}    `json:"on_value"`
	OffValue      interface{}    `json:"off_value"`
	Rules         []Rule         `json:"rules"`
	Prerequisites []Prerequisite `json:"prerequisites"`
	Version       uint           `json:"version"`
} // @name Configuration

type Configurations []Configuration // @name Configurations

type Target map[string]interface{} // @name Target

// GetAttrValue returns value from target with specified attribute
func (t Target) GetAttrValue(attr string) reflect.Value {
	var value reflect.Value

	attrVal, ok := t[attr] // first check custom attributes
	if ok {
		value = reflect.ValueOf(attrVal)
	}
	return value
}

type Evaluation struct {
	Project     string      `json:"project"`
	Environment string      `json:"environment"`
	Identifier  string      `json:"identifier"`
	Value       interface{} `json:"value"`
	err         error
} // @name Evaluation

type Evaluations []Evaluation // @name Evaluations

func (v Evaluation) IsNone() bool {
	return v.Value == nil
}

func (v Evaluation) Bool(defaultValue bool) bool {
	if v.Value == nil || v.err != nil {
		return defaultValue
	}
	switch v.Value.(type) {
	case bool:
		return v.Value.(bool)
	case string:
		value := v.Value.(string)
		return strings.ToLower(value) == "true" || value != "0"
	case int:
		value := v.Value.(int)
		return value > 0
	case float32:
		value := v.Value.(float32)
		return value > 0
	case float64:
		value := v.Value.(float64)
		return value > 0
	default:
		return defaultValue
	}
}

func (v Evaluation) String(defaultValue string) string {
	if v.Value == nil || v.err != nil {
		return defaultValue
	}
	switch v.Value.(type) {
	case bool:
		value := v.Value.(bool)
		if value {
			return "true"
		}
		return "false"
	case string:
		return v.Value.(string)
	case int:
		value := v.Value.(int)
		return strconv.Itoa(value)
	case float32:
		value := v.Value.(float32)
		return strconv.FormatFloat(float64(value), 'g', 15, 32)
	case float64:
		value := v.Value.(float64)
		return strconv.FormatFloat(value, 'g', 15, 32)
	default:
		return defaultValue
	}
}

func (v Evaluation) Int(defaultValue int) int {
	if v.Value == nil || v.err != nil {
		return defaultValue
	}
	switch v.Value.(type) {
	case bool:
		value := v.Value.(bool)
		if value {
			return 1
		}
		return 0
	case string:
		value := v.Value.(string)
		i, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return i
	case int:
		value := v.Value.(int)
		return value
	case float32:
		value := v.Value.(float32)
		return int(value)
	case float64:
		value := v.Value.(float64)
		return int(value)
	default:
		return defaultValue
	}
}

func (v Evaluation) Number(defaultValue float64) float64 {
	if v.Value == nil || v.err != nil {
		return defaultValue
	}
	switch v.Value.(type) {
	case bool:
		value := v.Value.(bool)
		if value {
			return 1
		}
		return 0
	case string:
		value := v.Value.(string)
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return defaultValue
		}
		return f
	case int:
		value := v.Value.(int)
		return float64(value)
	case float32:
		value := v.Value.(float32)
		return float64(value)
	case float64:
		value := v.Value.(float64)
		return value
	default:
		return defaultValue
	}
}

func (v Evaluation) Map(defaultValue map[string]interface{}) map[string]interface{} {
	if v.Value == nil || v.err != nil {
		return defaultValue
	}
	switch v.Value.(type) {
	case bool:
		value := v.Value.(bool)
		return map[string]interface{}{
			"value": value,
		}
	case string:
		value := v.Value.(string)
		var smap map[string]interface{}
		err := json.NewDecoder(strings.NewReader(value)).Decode(&smap)
		if err != nil {
			return defaultValue
		}
		return smap
	case int:
		value := v.Value.(int)
		return map[string]interface{}{
			"value": value,
		}
	case float32:
		value := v.Value.(float32)
		return map[string]interface{}{
			"value": strconv.FormatFloat(float64(value), 'f', -1, 32),
		}
	case float64:
		value := v.Value.(float64)
		return map[string]interface{}{
			"value": strconv.FormatFloat(value, 'f', -1, 32),
		}
	default:
		return defaultValue
	}
}
