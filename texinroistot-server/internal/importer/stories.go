package importer

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type importerStory struct {
	ID        id
	item      *db.Story
	writers   []id
	drawers   []id
	inventors []id
}

func (i *Importer) importStory(r row) (id, error) {
	orderNumStr := r.getValue("story_order_num")

	orderNum := -1

	if len(orderNumStr) > 0 {
		parsed, err := strconv.Atoi(orderNumStr)
		if err != nil {
			return 0, err
		}
		orderNum = parsed
	}

	hash := ""
	if orderNum != 0 {
		hash = crypt.Hash(fmt.Sprintf("%v", orderNum))
	} else {
		hash = crypt.Hash(r.getValue("story_title"))
	}

	var story *importerStory

	if !i.hasStoryWithHash(hash) {
		story = i.addStory(&db.Story{
			Hash:        hash,
			OrderNumber: orderNum,
		})
	} else {
		story = i.getStoryWithHash(hash)
	}

	if story == nil {
		return id(0), fmt.Errorf("story not found")
	}

	return story.ID, nil
}

func (i *Importer) getStory(storyID id) *importerStory {
	storyIdx := slices.IndexFunc(i.stories, func(s *importerStory) bool {
		return s.ID == storyID
	})
	if storyIdx != -1 {
		return i.stories[storyIdx]
	}
	return nil
}

func (i *Importer) getStoryWithHash(hash string) *importerStory {
	storyIdx := slices.IndexFunc(i.stories, func(s *importerStory) bool {
		return s.item.Hash == hash
	})
	if storyIdx != -1 {
		return i.stories[storyIdx]
	}
	return nil
}

func (i *Importer) hasStoryWithHash(hash string) bool {
	return i.getStoryWithHash(hash) != nil
}

func (i *Importer) addStory(story *db.Story) *importerStory {
	i.totalEntities++

	importerStory := &importerStory{
		ID:        id(i.totalEntities),
		item:      story,
		writers:   []id{},
		drawers:   []id{},
		inventors: []id{},
	}

	i.stories = append(i.stories, importerStory)

	return importerStory
}

func (i *Importer) setWriterForStory(storyID id, writerID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.writers, writerID) {
		story.writers = append(story.writers, writerID)
	}
}

func (i *Importer) setDrawerForStory(storyID id, drawerID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.drawers, drawerID) {
		story.drawers = append(story.drawers, drawerID)
	}
}

func (i *Importer) setInventorForStory(storyID id, inventorID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.inventors, inventorID) {
		story.inventors = append(story.inventors, inventorID)
	}
}
