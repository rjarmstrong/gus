package gus

import (
	"database/sql"
	"time"
	"github.com/asaskevich/govalidator"
)

var (
	ErrNameRequired error = ErrInvalid("'name' required.")
)

func NewOrgs(db *sql.DB) *Orgs {
	return &Orgs{db: db}
}

type Org struct {
	Id int64 `json:"id"`
	Name string `json:"name"`
	Type OrgType `json:"type"`
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
}

type Orgs struct {
	db *sql.DB
}

type CreateOrgParams struct {
	Name string `json:"name"`
	Type OrgType `json:"type"`
}

func (cp CreateOrgParams) Validate() error {
	if govalidator.IsNull(cp.Name) {
		return ErrNameRequired
	}
	return nil
}

func (us *Orgs) Create(p CreateOrgParams) (*Org, error) {
	stmt, err := us.db.Prepare("INSERT INTO orgs(name, type, updated, created, deleted) values(?,?, ?,?,?)")
	if err != nil {
		return nil, err
	}
	u := &Org{Name: p.Name, Type:p.Type, Created: time.Now(), Updated: time.Now()}
	res, err := stmt.Exec(u.Name, u.Type, u.Updated, u.Created, 0)
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.Id = id
	return u, nil
}

func (us *Orgs) Get(id int64) (*Org, error) {
	stmt, err := us.db.Prepare("SELECT id, name, type, created, updated from orgs WHERE id = ? AND deleted = 0 LIMIT 1")
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(id)
	var u Org
	err = CheckNotFound(row.Scan(&u.Id, &u.Name, &u.Type, &u.Created, &u.Updated))
	if err != nil {
		return nil, err
	}
	return &u, err
}

type UpdateOrgParams struct {
	Id   int64 `json:"id"`
	Name string `json:"name"`
}

func (up *UpdateOrgParams) Validate() error {
	if govalidator.IsNull(up.Name) {
		return ErrInvalid("'name' required.")
	}
	return nil
}

func (us *Orgs) Update(u UpdateOrgParams) error {
	stmt, err := us.db.Prepare("UPDATE orgs SET name = ?, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	err = CheckUpdated(stmt.Exec(u.Name, time.Now(), u.Id))
	if err != nil {
		return err
	}
	return nil
}

func (us *Orgs) Delete(id int64) error {
	stmt, err := us.db.Prepare("UPDATE orgs SET deleted = 1, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	return CheckUpdated(stmt.Exec(time.Now(), id))
}

type ListOrgsParams struct {
}

func (pm *ListOrgsParams) Validate() error {
	return nil
}

func (us *Orgs) List(p ListOrgsParams) ([]*Org, error) {
	stmt, err := us.db.Prepare("SELECT id, name, type, created, updated from orgs WHERE deleted = 0 ORDER by updated DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orgs := []*Org{}
	for rows.Next() {
		u := &Org{}
		rows.Scan(&u.Id, &u.Name, &u.Created, &u.Updated)
		orgs = append(orgs, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}
