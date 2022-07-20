package evaluation

import (
	"encoding/json"
	"github.com/spaolacci/murmur3"
)

func isEnabled(target Target, percentage int) bool {
	value, err := json.Marshal(target)
	if err != nil {
		return false
	}
	harsher := murmur3.New32()
	_, err = harsher.Write(value)
	if err != nil {
		log.Debugf("error %v", err)
	}
	hash := int(harsher.Sum32())
	result := (hash % oneHundred) + 1
	return percentage > 0 && result <= percentage
}

func evaluateItems(items []RolloutItem, target Target) interface{} {
	var value interface{}
	if len(items) == 0 {
		value = false
	}

	totalPercentage := 0
	for _, item := range items {
		value = item.Value
		totalPercentage += item.Weight
		if isEnabled(target, totalPercentage) {
			return value
		}
	}
	return value
}
