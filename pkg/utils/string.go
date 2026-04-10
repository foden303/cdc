package utils

import (
	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

// DerefString returns the value of the string pointer or a default value if nil.
func DerefString(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// RandomString generates a random alphanumeric string of length n.
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
