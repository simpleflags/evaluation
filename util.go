package evaluation

import (
	"strings"

	"github.com/spaolacci/murmur3"
)

func getNormalizedNumber(identifier, bucketBy string) int {
	value := []byte(strings.Join([]string{bucketBy, identifier}, ":"))
	hasher := murmur3.New32()
	_, err := hasher.Write(value)
	if err != nil {
		log.Debugf("error %v", err)
	}
	hash := int(hasher.Sum32())
	return (hash % oneHundred) + 1
}

func isEnabled(target Target, bucketBy string, percentage int) bool {
	value := target.GetAttrValue(bucketBy)
	identifier := value.String()
	if identifier == "" {
		return false
	}

	bucketID := getNormalizedNumber(identifier, bucketBy)
	return percentage > 0 && bucketID <= percentage
}

func evaluateDistribution(distribution *Distribution, target Target) string {
	variation := ""
	if distribution == nil {
		return variation
	}

	totalPercentage := 0
	for _, wv := range distribution.Variations {
		variation = wv.Variation
		totalPercentage += wv.Weight
		if isEnabled(target, distribution.BucketBy, totalPercentage) {
			return wv.Variation
		}
	}
	return variation
}
