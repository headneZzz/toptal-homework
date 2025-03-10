package domain

type User struct {
	Id           int
	Username     string
	PasswordHash string
	Admin        bool
}
