package util

import (
	"math/rand"
	"strings"
)

const (
	LETTERS = "ABCDEFGHJKLMNPQRSTUVWXYZ"
)

func GenerateRoomCode() string {
	var letters = []rune(LETTERS)
	b := make([]rune, 4)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func IsValidRoomCode(code string) bool {
	if len(code) != 4 {
		return false
	}
	for _, c := range code {
		if !strings.ContainsRune(LETTERS, c) {
			return false
		}
	}
	return true
}
