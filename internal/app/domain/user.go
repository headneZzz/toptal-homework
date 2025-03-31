package domain

import "fmt"

type User struct {
	id           int
	username     string
	passwordHash string
	admin        bool
}

func NewUser(id int, username string, passwordHash string, admin bool) (User, error) {
	user := User{}
	if err := user.SetId(id); err != nil {
		return user, err
	}
	if err := user.SetUsername(username); err != nil {
		return user, err
	}
	if err := user.SetPasswordHash(passwordHash); err != nil {
		return user, err
	}
	if err := user.SetAdmin(admin); err != nil {
		return user, err
	}
	return user, nil
}

func NewUserWithDefaultId(username string, passwordHash string) (User, error) {
	user := User{}
	if err := user.SetUsername(username); err != nil {
		return user, err
	}
	if err := user.SetPasswordHash(passwordHash); err != nil {
		return user, err
	}
	return user, nil
}

// Getter methods

func (u *User) Id() int {
	return u.id
}

func (u *User) Username() string {
	return u.username
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) Admin() bool {
	return u.admin
}

// Setter methods

func (u *User) SetId(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id: %d", id)
	}
	u.id = id
	return nil
}

func (u *User) SetUsername(username string) error {
	if username == "" {
		return fmt.Errorf("invalid user username: %s", username)
	}
	u.username = username
	return nil
}

func (u *User) SetPasswordHash(passwordHash string) error {
	if passwordHash == "" {
		return fmt.Errorf("invalid user password hash: %s", passwordHash)
	}
	u.passwordHash = passwordHash
	return nil
}

func (u *User) SetAdmin(admin bool) error {
	u.admin = admin
	return nil
}
