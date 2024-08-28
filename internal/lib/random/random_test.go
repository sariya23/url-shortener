package random

import (
	"regexp"
	"testing"
)

var testData = []struct {
	testCaseName string
	stringLen    int
}{
	{"Randow string with len 1", 1},
	{"Randow string with len 5", 5},
	{"Randow string with len 0", 0},
}

func TestGeneratorReturnDifferentStrings(t *testing.T) {
	for _, v := range testData {
		t.Run(v.testCaseName, func(*testing.T) {
			str1 := NewRandomString(v.stringLen)
			str2 := NewRandomString(v.stringLen)

			if str1 == str2 {
				t.Errorf("got identic strings from 2 generate (%s, %s)", str1, str2)
			}
		})
	}
}

func TestStringWithOnlySmallLetters(t *testing.T) {
	for _, v := range testData {
		t.Run(v.testCaseName, func(*testing.T) {
			s := NewRandomString(v.stringLen)
			matched, err := regexp.MatchString(`[a-z]+`, s)
			if err != nil {
				t.Errorf("something wrong (%v)", err)
			}
			if !matched {
				t.Errorf("There are characters other than small Latin letters in the string ")
			}
		})
	}
}
