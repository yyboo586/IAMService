package interfaces

//go:generate mockgen -source=./logics.go -destination=mock/logics_mock.go -package=mock

type LogicsUser interface {
	// Create creates a new user
	Create(user *User) error
	// Login validate password and return user id, jwtToken
	Login(name, password string) (string, string, error)
}
