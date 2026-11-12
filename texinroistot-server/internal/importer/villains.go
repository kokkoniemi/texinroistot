package importer

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

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
			// italic markup must not change the hash
			strings.Join(stripItalicMarkupAll(roles), "") +
			strings.Join(stripItalicMarkupAll(destiny), ""))
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
		FirstNames: firstNames,
		LastName:   lastName,
		// the ranks of this row only; the villain's ranks gather from all rows
		Ranks:      slices.Clone(villainRanks),
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

// applyLatestStoryNames sets a villain's own name from the names their stories
// give (each story keeps its own name in StoryVillain):
//   - The last name is the one from the latest story (highest story order)
//     that has one: a villain may be known by one name early on and revealed
//     to have another later ("Kid Rodelo" turns out to be Philip Rainey).
//   - The first names are gathered, in story order, from the stories that
//     share that last name or give none, so Raymond "Ray" Clemmons and Ray
//     Clemmons stay one name; a first name already contained in another, like
//     "Ray" in Raymond "Ray", is not repeated.
//   - Spelling variants (a letter or two apart) count as the same name and are
//     shown in parentheses after it: Mickey (Miky) Welt, Samuel Quantrill
//     (Quantrell).
//   - With a clearly different last name in the latest story, earlier first
//     names are left out: Kid Rodelo's card shows Philip Rainey.
func applyLatestStoryNames(villain *db.Villain) {
	stories := make([]*db.StoryVillain, len(villain.As))
	copy(stories, villain.As)
	sort.SliceStable(stories, func(a, b int) bool {
		return storyOrder(stories[a]) < storyOrder(stories[b])
	})

	lastName := ""
	for _, sv := range stories {
		if name := strings.TrimSpace(sv.LastName); name != "" {
			lastName = name
		}
	}

	var firstNames, lastNameVariants []string
	for _, sv := range stories {
		storyLastName := strings.TrimSpace(sv.LastName)
		if storyLastName != "" && !strings.EqualFold(storyLastName, lastName) {
			if !isSpellingVariant(storyLastName, lastName) {
				continue
			}
			if !slices.ContainsFunc(lastNameVariants, func(v string) bool {
				return strings.EqualFold(v, storyLastName)
			}) {
				lastNameVariants = append(lastNameVariants, storyLastName)
			}
		}
		for _, name := range sv.FirstNames {
			if name = strings.TrimSpace(name); name != "" {
				firstNames = append(firstNames, name)
			}
		}
	}

	if len(firstNames) > 0 {
		villain.FirstNames = withVariantsInParentheses(withoutContainedNames(firstNames))
	}
	if lastName != "" {
		villain.LastName = withVariantsInParentheses(append([]string{lastName}, lastNameVariants...))[0]
	}
}

// isSpellingVariant reports whether two different names are spellings of the
// same name: at most two letters apart, and at most a third of the longer
// name (Mickey/Miky, Quantrill/Quantrell; not Rodelo/Rainey).
func isSpellingVariant(a, b string) bool {
	ra, rb := []rune(strings.ToLower(a)), []rune(strings.ToLower(b))
	if slices.Equal(ra, rb) {
		return false
	}
	distance := editDistance(ra, rb)
	return distance <= 2 && distance*3 <= max(len(ra), len(rb))
}

// editDistance is the Levenshtein distance between two names.
func editDistance(a, b []rune) int {
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

// withVariantsInParentheses folds each name into an earlier one it is a
// spelling variant of, written in parentheses after it: ["Mickey", "Miky"]
// becomes ["Mickey (Miky)"]. Other names are kept as they are, in order.
func withVariantsInParentheses(names []string) []string {
	var kept []string
	variants := map[int][]string{}
	for _, name := range names {
		folded := false
		for idx, keptName := range kept {
			if isSpellingVariant(name, keptName) {
				variants[idx] = append(variants[idx], name)
				folded = true
				break
			}
		}
		if !folded {
			kept = append(kept, name)
		}
	}
	for idx, vs := range variants {
		kept[idx] = fmt.Sprintf("%s (%s)", kept[idx], strings.Join(vs, ", "))
	}
	return kept
}

func storyOrder(sv *db.StoryVillain) int {
	if sv.Story == nil {
		return 0
	}
	return sv.Story.OrderNumber
}

// withoutContainedNames drops names that repeat another one, either exactly
// or as a whole word inside it ("Ray" in `Raymond "Ray"`), keeping the order.
func withoutContainedNames(names []string) []string {
	words := func(name string) []string {
		return strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
	}
	containedIn := func(name, other string) bool {
		nameWords, otherWords := words(name), words(other)
		if len(nameWords) == 0 || len(nameWords) > len(otherWords) {
			return false
		}
		for start := 0; start+len(nameWords) <= len(otherWords); start++ {
			if slices.Equal(nameWords, otherWords[start:start+len(nameWords)]) {
				return true
			}
		}
		return false
	}

	var kept []string
	for idx, name := range names {
		repeated := false
		for other := range names {
			if other == idx {
				continue
			}
			// an exact duplicate keeps its first occurrence
			if strings.EqualFold(names[other], name) {
				if other < idx {
					repeated = true
				}
			} else if containedIn(name, names[other]) {
				repeated = true
			}
			if repeated {
				break
			}
		}
		if !repeated {
			kept = append(kept, name)
		}
	}
	return kept
}

func (i *importer) persistVillains(version *db.Version) error {
	villainRepo := db.NewVillainRepository()

	var villainItems []*db.Villain

	for _, importerVillain := range i.villains {
		importerVillain.item.As = i.getStoryVillainItems(importerVillain.ID)
		if len(importerVillain.item.As) == 0 {
			return fmt.Errorf("villain not found in any story")
		}
		applyLatestStoryNames(importerVillain.item)
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
