package datastructs

import "container/list"

func MapHasKey(m map[string]struct{}, k string) bool {
	if m == nil {
		return false
	}
	_, exists := m[k]
	return exists
}

func SliceToSet(sl []string) map[string]struct{} {
	res := map[string]struct{}{}
	for _, e := range sl {
		res[e] = struct{}{}
	}
	return res
}

func ListToSlice(l *list.List) []string {
	res := make([]string, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		res = append(res, e.Value.(string))
	}
	return res
}
