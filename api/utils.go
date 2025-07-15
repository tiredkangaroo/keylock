package api

func expand(i map[string]string) map[string][]string {
	v := make(map[string][]string)
	for key, val := range i {
		v[key] = []string{val}
	}
	return v
}
func shrink(i map[string][]string) map[string]string {
	m := make(map[string]string)
	for key, val := range i {
		m[key] = val[0]
	}
	return m
}
