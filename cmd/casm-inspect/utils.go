package main

// lookupKeys performs a multi-level map search given a list of keys to query.
// Given a map like {"a": {"b": {"c": 10}}} and keys ["a", "b", "c"] this
// function will return 10 (a value of the deepest lookup).
func lookupKeys(m map[string]any, keys ...string) any {
	var current any = m
	for _, k := range keys {
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = asMap[k]
	}
	return current
}
