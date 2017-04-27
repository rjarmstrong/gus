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
	return &Orgs{db: db, Suspender: NewSuspender("users", db)}
}

type Org struct {
	Id      int64 `json:"id"`
	Name    string `json:"name"`
	Type    OrgType `json:"type"`
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
}

type Orgs struct {
	db *sql.DB
	*Suspender
}

type CreateOrgParams struct {
	Name string `json:"name"`
	Type OrgType `json:"type"`
	CustomValidator `json:"-"`
}

func (va CreateOrgParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if govalidator.IsNull(va.Name) {
		return ErrNameRequired
	}
	return nil
}

func (us *Orgs) Create(p CreateOrgParams) (*Org, error) {
	stmt, err := us.db.Prepare("INSERT INTO orgs(name, type, updated, created, deleted) values(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	u := &Org{Name: p.Name, Type: p.Type, Created: time.Now(), Updated: time.Now()}
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
	Id   *int64 `json:"id"`
	Name *string `json:"name"`
	CustomValidator `json:"-"`
}

func (va *UpdateOrgParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if govalidator.IsNull(*va.Name) {
		return ErrInvalid("'name' required.")
	}
	return nil
}

func (us *Orgs) Update(p UpdateOrgParams) error {
	o, err := us.Get(*p.Id)
	if err != nil {
		return err
	}
	ApplyUpdates(o, p)
	stmt, err := us.db.Prepare("UPDATE orgs SET name = ?, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	err = CheckUpdated(stmt.Exec(o.Name, time.Now(), o.Id))
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
	ListArgs
	CustomValidator `json:"-"`
}

func (va *ListOrgsParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	return nil
}

func (us *Orgs) List(p ListOrgsParams) ([]*Org, error) {
	q := "SELECT id, name, type, created, updated from orgs WHERE deleted = 0"
	rows, err := GetRows(us.db, q, p.ListArgs)
	if err != nil {
		return nil, err
	}
	ogs := []*Org{}
	for rows.Next() {
		u := &Org{}
		rows.Scan(&u.Id, &u.Name, &u.Type, &u.Created, &u.Updated)
		ogs = append(ogs, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ogs, nil
}
