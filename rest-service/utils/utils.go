package utils

import (
	"math/rand"
	"strings"
)

const alphameric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphameric)

	for i := 0; i < n; i++ {
		c := alphameric[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}
