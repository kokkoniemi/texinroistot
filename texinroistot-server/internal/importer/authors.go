package importer

import (
	"slices"
	"strings"

	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type importerAuthor struct {
	ID   id
	item *db.Author
}

func (i *sprdImporter) importWriters(storyID id, r row) {
	writers := i.loadAuthorColumn(r, "story_written_by")
	for _, writer := range writers {
		writer.item.IsWriter = true
		i.setWriterForStory(storyID, writer.ID)
	}
}

func (i *sprdImporter) importDrawer(storyID id, r row) {
	drawers := i.loadAuthorColumn(r, "story_drawn_by")
	for _, drawer := range drawers {
		drawer.item.IsDrawer = true
		i.setDrawerForStory(storyID, drawer.ID)
	}
}

func (i *sprdImporter) importInventor(storyID id, r row) {
	inventors := i.loadAuthorColumn(r, "story_invented_by")
	for _, inventor := range inventors {
		inventor.item.IsInventor = true
		i.setInventorForStory(storyID, inventor.ID)
	}
}

func (i *sprdImporter) getAuthorIndexWithName(firstName string, lastName string) int {
	return slices.IndexFunc(i.authors, func(a *importerAuthor) bool {
		return a.item.FirstName == firstName && a.item.LastName == lastName
	})
}

func (i *sprdImporter) hasAuthorWithName(firstName string, lastName string) bool {
	return i.getAuthorIndexWithName(firstName, lastName) != -1
}

func (i *sprdImporter) getAuthorWithName(firstName string, lastName string) *importerAuthor {
	idx := i.getAuthorIndexWithName(firstName, lastName)
	if idx != -1 {
		return i.authors[idx]
	}
	return nil
}

func (i *sprdImporter) addAuthor(author *db.Author) *importerAuthor {
	i.totalEntities++

	importerAuthor := &importerAuthor{
		ID:   id(i.totalEntities),
		item: author,
	}

	i.authors = append(i.authors, importerAuthor)

	return importerAuthor
}

func (i *sprdImporter) loadAuthorColumn(r row, columnName string) []*importerAuthor {
	names := strings.Split(r.getValue(columnName), ";")
	var authors []*importerAuthor
	for _, n := range names {
		if len(n) == 0 {
			continue
		}
		nameParts := strings.Split(n, ",")
		var firstName string
		var lastName string
		if len(nameParts) == 1 {
			firstName = strings.TrimSpace(nameParts[0])
		} else {
			lastName = strings.TrimSpace(nameParts[0])
			firstName = strings.TrimSpace(strings.Join(nameParts[1:], " "))
		}

		var author *importerAuthor

		if !i.hasAuthorWithName(firstName, lastName) {
			author = i.addAuthor(&db.Author{
				Hash:      crypt.Hash(firstName + lastName),
				FirstName: firstName,
				LastName:  lastName,
			})
		} else {
			author = i.getAuthorWithName(firstName, lastName)
		}

		authors = append(authors, author)
	}

	return authors
}
