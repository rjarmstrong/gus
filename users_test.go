package gus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var cp = SignUpParams{Email: "user@mail.com", OrgId: 1}

func TestUsers_Create(t *testing.T) {
	u, _, err := us.SignUp(cp)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)
	assert.True(t, u.Id > 0)

	// Should not allow create for existing email
	_, _, err = us.SignUp(cp)
	assert.Error(t, err)

	// Get
	u, err = us.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)

	// List
	users, err := us.List(ListUsersParams{OrgId: 1})
	ErrIf(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, int64(1), users[0].Id)
	assert.Equal(t, cp.Email, users[0].Email)

	i := 5
	for i > 0 {
		u, _, err = us.SignUp(SignUpParams{Email: fmt.Sprintf("%d@mail.com", i)})
		ErrIf(t, err)
		i--
	}
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 3},
	})
	assert.Equal(t, 3, len(users))

	// 2nd Page shorter than size
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 4, Page: 1, OrderBy: "id", Direction: DirectionAsc},
	})
	ErrIf(t, err)
	assert.Equal(t, 2, len(users))

	// Order by id desc
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 20, Page: 0, OrderBy: "id", Direction: DirectionDesc},
	})
	ErrIf(t, err)
	assert.Equal(t, int64(6), users[0].Id)

	// Order by id asc
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 20, Page: 0, OrderBy: "id", Direction: DirectionAsc},
	})
	ErrIf(t, err)
	assert.Equal(t, int64(1), users[0].Id)
}

func TestUsers_Update(t *testing.T) {
	cp.Email = "update@mail.com"
	u, _, err := us.SignUp(cp)
	ErrIf(t, err)
	email := "donkey@kong.com"
	fname := "Donkey"
	phone := "0345345"
	up := UpdateUserParams{Id: &u.Id, Email: &email, FirstName: &fname, Phone: &phone}
	err = us.Update(up)
	ErrIf(t, err)
	u, _ = us.Get(u.Id)
	assert.Equal(t, *up.Email, u.Email)
	assert.Equal(t, *up.FirstName, u.FirstName)
	assert.Equal(t, *up.Phone, u.Phone)
	// untouched
	assert.Equal(t, cp.LastName, u.LastName)

	// Should not update to existing email
	cp.Email = "update2@mail.com"
	u2, _, err := us.SignUp(cp)
	ErrIf(t, err)
	up.Id = &u2.Id
	err = us.Update(up)
	assert.Error(t, err)

	// Should not allow update of non-existing record
	id := int64(33453453)
	up.Id = &id
	err = us.Update(up)
	assert.Error(t, err)
}

func TestUsers_Delete(t *testing.T) {
	cp.Email = "delete@mail.com"
	u, _, err := us.SignUp(cp)
	id := u.Id
	ErrIf(t, err)
	err = us.Delete(id)
	ErrIf(t, err)
	u, err = us.Get(id)
	assert.Nil(t, u)
	assert.Error(t, err)
	err = us.UnDelete(id)
	ErrIf(t, err)
	u, err = us.Get(id)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)
}

func TestUsers_AssignRole(t *testing.T) {
	cp.Email = "assign@mail.com"
	u, _, err := us.SignUp(cp)
	ErrIf(t, err)
	role := Role(55)
	err = us.AssignRole(AssignRoleParams{Id: &u.Id, Role: &role})
	ErrIf(t, err)
	u, err = us.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Role, role)
	err = us.AssignRole(AssignRoleParams{Id: &u.Id, Role: nil})
	ErrIf(t, err)
	u, err = us.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Role, Role(0))
}

func TestUsers_SignIn(t *testing.T) {
	cp.Email = "suspend@mail.com"
	u, tempPassword, err := us.SignUp(cp)
	id := u.Id
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: tempPassword})
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Email: cp.Email, Password: tempPassword})
	ErrIf(t, err)
	err = us.Suspend(id)
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: tempPassword})
	assert.Error(t, err)
	err = us.Restore(id)
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: tempPassword})
	ErrIf(t, err)
	ErrIf(t, orgsv.Suspend(cp.OrgId))
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: tempPassword})
	assert.Error(t, err)
}

func TestUsers_Lock(t *testing.T) {
	username := "lock@mail.com"
	assert.False(t, us.isLocked(username))
	assert.False(t, us.isLocked(username))
	assert.False(t, us.isLocked(username))
	assert.False(t, us.isLocked(username))
	assert.False(t, us.isLocked(username))
	assert.True(t, us.isLocked(username))
	time.Sleep(time.Millisecond * time.Duration(1010))
	assert.False(t, us.isLocked(username))
}
