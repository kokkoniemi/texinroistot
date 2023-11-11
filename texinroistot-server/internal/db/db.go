package db

const (
	DefaultPageSize = 25
	StartPage       = 0
)

type ListMeta struct {
	Total     int
	PageIndex int
	PageSize  int
}

type UserRepository interface {
	List(pageIndex int) ([]*User, *ListMeta, error)
	Create(user User) (*User, error)
	Remove(userHash string) error
	SetAdmin(userHash string) (*User, error)
}
