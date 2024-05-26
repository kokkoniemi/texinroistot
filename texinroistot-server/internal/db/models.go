package db

import (
	"reflect"
	"time"
)

type User struct {
	ID        int       `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	Hash      string    `json:"hash"`
	IsAdmin   bool      `json:"isAdmin"`
}

type Author struct {
	ID         int    `json:"-"`
	Hash       string `json:"hash"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	IsWriter   bool   `json:"isWriter"`
	IsDrawer   bool   `json:"isDrawer"`
	IsInventor bool   `json:"isInventor"`
}

type AuthorBlueprint struct {
	ID         interface{}
	Hash       interface{}
	FirstName  interface{}
	LastName   interface{}
	IsWriter   interface{}
	IsDrawer   interface{}
	IsInventor interface{}
}

func (a *AuthorBlueprint) AuthorExists() bool {
	return a.ID != nil
}

// ToAuthor converts AuthorBlueprint to corresponding Author struct
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

// Publication may have multiple stories and a story can be published in multiple publications
type Publication struct {
	ID    int    `json:"-"`
	Hash  string `json:"hash"`
	Type  string `json:"type"`
	Year  int    `json:"year"`
	Issue string `json:"issue"`
}

type StoryPublication struct {
	ID    int          `json:"-"`
	Title string       `json:"title"`
	In    *Publication `json:"in"`
}

type Villain struct {
	ID         int             `json:"-"`
	Hash       string          `json:"hash"`
	Ranks      []string        `json:"ranks"`
	FirstNames []string        `json:"firstNames"`
	LastName   string          `json:"lastName"`
	As         []*StoryVillain `json:"as"`
}

type StoryVillain struct {
	ID        int      `json:"-"`
	Hash      string   `json:"hash"`
	Nicknames []string `json:"nicknames"`
	Aliases   []string `json:"aliases"`
	Roles     []string `json:"roles"`
	Destiny   []string `json:"destiny"`
	Story     *Story   `json:"story"`
}

type Story struct {
	ID           int                 `json:"-"`
	Hash         string              `json:"hash"`
	OrderNumber  int                 `json:"orderNumber"`
	WrittenBy    []*Author           `json:"writtenBy"`
	DrawnBy      []*Author           `json:"drawnBy"`
	InventedBy   []*Author           `json:"inventedBy"`
	Publications []*StoryPublication `json:"publications"`
}

func (s *Story) GetOrderNumberForDB() interface{} {
	if s.OrderNumber != 0 {
		return s.OrderNumber
	}
	return nil
}

// func (s *Story) GetWriterIDForDB() interface{} {
// 	if s.WrittenBy != nil {
// 		return s.WrittenBy.ID
// 	}
// 	return nil
// }

// func (s *Story) GetDrawerIDForDB() interface{} {
// 	if s.DrawnBy != nil {
// 		return s.DrawnBy.ID
// 	}
// 	return nil
// }

// func (s *Story) GetInventorIDForDB() interface{} {
// 	if s.InventedBy != nil {
// 		return s.InventedBy.ID
// 	}
// 	return nil
// }
