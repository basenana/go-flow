package utils

type StringSet map[string]struct{}

func (s StringSet) Insert(t string) {
	s[t] = struct{}{}
}
func (s StringSet) Has(task string) bool {
	_, ok := s[task]
	return ok
}

func (s StringSet) Del(t string) {
	if _, ok := s[t]; ok {
		delete(s, t)
	}
}

func (s StringSet) List() (result []string) {
	for k := range s {
		result = append(result, k)
	}
	return
}

func (s StringSet) Len() int {
	return len(s)
}

func NewStringSet() StringSet {
	return map[string]struct{}{}
}
