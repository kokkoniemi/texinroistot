package importer

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type importerAuthor struct {
	ID   id
	item *db.Author
}

type parsedTranslator struct {
	author  *importerAuthor
	details string
}

var translatorDetailAtEndRegex = regexp.MustCompile(`\s+([0-9][0-9\.\-\s]*p\.?)$`)
var authorAndSeparatorRegex = regexp.MustCompile(`\s*&\s*`)

func (i *importer) loadWriters(storyID id, r row) {
	writers := i.loadAuthorColumn(r, "story_written_by")
	for _, writer := range writers {
		writer.item.IsWriter = true
		i.setWriterForStory(storyID, writer.ID)
	}
}

func (i *importer) loadDrawers(storyID id, r row) {
	drawers := i.loadAuthorColumn(r, "story_drawn_by")
	for _, drawer := range drawers {
		drawer.item.IsDrawer = true
		i.setDrawerForStory(storyID, drawer.ID)
	}
}

func (i *importer) loadTranslators(storyID id, r row) {
	translators := i.loadTranslatorColumn(r)
	for _, translator := range translators {
		translator.author.item.IsTranslator = true
		i.setTranslatorForStory(storyID, translator.author.ID, translator.details)
	}
}

func splitTranslatorFirstNameAndDetails(firstName string) (string, string) {
	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		return "", ""
	}

	// Format: "Renne (2. - 3. p)"
	if openIdx := strings.Index(firstName, "("); openIdx != -1 && strings.HasSuffix(firstName, ")") {
		namePart := strings.TrimSpace(firstName[:openIdx])
		detailPart := strings.TrimSpace(strings.TrimSuffix(firstName[openIdx+1:], ")"))
		return namePart, detailPart
	}

	// Format: "Renne 2. - 3. p"
	matches := translatorDetailAtEndRegex.FindStringSubmatchIndex(firstName)
	if matches != nil && len(matches) >= 4 {
		namePart := strings.TrimSpace(firstName[:matches[0]])
		detailPart := strings.TrimSpace(firstName[matches[2]:matches[3]])
		return namePart, detailPart
	}

	return firstName, ""
}

func splitAuthorColumnNames(raw string) []string {
	parts := strings.Split(raw, ";")
	var names []string

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}

		subParts := authorAndSeparatorRegex.Split(trimmedPart, -1)
		for _, subPart := range subParts {
			trimmedName := strings.TrimSpace(subPart)
			if trimmedName == "" {
				continue
			}
			names = append(names, trimmedName)
		}
	}

	return names
}

func parseAuthorName(name string) (string, string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ""
	}

	if !strings.Contains(name, ",") {
		fields := strings.Fields(name)
		if len(fields) == 0 {
			return "", ""
		}
		if len(fields) == 1 {
			return fields[0], ""
		}
		firstName := strings.Join(fields[:len(fields)-1], " ")
		lastName := fields[len(fields)-1]
		return strings.TrimSpace(firstName), strings.TrimSpace(lastName)
	}

	nameParts := strings.Split(name, ",")
	// Excel format has been inverted:
	// firstName is now on the left side of comma, lastName on the right.
	firstName := strings.TrimSpace(nameParts[0])
	lastName := strings.TrimSpace(strings.Join(nameParts[1:], " "))
	return firstName, lastName
}

func splitTranslatorGroup(raw string) ([]string, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, ""
	}

	sharedDetails := ""
	if openIdx := strings.LastIndex(trimmed, "("); openIdx > 0 && strings.HasSuffix(trimmed, ")") {
		sharedDetails = strings.TrimSpace(strings.TrimSuffix(trimmed[openIdx+1:], ")"))
		trimmed = strings.TrimSpace(trimmed[:openIdx])
	} else {
		matches := translatorDetailAtEndRegex.FindStringSubmatchIndex(trimmed)
		if matches != nil && len(matches) >= 4 {
			sharedDetails = strings.TrimSpace(trimmed[matches[2]:matches[3]])
			trimmed = strings.TrimSpace(trimmed[:matches[0]])
		}
	}

	if trimmed == "" {
		return nil, strings.TrimSpace(sharedDetails)
	}

	names := authorAndSeparatorRegex.Split(trimmed, -1)
	var trimmedNames []string
	for _, name := range names {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		trimmedNames = append(trimmedNames, trimmedName)
	}

	return trimmedNames, strings.TrimSpace(sharedDetails)
}

