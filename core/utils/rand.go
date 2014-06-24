package utils

import (
	"errors"
	"math/rand"
	"time"
)

var MinMaxError = errors.New("Min cannot be greater than max.")

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
