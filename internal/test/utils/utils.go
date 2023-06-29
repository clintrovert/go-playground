package utils

import (
	"math/rand"

	"github.com/clintrovert/go-playground/api/model"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

func RandomPrefixedString(prefix string, n int) string {
	return prefix + "-" + RandomString(n)
}

func GenerateRandomUser() *model.User {
	return &model.User{
		Id:   int32(rand.Intn(1000)),
		Name: RandomPrefixedString("name", 10),
		Email: RandomPrefixedString("email", 5) +
			"@" + RandomPrefixedString("domain", 5) + ".com",
		Password: RandomPrefixedString("pwd", 10),
	}
}
