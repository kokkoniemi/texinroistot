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

func (i *importer) loadVillain(storyID id, r row) error {
	villainID, err := strconv.ParseInt(strings.TrimSpace(r.getValue("villain_id")), 10, 64)
	if err != nil {
		villainID = -1
	}
	villainRanks := i.TrimmedSplit(r.getValue("ranks"), ";")
	firstNames := i.TrimmedSplit(r.getValue("first_names"), ";")
	lastName := r.getValue("last_name")
	nicknames := i.TrimmedSplit(r.getValue("nicknames"), ";")
	otherNames := i.TrimmedSplit(r.getValue("other_names"), ";")
	codeNames := i.TrimmedSplit(r.getValue("code_names"), ";")
	roles := i.TrimmedSplit(r.getValue("roles"), ";")
	destiny := i.TrimmedSplit(r.getValue("destiny"), ";")

	var villain *importerVillain

	createHash := func() string {
		if villainID != -1 {
			return crypt.Hash(strconv.FormatInt(villainID, 10))
		}
		return crypt.Hash(fmt.Sprintf("%d", r.index) +
			strings.Join(firstNames, "") +
			lastName +
			strings.Join(nicknames, "") +
			strings.Join(otherNames, "") +
			strings.Join(codeNames, "") +
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

	// same villain + same story must not appear twice in the source data
	if i.hasStoryVillain(villain.ID, storyID) {
		if villainID != -1 {
			return fmt.Errorf(
				"duplicate villain_id %d for story %d at row %d",
				villainID,
				storyID,
				r.index+2,
			)
		}
		return fmt.Errorf("duplicate villain in story %d at row %d", storyID, r.index+2)
	}

	i.addStoryVillain(&db.StoryVillain{
		Hash:       crypt.Hash(fmt.Sprintf("%d%d", villain.ID, storyID)),
		Nicknames:  nicknames,
		OtherNames: otherNames,
		CodeNames:  codeNames,
		Roles:      roles,
		Destiny:    destiny,
	}, villain.ID, storyID)

	return nil
}

func (i *importer) getStoryVillainItems(villainID id) []*db.StoryVillain {
	var filtered []*db.StoryVillain
	for _, sv := range i.storyVillains {
		if sv.villain == villainID {
			filtered = append(filtered, sv.item)
			importerStory := i.getStory(sv.story)
			sv.item.Story = importerStory.item
		}
	}
	return filtered
}

func (i *importer) persistVillains(version *db.Version) error {
	villainRepo := db.NewVillainRepository()

	var villainItems []*db.Villain

	for _, importerVillain := range i.villains {
		importerVillain.item.As = i.getStoryVillainItems(importerVillain.ID)
		if len(importerVillain.item.As) == 0 {
			return fmt.Errorf("villain not found in any story")
		}
		villainItems = append(villainItems, importerVillain.item)
	}

	chunks := ChunkSlice(villainItems, db.MaxBulkCreateSize)
	for _, chunk := range chunks {
		_, err := villainRepo.BulkCreate(chunk, version)
		if err != nil {
			return err
		}
	}

	return nil
}
