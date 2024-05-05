package importer

import (
	"fmt"
	"slices"
	"strconv"
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

func (i *importer) getVillainIndex(villainID int64) int {
	return slices.IndexFunc(i.villains, func(v *importerVillain) bool {
		return v.externalVillainID == villainID
	})
}

func (i *importer) hasVillain(villainID int64) bool {
	return i.getVillainIndex(villainID) != -1
}

func (i *importer) getVillain(villainID int64) *importerVillain {
	idx := i.getVillainIndex(villainID)
	if idx != -1 {
		return i.villains[idx]
	}
	return nil
}

func (i *importer) addVillain(villain *db.Villain, villainID int64) *importerVillain {
	i.totalEntities++

	importerVillain := &importerVillain{
		ID:                id(i.totalEntities),
		externalVillainID: villainID,
		item:              villain,
	}
	i.villains = append(i.villains, importerVillain)

	return importerVillain
}

func (i *importer) getStoryVillainIndex(villainID id, storyID id) int {
	return slices.IndexFunc(i.storyVillains, func(sv *importerStoryVillain) bool {
		return sv.villain == villainID && sv.story == storyID
	})
}

func (i *importer) hasStoryVillain(villainID id, storyID id) bool {
	return i.getStoryVillainIndex(villainID, storyID) != -1
}

func (i *importer) getStoryVillain(villainID id, storyID id) *importerStoryVillain {
	idx := i.getStoryVillainIndex(villainID, storyID)
	if idx != -1 {
		return i.storyVillains[idx]
	}
	return nil
}

func (i *importer) addStoryVillain(storyVillain *db.StoryVillain, villainID id, storyID id) *importerStoryVillain {
	i.totalEntities++

	importerStoryVillain := &importerStoryVillain{
		ID:      id(i.totalEntities),
		story:   storyID,
		villain: villainID,
		item:    storyVillain,
	}
	i.storyVillains = append(i.storyVillains, importerStoryVillain)

	return importerStoryVillain
}

func (i *importer) importVillain(storyID id, r row) {
	villainID, err := strconv.ParseInt(strings.TrimSpace(r.getValue("villain_id")), 10, 64)
	if err != nil {
		villainID = -1
	}
	villainRanks := i.TrimmedSplit(r.getValue("ranks"), ";")
	firstNames := i.TrimmedSplit(r.getValue("first_names"), ";")
	lastName := r.getValue("last_name")
	nicknames := i.TrimmedSplit(r.getValue("nicknames"), ";")
	aliases := i.TrimmedSplit(r.getValue("aliases"), ";")
	roles := i.TrimmedSplit(r.getValue("roles"), ";")
	destiny := i.TrimmedSplit(r.getValue("destiny"), ";")

	var villain *importerVillain

	createHash := func() string {
		if villainID != -1 {
			return crypt.Hash(strconv.FormatInt(villainID, 10))
		}
		return crypt.Hash(strings.Join(firstNames, "") +
			lastName +
			strings.Join(nicknames, "") +
			strings.Join(aliases, "") +
			strings.Join(roles, "") +
			strings.Join(destiny, ""))
	}

	// create villain if not exist
	if villainID == -1 || !i.hasVillain(villainID) {
		villain = i.addVillain(&db.Villain{
			Hash:       createHash(),
			Ranks:      villainRanks,
			FirstNames: firstNames,
			LastName:   lastName,
		}, villainID)
	} else {
		// update existing villain with row info
		villain = i.getVillain(villainID)

		for _, rank := range villainRanks {
			if !slices.Contains(villain.item.Ranks, rank) {
				villain.item.Ranks = append(villain.item.Ranks, rank)
			}
		}

		for _, name := range firstNames {
			if !slices.Contains(villain.item.FirstNames, name) {
				villain.item.FirstNames = append(villain.item.FirstNames, name)
			}
		}

		if len(villain.item.LastName) == 0 && len(lastName) > 0 {
			villain.item.LastName = lastName
		}

	}

	var storyVillain *importerStoryVillain

	// create storyVillain if not exist
	if !i.hasStoryVillain(villain.ID, storyID) {
		storyVillain = i.addStoryVillain(&db.StoryVillain{
			Hash:      crypt.Hash(fmt.Sprintf("%d%d", villain.ID, storyID)),
			Nicknames: nicknames,
			Aliases:   aliases,
			Roles:     roles,
			Destiny:   destiny,
		}, villain.ID, storyID)
	} else {
		// update existing storyVillain with row info
		storyVillain = i.getStoryVillain(villain.ID, storyID)

		for _, nick := range nicknames {
			if !slices.Contains(storyVillain.item.Nicknames, nick) {
				storyVillain.item.Nicknames = append(storyVillain.item.Nicknames, nick)
			}
		}

		for _, alias := range aliases {
			if !slices.Contains(storyVillain.item.Aliases, alias) {
				storyVillain.item.Aliases = append(storyVillain.item.Aliases, alias)
			}
		}

		for _, role := range roles {
			if !slices.Contains(storyVillain.item.Roles, role) {
				storyVillain.item.Roles = append(storyVillain.item.Roles, role)
			}
		}
	}
}
