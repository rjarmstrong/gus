package gus

import (
	"database/sql"
	"github.com/asaskevich/govalidator"
	"time"
)

var (
	ErrNameRequired error = ErrInvalid("'name' required.")
)

type OrgType int64

func NewOrgs(db *sql.DB) *Orgs {
	return &Orgs{db: db, Suspender: NewSuspender("orgs", db)}
}

type Org struct {
	Id   int64   `json:"id"`
	Name string  `json:"name"`
	Type OrgType `json:"type"`

	Street   string `json:"street"`
	Suburb   string `json:"suburb"`
	Town     string `json:"town"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`

	Updated   int64 `json:"updated"`
	Created   int64 `json:"created"`
	Suspended bool  `json:"suspended"`
}

type Orgs struct {
	db *sql.DB
	*Suspender
}

type CreateOrgParams struct {
	Name string  `json:"name"`
	Type OrgType `json:"type"`

	Street   string `json:"street"`
	Suburb   string `json:"suburb"`
	Town     string `json:"town"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`

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
	stmt, err := us.db.Prepare("INSERT INTO orgs(name, type, street, suburb, town, postcode , country, updated, created, deleted, suspended) values(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	u := &Org{Name: p.Name, Type: p.Type, Street: p.Street, Suburb: p.Suburb, Town: p.Town, Postcode: p.Postcode, Country: p.Country, Created: Milliseconds(time.Now()), Updated: Milliseconds(time.Now())}
	res, err := stmt.Exec(u.Name, u.Type, u.Street, u.Suburb, u.Town, u.Postcode, u.Country, u.Updated, u.Created, 0, false)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.Id = id
	return u, nil
}

func (us *Orgs) Get(id int64) (*Org, error) {
	stmt, err := us.db.Prepare("SELECT id, name, type, street, suburb, town, postcode, country, created, updated, suspended from orgs WHERE id = ? AND deleted = 0 LIMIT 1")
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(id)
	var u Org
	var suspended int8
	err = CheckNotFound(row.Scan(&u.Id, &u.Name, &u.Type, &u.Street, &u.Suburb, &u.Town, &u.Postcode, &u.Country,
		&u.Created, &u.Updated, &suspended))
	if err != nil {
		return nil, err
	}
	u.Suspended = suspended > 0
	return &u, err
}

type UpdateOrgParams struct {
	Id              *int64  `json:"id"`
	Name            *string `json:"name"`
	Street          *string `json:"street"`
	Suburb          *string `json:"suburb"`
	Town            *string `json:"town"`
	Postcode        *string `json:"postcode"`
	Country         *string `json:"country"`
	CustomValidator `json:"-"`
}

func (va *UpdateOrgParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if va.Name != nil && *va.Name == "" {
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
	stmt, err := us.db.Prepare("UPDATE orgs SET name = ?, street = ?, suburb = ?, town = ?, postcode = ?, country = ?, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	err = CheckUpdated(stmt.Exec(o.Name, o.Street, o.Suburb, o.Town, o.Postcode, o.Country, Milliseconds(time.Now()), o.Id))
	if err != nil {
		return err
	}
	return nil
}

type ListOrgsParams struct {
	Name string `json:"name"`
	ListArgs
	CustomValidator `json:"-"`
}

func (va *ListOrgsParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	return nil
}

type OrgListResponse struct {
	ListArgs
	Total int64  `json:"total"`
	Items []*Org `json:"items"`
}

func (us *Orgs) List(p ListOrgsParams) (*OrgListResponse, error) {
	q := "SELECT id, name, type, street, suburb, town, postcode, country, created, updated, suspended from orgs WHERE 1"
	countq := "SELECT count(id) FROM orgs WHERE 1"

	args := []interface{}{}
	if !p.Deleted {
		q += " AND deleted = 0"
		countq += " AND deleted = 0"
	}
	if p.Name != "" {
		q += " AND name like ?"
		countq += " AND name like ?"
		name := "%" + p.Name + "%"
		args = append(args, name, name)
	}
	rows, err := GetRows(us.db, q, &p.ListArgs, args...)
	if err != nil {
		return nil, err
	}
	row := us.db.QueryRow(countq, args...)
	var total int64
	err = row.Scan(&total)
	if err != nil {
		return nil, err
	}
	ogs := []*Org{}
	for rows.Next() {
		u := &Org{}
		var suspended int
		rows.Scan(&u.Id, &u.Name, &u.Type, &u.Street, &u.Suburb, &u.Town, &u.Postcode, &u.Country,
			&u.Created, &u.Updated, &suspended)
		u.Suspended = suspended > 0
		ogs = append(ogs, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &OrgListResponse{
		Total: total,
		Items: ogs,
		ListArgs: ListArgs{
			Size:      p.Size,
			Page:      p.Page,
			Direction: p.Direction,
			OrderBy:   p.OrderBy,
			Deleted:   p.Deleted,
		}}, nil
}
