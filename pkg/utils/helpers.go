package utils

import (
	"math/rand"
	"time"
)

const uidBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const lettersBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func GenerateUID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = uidBytes[rand.Intn(len(uidBytes))]
	}
	return string(b)
}

func GenerateName(n int) string {
	b := make([]byte, n)
	b[0] = lettersBytes[rand.Intn(len(lettersBytes))]
	for i := range b[1:] {
		b[i+1] = uidBytes[rand.Intn(len(uidBytes))]
	}
	return string(b)
}

// Must panic if err != nil
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Min is like math.Min for int
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// MinDuration is like math.Min for time.Duration
func MinDuration(x, y time.Duration) time.Duration {
	if x < y {
		return x
	}
	return y
}

// MaxDuration is like math.Max for time.Duration
func MaxDuration(x, y time.Duration) time.Duration {
	if x >= y {
		return x
	}
	return y
}

func ConvertArgsInt64(args []int64) []interface{} {
	c := make([]interface{}, len(args))
	for i := range args {
		c[i] = args[i]
	}
	return c
}

func ConvertArgsString(args []string) []interface{} {
	c := make([]interface{}, len(args))
	for i := range args {
		c[i] = args[i]
	}
	return c
}
