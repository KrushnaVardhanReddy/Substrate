package gates

// EvaluateGate returns true if the quality gate checks pass.
func EvaluateGate(gate string, breakingCount, warningCount int, crossRepoSafe bool) bool {
	switch gate {
	case "strict":
		return breakingCount == 0 && warningCount == 0 && crossRepoSafe
	case "permissive":
		return crossRepoSafe
	case "standard":
		fallthrough
	default:
		return breakingCount == 0 && crossRepoSafe
	}
}
