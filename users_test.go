package gus

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var cp = CreateUserParams{Email: "richard.armstrong@gimanzo.com", ProviderId: 1}

func TestUsers_Create(t *testing.T) {
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)
	assert.True(t, u.Id > 0)

	// Should not allow create for existing email
	u, _, err = us.Create(cp)
	assert.Error(t, err)
}

func TestUsers_Get(t *testing.T) {
	Seed(db)
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	u, err = us.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)
}

func TestUsers_Update(t *testing.T) {
	Seed(db)
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	up := UpdateUserParams{Id: u.Id, Email:"donkey@kong.com", FirstName:"Donkey", LastName:"Kong", Phone : "0345345"}
	err = us.Update(up)
	ErrIf(t, err)
	u, _ = us.Get(u.Id)
	assert.Equal(t, up.Email, u.Email)
	assert.Equal(t, up.FirstName, u.FirstName)
	assert.Equal(t, up.LastName, u.LastName)
	assert.Equal(t, up.Phone, u.Phone)

	// Should not update to existing email
	u, _, err = us.Create(cp)
	up.Id = u.Id
	err = us.Update(up)
	assert.Error(t, err)

	// Should not allow update of non-existing record
	up.Id = 33453453
	err = us.Update(up)
	assert.Error(t, err)
}

func TestUsers_Delete(t *testing.T) {
	Seed(db)
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	err = us.Delete(u.Id)
	ErrIf(t, err)
	u, err = us.Get(u.Id)
	assert.Nil(t, u)
	assert.Error(t, err)
}