package random

import "math/rand/v2"

const DefaultStringLen = 6

func NewRandomString(strLen int) string {
	min := 97
	max := 123
	var res string

	if strLen == 0 {
		strLen = DefaultStringLen
	}

	for i := 0; i < strLen; i++ {
		res += string(rune(rand.IntN(max-min) + min))
	}

	return res
}
