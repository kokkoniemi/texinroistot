package db

import (
	"fmt"
)

type villainRepo struct{}

func (v *villainRepo) BulkCreate(villains []*Villain, version *Version) ([]*Villain, error) {
	// save villains
	if len(villains) > MaxBulkCreateSize {
		return nil, fmt.Errorf("max number of %d villains exceeded", MaxBulkCreateSize)
	}

	var villainValues [][]interface{}

	for _, v := range villains {
		villainValues = append(villainValues, []interface{}{
			v.Hash,
			ArrayParam(v.Ranks),
			ArrayParam(v.FirstNames),
			v.LastName,
			version.ID,
		})
	}
	numRows, err := BulkInsertTxn(bulkInsertParams{
		Table:   "villains",
		Columns: []string{"hash", "ranks", "first_names", "last_name", "version"},
		Values:  villainValues,
	})
	if err != nil {
		return nil, err
	}

	villains, err = v.setIDsFromDB(villains, numRows)
	if err != nil {
		return nil, err
	}

	// save story villains
	var storyVillainValues [][]interface{}

	for _, v := range villains {
		for _, sv := range v.As {
			if v.ID == 0 {
				fmt.Println("v: ", v)
			}
			storyVillainValues = append(storyVillainValues, []interface{}{
				v.ID,
				sv.Story.ID, // TODO: make sure that this is found
				sv.Hash,
				ArrayParam(sv.Nicknames),
				ArrayParam(sv.Aliases),
				ArrayParam(sv.Destiny),
				ArrayParam(sv.Roles),
			})
		}
	}

	_, err = BulkInsertTxn(bulkInsertParams{
		Table:   "villains_in_stories",
		Columns: []string{"villain", "story", "hash", "nicknames", "aliases", "destiny", "roles"},
		Values:  storyVillainValues,
	})
	if err != nil {
		return nil, err
	}

	// list created villains

	return nil, nil
}

const setVillainIDsSQL = `
SELECT
	v.id,
	v.hash
FROM villains as v
ORDER BY v.id DESC
LIMIT %v;
`

func (v *villainRepo) setIDsFromDB(villains []*Villain, savedRows int64) ([]*Villain, error) {
	queryString := fmt.Sprintf(setVillainIDsSQL, savedRows)
	rows, err := Query(queryString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var row struct {
			ID   int
			Hash string
		}
		if err = rows.Scan(&row.ID, &row.Hash); err != nil {
			return nil, err
		}

		for idx := range villains {
			if villains[idx].Hash == row.Hash {
				villains[idx].ID = row.ID
				break
			}
			if idx == len(villains)-1 && villains[idx].Hash != row.Hash {
				return nil, fmt.Errorf("matching villain not found from db")
			}
		}
	}

	return villains, nil
}

func NewVillainRepository() VillainRepository {
	return &villainRepo{}
}
