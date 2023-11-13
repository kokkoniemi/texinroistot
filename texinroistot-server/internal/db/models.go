package db

import (
	"reflect"
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

type AuthorBlueprint struct {
	ID         interface{}
	FirstName  interface{}
	LastName   interface{}
	IsWriter   interface{}
	IsDrawer   interface{}
	IsInventor interface{}
}

func (a *AuthorBlueprint) AuthorExists() bool {
	return a.ID != nil
}

func (a *AuthorBlueprint) ToAuthor() *Author {
	from := reflect.ValueOf(a).Elem()
	to := &Author{}
	t := reflect.TypeOf(to).Elem()
	v := reflect.ValueOf(to).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		fieldName := t.Field(i).Name

		fromf := from.FieldByName(fieldName)
		if fromf.IsNil() || !f.IsValid() || !f.CanSet() {
			continue
		}
		switch f.Kind() {
		case reflect.Int, reflect.Int64:
			f.SetInt(fromf.Elem().Int())
		case reflect.Bool:
			val := fromf.Elem().Bool()
			f.SetBool(val)
		case reflect.String:
			f.SetString(fromf.Elem().String())
		}
	}

	return to
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

func (s *Story) GetOriginalTitleForDB() interface{} {
	if s.OriginalTitle != "" {
		return s.OriginalTitle
	}
	return nil
}

func (s *Story) GetOrderNumberForDB() interface{} {
	if s.OrderNumber != 0 {
		return s.OrderNumber
	}
	return nil
}

func (s *Story) GetWriterIDForDB() interface{} {
	if s.WrittenBy != nil {
		return s.WrittenBy.ID
	}
	return nil
}

func (s *Story) GetDrawerIDForDB() interface{} {
	if s.DrawnBy != nil {
		return s.DrawnBy.ID
	}
	return nil
}

func (s *Story) GetInventorIDForDB() interface{} {
	if s.InventedBy != nil {
		return s.InventedBy.ID
	}
	return nil
}
