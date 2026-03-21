package db

const (
	DefaultPageSize   = 25
	StartPage         = 0
	MaxBulkCreateSize = 100
)

type ListMeta struct {
	Total     int
	PageIndex int
	PageSize  int
}

type StoryListParams struct {
	Publication string
	Sort        string
	Search      string
	Year        int
	Page        int
	PageSize    int
}

type VillainListParams struct {
	Publication string
	Sort        string
	Search      string
	Page        int
	PageSize    int
}

type UserRepository interface {
	List(pageIndex int) ([]*User, *ListMeta, error)
	ReadByHash(userHash string) (*User, error)
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
	GetActive() (*Version, error)
	GetStats(versionID int) (*VersionStats, error)
}

type AuthorRepository interface {
	List(version *Version) ([]*Author, error)
	Read(authorID int) (*Author, error)
	BulkCreate(authors []*Author, version *Version) ([]*Author, error)
}

type StoryRepository interface {
	List(version *Version, limit int, offset int) ([]*Story, error)
	ListFiltered(version *Version, params StoryListParams) ([]*Story, int, error)
	ListByAuthorHash(version *Version, authorHash string, authorType string) ([]*Story, bool, error)
	BulkCreate(stories []*Story, version *Version) ([]*Story, error)
	BulkCreatePublications(publications []*Publication, version *Version) ([]*Publication, error)
}

type VillainRepository interface {
	BulkCreate(villains []*Villain, version *Version) ([]*Villain, error)
	ListFiltered(version *Version, params VillainListParams) ([]*Villain, int, error)
	ListByStoryHash(version *Version, storyHash string) ([]*Villain, bool, error)
	//BulkCreateStoryVillain(storyVillains []*StoryVillain) ([]*StoryVillain, error)
}
