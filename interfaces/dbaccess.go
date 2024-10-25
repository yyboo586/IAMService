package interfaces

//go:generate mockgen -source=./dbaccess.go -destination=mock/dbaccess_mock.go -package=mock

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type DBUser interface {
	// Create creates a new user.
	Create(user *User) error
	// FetchByName fetches a user by name.
	FetchByName(name string) (*User, bool, error)
	// UpdateLoginTime updates the last login time of a user.
	UpdateLoginTime(id string) error
}
