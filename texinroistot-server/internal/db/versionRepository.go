package db

import "fmt"

type versionRepo struct{}

const setVersionActiveSQL = `
UPDATE versions
SET is_active = false
WHERE NOT id = $1;

UPDATE versions
SET is_active = true
WHERE id = $1
`

// SetActive implements VersionRepository.
func (*versionRepo) SetActive(versionID int) error {
	_, err := Execute(setVersionActiveSQL, versionID)
	return err
}

const getActiveVersionSQL = `
SELECT id, created_at, is_active
FROM versions
WHERE is_active = TRUE;
`

func (*versionRepo) GetActive() (*Version, error) {
	rows, err := Query(getActiveVersionSQL)
	if err != nil {
		return nil, err
	}
	var v Version
	count := 0
	for rows.Next() {
		if err = rows.Scan(&v.ID, &v.CreatedAt, &v.IsActive); err != nil {
			return nil, err
		}
		count++
	}
	if count != 1 {
		return nil, fmt.Errorf("invalid number of active versions: %d", count)
	}
	return &v, nil
}

const getVersionStatsSQL = `
SELECT
	(SELECT COUNT(*) FROM villains WHERE version = $1) AS villains,
	(SELECT COUNT(*) FROM stories WHERE version = $1) AS stories,
	(SELECT COUNT(*) FROM authors WHERE version = $1 AND is_drawer = true) AS drawers,
	(SELECT COUNT(*) FROM authors WHERE version = $1 AND is_writer = true) AS writers,
	(SELECT COUNT(*) FROM authors WHERE version = $1 AND is_inventor = true) AS translators;
`

// GetStats implements VersionRepository.
func (*versionRepo) GetStats(versionID int) (*VersionStats, error) {
	rows, err := Query(getVersionStatsSQL, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats VersionStats
	for rows.Next() {
		if err = rows.Scan(
			&stats.Villains,
			&stats.Stories,
			&stats.Drawers,
			&stats.Writers,
			&stats.Translators,
		); err != nil {
			return nil, err
		}
		return &stats, nil
	}

	return &stats, nil
}

const createVersionSQL = `INSERT INTO versions(is_active) VALUES(false);`

// Create implements VersionRepository.
func (v *versionRepo) Create(version Version) (*Version, error) {
	_, err := Execute(createVersionSQL)
	if err != nil {
		return nil, err
	}

	versions, err := v.List()
	if err != nil {
		return nil, err
	}
	lastVersion := versions[len(versions)-1]

	return lastVersion, nil
}

const readVersionSQL = `
SELECT
	id,
	created_at,
	is_active
FROM versions
WHERE id = $1;
`

// Read implements VersionRepository.
func (*versionRepo) Read(versionID int) (*Version, error) {
	rows, err := Query(readVersionSQL, versionID)
	if err != nil {
		return nil, err
	}
	var v Version
	for rows.Next() {
		if err = rows.Scan(&v.ID, &v.CreatedAt, &v.IsActive); err != nil {
			return nil, err
		}
	}
	if &v == nil {
		return nil, fmt.Errorf("version not found")
	}
	return &v, nil
}

const listVersionsSQL = `
SELECT
	id,
	created_at,
	is_active
FROM versions
ORDER BY created_at;
`

// List implements VersionRepository.
func (*versionRepo) List() ([]*Version, error) {
	rows, err := Query(listVersionsSQL)
	if err != nil {
		return nil, err
	}
	var versions []*Version

	for rows.Next() {
		var v Version
		if err = rows.Scan(&v.ID, &v.CreatedAt, &v.IsActive); err != nil {
			return nil, err
		}
		versions = append(versions, &v)
	}

	return versions, nil
}

const removeVersionSQL = `
DELETE FROM versions
WHERE id = $1;
`

// Remove implements VersionRepository.
func (*versionRepo) Remove(versionID int) error {
	_, err := Execute(removeVersionSQL, versionID)
	return err
}

func NewVersionRepository() VersionRepository {
	return &versionRepo{}
}
