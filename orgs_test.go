package gus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var corg = CreateOrgParams{
	Name:   "Trainers Inc.",
	Street: "23 Dinbat", Suburb: "Bollbub", Town: "Zingford", Postcode: "ABC23", Country: "Bingaloo",
}

func TestOrgs_Create(t *testing.T) {
	u, err := orgsv.Create(corg)
	ErrIf(t, err)
	assert.Equal(t, u.Name, corg.Name)
	assert.True(t, u.Id > 0)

	// Get
	u, err = orgsv.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Name, corg.Name)
	assert.Equal(t, u.Suburb, corg.Suburb)

	orgs, err := orgsv.List(ListOrgsParams{})
	ErrIf(t, err)
	assert.Equal(t, 1, len(orgs.Items))
	assert.Equal(t, corg.Name, orgs.Items[0].Name)
}

func TestOrgs_Update(t *testing.T) {
	u, err := orgsv.Create(corg)
	ErrIf(t, err)
	name := "New Name"
	street := "New Street"
	up := UpdateOrgParams{Id: &u.Id, Name: &name, Street: &street}
	err = orgsv.Update(up)
	ErrIf(t, err)
	u, _ = orgsv.Get(u.Id)
	assert.Equal(t, *up.Name, u.Name)
	assert.Equal(t, *up.Street, u.Street)

	// Should not allow update of non-existing record
	id := int64(33453453)
	up.Id = &id
	err = orgsv.Update(up)
	assert.Error(t, err)
}

func TestOrgs_Delete(t *testing.T) {
	u, err := orgsv.Create(corg)
	ErrIf(t, err)
	err = orgsv.Delete(u.Id)
	ErrIf(t, err)
	u, err = orgsv.Get(u.Id)
	assert.Nil(t, u)
	assert.Error(t, err)
}
