package strings_utility

func Map(arr []string, fn func(string) string) []string {
	ret := make([]string, len(arr))
	for idx, v := range arr {
		ret[idx] = fn(v)
	}
	return ret
}
