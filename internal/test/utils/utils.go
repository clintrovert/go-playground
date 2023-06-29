package utils

import (
	"database/sql"
	"math/rand"

	"github.com/clintrovert/go-playground/pkg/postgres/database"
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

func GenerateRandomUser() database.User {
	return database.User{
		UserID: int32(rand.Intn(1000)),
		Name: sql.NullString{
			String: RandomPrefixedString("name", 10),
			Valid:  true,
		},
		Email: sql.NullString{
			String: RandomPrefixedString("email", 5) +
				"@" + RandomPrefixedString("domain", 5) + ".com",
			Valid: true,
		},
		Password: sql.NullString{
			String: RandomPrefixedString("pwd", 10),
			Valid:  true,
		},
	}
}
