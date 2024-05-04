package importer

import (
	"strings"

	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type importerVillain struct {
	ID                id
	externalVillainID int64 // not saved to db
	item              *db.Villain
}

type importerStoryVillain struct {
	ID      id
	story   id
	villain id
	item    *db.StoryVillain
}

func (i *importer) importVillain(storyID id, r row) {
	villainID := strings.TrimSpace(r.getValue("villain_id"))
	villainRanks := i.TrimmedSplit(r.getValue("ranks"), ";")
	firstNames := i.TrimmedSplit(r.getValue("first_names"), ";")
	lastName := r.getValue("last_name")
	nicknames := i.TrimmedSplit(r.getValue("nicknames"), ";")
	aliases := i.TrimmedSplit(r.getValue("aliases"), ";")
	roles := i.TrimmedSplit(r.getValue("roles"), ";")
	destiny := i.TrimmedSplit(r.getValue("destiny"), ";")

	var villain *importerVillain

	createHash := func() string {
		if len(villainID) > 0 {
			return crypt.Hash(villainID)
		}
		return crypt.Hash(strings.Join(firstNames, "") +
			lastName +
			strings.Join(nicknames, "") +
			strings.Join(aliases, "") +
			strings.Join(roles, "") +
			strings.Join(destiny, ""))
	}

	if !i.hasVillain(villainID) {
		villain = i.addVillain(&db.Villain{
			Hash:       createHash(),
			Ranks:      ranks,
			FirstNames: firstNames,
			LastName:   lastName})
	} else {
		villain = i.getVillain(villainID)
		// check if villain has rank, fistNames and lastNames
		// and add those if they're missing

	}

	// check if importerStoryVillain exists and load
	// 	else create empty storyVillain

	// check if storyVillain has nicknames, aliases, roles
	// and desiny and add those if they're missing
}
