package finops

// EstimateResponseByteSize recursively calculates a heuristic byte size
// for an OpenAPI-like JSON schema.
func EstimateResponseByteSize(schema map[string]interface{}) int64 {
	if schema == nil {
		return 0
	}

	// If type is defined
	if t, ok := schema["type"].(string); ok {
		switch t {
		case "string":
			return 50
		case "integer", "number":
			return 8
		case "boolean":
			return 1
		case "array":
			// Estimate array multiplier as 10 elements
			if items, ok := schema["items"].(map[string]interface{}); ok {
				return 10 * EstimateResponseByteSize(items)
			}
			return 0
		case "object":
			var size int64 = 0
			if props, ok := schema["properties"].(map[string]interface{}); ok {
				for _, prop := range props {
					if propMap, ok := prop.(map[string]interface{}); ok {
						size += EstimateResponseByteSize(propMap)
					}
				}
			}
			return size
		}
	}

	// Handle case where type is not explicitly defined but properties are present
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		var size int64 = 0
		for _, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				size += EstimateResponseByteSize(propMap)
			}
		}
		return size
	}

	return 0
}
