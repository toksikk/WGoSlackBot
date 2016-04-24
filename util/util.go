// vi:ts=4:sts=4:sw=4:noet:tw=72

package util

import (
	"bytes"
	"math/rand"
	"strconv"
	"time"
)

func ToInt(v interface{}) int {
	switch v.(type) {
	case int:
		return v.(int)
	case float64:
		return int(v.(float64))
	case string:
		ret, _ := strconv.Atoi(v.(string))
		return ret
	default:
		return 0
	}
}

func NumberToString(n int, sep rune) string {
	start := 0
	var buf bytes.Buffer

	s := strconv.Itoa(n)
	if n < 0 {
		start = 1
		buf.WriteByte('-')
	}
	l := len(s)
	ci := 3 - ((l - start) % 3)
	if ci == 3 {
		ci = 0
	}
	for i := start; i < l; i++ {
		if ci == 3 {
			buf.WriteRune(sep)
			ci = 0
		}
		ci++
		buf.WriteByte(s[i])
	}
	return buf.String()
}

func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
