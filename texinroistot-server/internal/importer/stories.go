package importer

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

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

func (i *importer) loadStory(r row) (id, error) {
	orderNumStr := r.getValue("story_order_num")

	orderNum := -1

	if len(orderNumStr) > 0 {
		parsed, err := strconv.Atoi(strings.TrimSpace(orderNumStr))
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

func (i *importer) getStory(storyID id) *importerStory {
	storyIdx := slices.IndexFunc(i.stories, func(s *importerStory) bool {
		return s.ID == storyID
	})
	if storyIdx != -1 {
		return i.stories[storyIdx]
	}
	return nil
}

func (i *importer) getStoryWithHash(hash string) *importerStory {
	storyIdx := slices.IndexFunc(i.stories, func(s *importerStory) bool {
		return s.item.Hash == hash
	})
	if storyIdx != -1 {
		return i.stories[storyIdx]
	}
	return nil
}

func (i *importer) hasStoryWithHash(hash string) bool {
	return i.getStoryWithHash(hash) != nil
}

func (i *importer) addStory(story *db.Story) *importerStory {
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

func (i *importer) setWriterForStory(storyID id, writerID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.writers, writerID) {
		story.writers = append(story.writers, writerID)
	}
}

func (i *importer) setDrawerForStory(storyID id, drawerID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.drawers, drawerID) {
		story.drawers = append(story.drawers, drawerID)
	}
}

func (i *importer) setInventorForStory(storyID id, inventorID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.inventors, inventorID) {
		story.inventors = append(story.inventors, inventorID)
	}
}

func (i *importer) setStoryItems(items []*db.Story) error {
	for idx := range i.stories {
		for _, story := range items {
			if i.stories[idx].item.Hash == story.Hash {
				i.stories[idx].item = story
				break
			}
			// TODO: must be checked against chunked story items
			//if idx2 == len(items)-1 && i.stories[idx].item.Hash != story.Hash {
			//	return fmt.Errorf("setStoryItems failed")
			//}
		}
	}

	return nil

}

func (i *importer) persistStories(version *db.Version) error {
	var err error

	// set authors for stories in a loop
	// set storyPublications for stories in a loop
	var storyItems []*db.Story

	for _, importerStory := range i.stories {
		importerStory.item.WrittenBy = i.getAuthorItemsWithIDs(importerStory.writers)
		importerStory.item.DrawnBy = i.getAuthorItemsWithIDs(importerStory.drawers)
		importerStory.item.InventedBy = i.getAuthorItemsWithIDs(importerStory.inventors)

		storyID := importerStory.ID
		importerStoryPublications := i.getStoryPublications(storyID)
		for _, sp := range importerStoryPublications {
			importerStory.item.Publications = append(importerStory.item.Publications,
				&db.StoryPublication{
					Title: sp.title,
					In:    i.getPublicationWithID(sp.publication).item,
				})
		}

		storyItems = append(storyItems, importerStory.item)
	}

	// create chunks of stories
	storyRepo := db.NewStoryRepository()
	chunks := ChunkSlice(storyItems, db.MaxBulkCreateSize)
	for _, chunk := range chunks {
		stories, err := storyRepo.BulkCreate(chunk, version)
		if err != nil {
			return err
		}
		err = i.setStoryItems(stories)
		if err != nil {
			return err
		}
	}

	// bulkCreate stories one chunk at a time

	return err
}
