package db

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
