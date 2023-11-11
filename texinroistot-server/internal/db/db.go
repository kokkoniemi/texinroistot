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

type VersionRepository interface {
	List() ([]*Version, error)
	Read(versionID int) (*Version, error)
	Create(version Version) (*Version, error)
	Remove(versionID int) error
	SetActive(versionID int) error
}

type AuthorRepository interface {
	List(version Version) ([]*Author, error)
	Read(authorID int) (*Author, error)
	BulkCreate(authors []*Author, version Version) ([]*Author, error)
}

type StoryRepository interface {
	List(version Version) ([]*Story, error)
	BulkCreate(stories []*Story, version Version) ([]*Story, error)
}
