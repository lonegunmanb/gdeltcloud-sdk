package gdeltcloud

import (
	"net/url"
	"strconv"
	"strings"
)

// setStr adds a string parameter if non-empty.
func setStr(v url.Values, key, val string) {
	if val != "" {
		v.Set(key, val)
	}
}

// setCSV adds a comma-separated list parameter if the slice is non-empty.
func setCSV(v url.Values, key string, vals []string) {
	if len(vals) > 0 {
		v.Set(key, strings.Join(vals, ","))
	}
}

// setInt adds an integer parameter if it is greater than zero.
func setInt(v url.Values, key string, val int) {
	if val > 0 {
		v.Set(key, strconv.Itoa(val))
	}
}

// setBool adds a boolean parameter as "true"/"false" when set is true.
func setBool(v url.Values, key string, val, set bool) {
	if set {
		v.Set(key, strconv.FormatBool(val))
	}
}
