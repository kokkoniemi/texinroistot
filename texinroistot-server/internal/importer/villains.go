package importer

import "github.com/kokkoniemi/texinroistot/internal/db"

type importerVillain struct {
	ID   id
	item *db.Villain
}

type importerStoryVillain struct {
	ID      id
	story   id
	villain id
}

func (i *importer) importVillain(storyID id, r row) {}
