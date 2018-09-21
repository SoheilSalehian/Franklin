package main

import "math/rand"

// FIXME: need more robust validations (github.com/asaskevich/govalidator)
func userValidations(username string, password string) bool {
	if len(username) >= 255 {
		return false
	}
	return true

}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func compare(X, Y []int) []int {
	m := make(map[int]int)

	for _, y := range Y {
		m[y]++
	}

	var ret []int
	for _, x := range X {
		if m[x] > 0 {
			m[x]--
			continue
		}
		ret = append(ret, x)
	}

	return ret
}
