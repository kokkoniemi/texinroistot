package db

import (
	"fmt"
	"time"
)

type User struct {
	ID        int
	CreatedAt time.Time `json:"createdAt"`
	Hash      string    `json:"hash"`
	IsAdmin   bool      `json:"isAdmin"`
}

type Author struct {
	ID         int
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	IsWriter   bool   `json:"isWriter"`
	IsDrawer   bool   `json:"isDrawer"`
	IsInventor bool   `json:"isInventor"`
}

func (a *Author) Exists() bool {
	return a.ID != 0 && len(a.FirstName) > 0
}

type Version struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	IsActive  bool      `json:"isActive"`
}

type Publication struct {
	ID    int
	Type  string `json:"type"`
	Year  int    `json:"year"`
	Issue string `json:"issue"`
}

type Villain struct {
	ID         int
	Ranks      []string `json:"ranks"`
	FirstNames []string `json:"firstNames"`
	LastName   string   `json:"lastName"`
	Nicnames   []string `json:"nicknames"`
	Aliases    []string `json:"aliases"`
	Role       string   `json:"role"`
	Destiny    string   `json:"destiny"`
}

type Story struct {
	ID            int
	Title         string `json:"title"`
	OriginalTitle string `json:"originalTitle"`
	OrderNumber   int    `json:"orderNumber"`
	WrittenBy     *Author
	DrawnBy       *Author
	InventedBy    *Author
}

func (s *Story) GetOriginalTitleForDB() string {
	if s.OriginalTitle != "" {
		return fmt.Sprintf("%v", s.OriginalTitle)
	}
	return "null"
}

func (s *Story) GetOrderNumberForDB() string {
	if s.OrderNumber != 0 {
		return fmt.Sprintf("%v", s.OrderNumber)
	}
	return "null"
}

func (s *Story) GetWriterIDForDB() string {
	if s.WrittenBy != nil {
		return fmt.Sprintf("%v", s.WrittenBy.ID)
	}
	return "null"
}

func (s *Story) GetDrawerIDForDB() string {
	if s.DrawnBy != nil {
		return fmt.Sprintf("%v", s.DrawnBy.ID)
	}
	return "null"
}

func (s *Story) GetInventorIDForDB() string {
	if s.InventedBy != nil {
		return fmt.Sprintf("%v", s.InventedBy.ID)
	}
	return "null"
}