func (i *importer) loadTranslatorColumn(r row) []parsedTranslator {
	names := strings.Split(r.getValue("story_translated_by"), ";")
	var translators []parsedTranslator

	for _, rawGroup := range names {
		trimmedGroup := strings.TrimSpace(rawGroup)
		if trimmedGroup == "" {
			continue
		}

		isParenthesized := strings.HasPrefix(trimmedGroup, "(") && strings.HasSuffix(trimmedGroup, ")")
		if isParenthesized {
			trimmedGroup = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(trimmedGroup, ")"), "("))
		}

		translatorNames, sharedDetails := splitTranslatorGroup(trimmedGroup)
		for _, trimmedName := range translatorNames {
			parsedName, details := splitTranslatorFirstNameAndDetails(trimmedName)
			if details == "" {
				details = sharedDetails
			}
			if parsedName == "" && details == "" {
				continue
			}

			firstName, lastName := parseAuthorName(parsedName)

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

			translators = append(translators, parsedTranslator{
				author:  author,
				details: strings.TrimSpace(details),
			})
		}
	}

	return translators
}

func (i *importer) getAuthorIndexWithName(firstName string, lastName string) int {
	return slices.IndexFunc(i.authors, func(a *importerAuthor) bool {
		return a.item.FirstName == firstName && a.item.LastName == lastName
	})
}

func (i *importer) getAuthorIndexWithHash(hash string) int {
	return slices.IndexFunc(i.authors, func(a *importerAuthor) bool {
		return a.item.Hash == hash
	})
}

func (i *importer) hasAuthorWithName(firstName string, lastName string) bool {
	return i.getAuthorIndexWithName(firstName, lastName) != -1
}

func (i *importer) getAuthorWithName(firstName string, lastName string) *importerAuthor {
	idx := i.getAuthorIndexWithName(firstName, lastName)
	if idx != -1 {
		return i.authors[idx]
	}
	return nil
}

func (i *importer) getAuthorItemsWithIDs(ids []id) []*db.Author {
	var filtered []*db.Author
	for idx := range i.authors {
		if slices.Contains(ids, i.authors[idx].ID) {
			filtered = append(filtered, i.authors[idx].item)
		}
	}
	return filtered
}

func (i *importer) addAuthor(author *db.Author) *importerAuthor {
	i.totalEntities++

	importerAuthor := &importerAuthor{
		ID:   id(i.totalEntities),
		item: author,
	}

	i.authors = append(i.authors, importerAuthor)

	return importerAuthor
}

func (i *importer) loadAuthorColumn(r row, columnName string) []*importerAuthor {
	names := splitAuthorColumnNames(r.getValue(columnName))
	var authors []*importerAuthor
	for _, n := range names {
		if len(strings.TrimSpace(n)) == 0 {
			continue
		}
		firstName, lastName := parseAuthorName(n)

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

func (i *importer) getAuthorItems() []*db.Author {
	var items []*db.Author

	for index := range i.authors {
		items = append(items, i.authors[index].item)
	}

	return items
}

// setAuthorItems sets persisted Authors to importer after save to db
func (i *importer) setAuthorItems(items []*db.Author) error {
	if len(items) != len(i.authors)%db.MaxBulkCreateSize && len(items) != db.MaxBulkCreateSize {
		return fmt.Errorf("Mismatch in the number of Authors")
	}
	for index := range items {
		importerIndex := i.getAuthorIndexWithHash(items[index].Hash)
		if importerIndex == -1 {
			return fmt.Errorf("Tried to set unknown Author")
		}
		i.authors[importerIndex].item = items[index]
	}
	return nil
}

// persistAuthors writes Authors loaded in importer to db
func (i *importer) persistAuthors(version *db.Version) error {
	var err error
	authorRepo := db.NewAuthorRepository()
	chunks := ChunkSlice(i.getAuthorItems(), db.MaxBulkCreateSize)
	for _, chunk := range chunks {
		authors, err := authorRepo.BulkCreate(chunk, version)
		if err != nil {
			return err
		}
		err = i.setAuthorItems(authors)
		if err != nil {
			return err
		}
	}
	return err
}
