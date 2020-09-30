package pager

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type (
	// RangeKey Data field
	RangeKey string
	// RangeType Types
	//  Gte: Greater than or equal to
	//  Lte: Less than or equal to
	RangeType int
	// Range Built range query parameters
	Range map[RangeKey]map[RangeType]int64
)

// Int64Slice int64 sort
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

const (
	// Gte Greater than or equal to
	Gte RangeType = iota
	// Lte Less than or equal to
	Lte
)

func (s RangeType) String() string {
	switch s {
	case Gte:
		return "gte"
	case Lte:
		return "lte"
	}
	return ""
}

// Add add
func (r Range) Add(key RangeKey, typ RangeType, val int64) {
	if r[key] == nil {
		r[key] = make(map[RangeType]int64)
	}
	r[key][typ] = val
}

// Parse parse
func (r Range) Parse(request *http.Request) {
	query := request.URL.Query()["range"]
	for _, v := range query {
		key, val := r.parseVal(v)
		switch true {
		case r.isGte(key, val):
			r.Add(r.getKey(key), Gte, r.getVal(val)[0])
		case r.isLte(key, val):
			r.Add(r.getKey(key), Lte, r.getVal(val)[0])
		case r.isGteLte(val):
			rs := r.getVal(val)
			r.Add(r.getKey(key), Gte, rs[0])
			r.Add(r.getKey(key), Lte, rs[1])
		}
	}
}

func (r Range) isGte(key, val string) bool {
	return !strings.HasPrefix(key, "-") && len(strings.Split(val, ",")) == 1
}

func (r Range) isLte(key, val string) bool {
	return strings.HasPrefix(key, "-") && len(strings.Split(val, ",")) == 1
}

func (r Range) isGteLte(val string) bool {
	return len(strings.Split(val, ",")) == 2
}

func (r Range) getVal(val string) []int64 {
	vals := strings.Split(val, ",")
	if len(vals) == 1 {
		v, err := strconv.Atoi(vals[0])
		if err != nil {
			return []int64{0}
		}
		return []int64{int64(v)}
	}

	var rs []int64
	for i := 0; i < 2; i++ {
		v, err := strconv.Atoi(vals[i])
		if err != nil {
			rs = append(rs, 0)
		} else {
			rs = append(rs, int64(v))
		}
	}

	sort.Sort(Int64Slice(rs))
	return rs

}

func (r Range) getKey(key string) RangeKey {
	return RangeKey(strings.TrimPrefix(strings.TrimPrefix(key, "-"), "+"))
}

func (r Range) parseVal(v string) (key, val string) {
	index := strings.Index(v, ":")
	runeVal := []rune(v)
	return string(runeVal[0:index]), string(runeVal[index+1:])
}
