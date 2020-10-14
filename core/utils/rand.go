package utils

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var MinMaxError = errors.New("Min cannot be greater than max.")
var Char_Buff = [26]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s",
	"t", "u", "v", "w", "x", "y", "z"}

func RandChoice(choices []interface{}) (interface{}, error) {
	var winner interface{}
	length := len(choices)
	i, err := IntRange(0, length)
	if err != nil {
		return nil, err
	}
	winner = choices[i]
	return winner, nil
}

func IntRange(min, max int) (int, error) {
	var result int
	switch {
	case min > max:
		return result, MinMaxError
	case min == max:
		result = max
	case min < max:
		rand.Seed(time.Now().UnixNano())
		result = min + rand.Intn(max-min)
	}
	return result, nil
}

func RandCode(codelen int) string {
	if codelen == 0 {
		return ""
	}
	numLen := rand.Intn(codelen)
	charLen := codelen - numLen
	var buff string
	for i := 0; i < numLen; i++ {
		buff = buff + strconv.Itoa(rand.Intn(10))
	}
	for i := 0; i < charLen; i++ {
		buff = buff + Char_Buff[rand.Intn(26)]
	}
	var code string
	arr := rand.Perm(codelen)
	for i := 0; i < 6; i++ {
		code = code + string(buff[arr[i]])
	}
	return strings.ToUpper(code)
}
func RandNumCode(codelen int) string {
	if codelen == 0 {
		return ""
	}
	var buff string
	for i := 0; i < codelen; i++ {
		buff = buff + strconv.Itoa(rand.Intn(10))
	}
	return strings.ToUpper(buff)
}
