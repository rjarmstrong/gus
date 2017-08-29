package gus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var cp = SignUpParams{Email: "user@mail.com", OrgId: 1}

func TestUsers_SignUp(t *testing.T) {
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
	users, err := us.List(ListUsersParams{UserFilters: UserFilters{OrgId: 1}})
	ErrIf(t, err)
	assert.Equal(t, 1, len(users.Items))
	assert.Equal(t, int64(1), users.Total)
	assert.Equal(t, int64(1), users.Items[0].Id)
	assert.Equal(t, cp.Email, users.Items[0].Email)

	i := 5
	for i > 0 {
		u, _, err = us.SignUp(SignUpParams{Email: fmt.Sprintf("%d@mail.com", i)})
		ErrIf(t, err)
		i--
	}
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 3},
	})
	assert.Equal(t, 3, len(users.Items))
	assert.Equal(t, int64(6), users.Total)

	// 2nd Page shorter than size
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 4, Page: 1, OrderBy: "id", Direction: DirectionAsc},
	})
	ErrIf(t, err)
	assert.Equal(t, 2, len(users.Items))

	// Order by id desc
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 20, Page: 0, OrderBy: "id", Direction: DirectionDesc},
	})
	ErrIf(t, err)
	assert.Equal(t, int64(6), users.Items[0].Id)

	// Order by id asc
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 20, Page: 0, OrderBy: "id", Direction: DirectionAsc},
	})
	ErrIf(t, err)
	assert.Equal(t, int64(1), users.Items[0].Id)
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
	ErrIf(t, err)
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
	// With a given password
	cp.Email = "given-pword@mail.com"
	cp.Password = "M0nk3yNutz5"
	u, givenPassword, err := us.SignUp(cp)
	ErrIf(t, err)
	assert.Equal(t, "", givenPassword)
	_, err = us.SignIn(SignInParams{Email: cp.Email, Password: cp.Password})
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: cp.Password})
	ErrIf(t, err)

	// With a generated password
	cp.Email = "generated@mail.com"
	cp.Password = ""
	u, tempPassword, err := us.SignUp(cp)
	id := u.Id
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: tempPassword})
	assert.IsType(t, ErrNotAuth, err)
	newPass := "asdfj23£$sdfD"
	err = us.ChangePassword(ChangePasswordParams{Email: cp.Email, ResetToken: tempPassword, NewPassword: newPass})
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Email: cp.Email, Password: newPass})
	ErrIf(t, err)

	// Suspend
	err = us.Suspend(id)
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: newPass})
	assert.Error(t, err)
	sususer, err := us.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, true, sususer.Suspended)
	err = us.Restore(id)
	ErrIf(t, err)
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: newPass})
	ErrIf(t, err)
	sususer, err = us.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, false, sususer.Suspended)
	// TODO: create organisation within this test so we don't have to run whole suite.
	ErrIf(t, orgsv.Suspend(cp.OrgId))
	_, err = us.SignIn(SignInParams{Username: cp.Email, Password: newPass})
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
	// TODO: check the logic as lock time varies slightly and makes test indeterminate
	time.Sleep(time.Millisecond * time.Duration(2500))
	assert.False(t, us.isLocked(username))
}

func TestUsers_PasswordReset(t *testing.T) {
	email := "reset@mail.com"
	password := "M0nk3yNutz5"
	u, _, err := us.SignUp(SignUpParams{Email: email, Password: password})
	ErrIf(t, err)
	newP := "newPassword1!"
	err = us.ChangePassword(ChangePasswordParams{Email: email, ExistingPassword: password, NewPassword: newP})
	ErrIf(t, err)
	uc, err := us.SignIn(SignInParams{Username: u.Email, Password: newP})
	ErrIf(t, err)
	assert.Equal(t, email, uc.Email)

	// RESET TOKEN
	token, err := us.ResetPassword(ResetPasswordParams{Email: email})
	ErrIf(t, err)
	assert.NotEmpty(t, token)
	newP2 := "sdf@348DFsdf"
	err = us.ChangePassword(ChangePasswordParams{Email: email, ResetToken: token, NewPassword: newP2})
	ErrIf(t, err)
	uc, err = us.SignIn(SignInParams{Username: u.Email, Password: newP2})
	ErrIf(t, err)
	assert.Equal(t, email, uc.Email)

	// CAN'T USE SAME TOKEN TWICE
	err = us.ChangePassword(ChangePasswordParams{Email: email, ResetToken: token, NewPassword: "SSDFU23@£Dsdf"})
	assert.IsType(t, ErrNotFound, err)

	// RESET TOKEN EXPIRED
	token, err = us.ResetPassword(ResetPasswordParams{Email: email})
	ErrIf(t, err)
	time.Sleep(time.Millisecond * time.Duration(2000))
	newP3 := "sdfASDF34&8DFsdf"
	err = us.ChangePassword(ChangePasswordParams{Email: email, ResetToken: token, NewPassword: newP3})
	assert.Equal(t, ErrTokenExpired, err)

	// INVALID TOKEN
	token, err = us.ResetPassword(ResetPasswordParams{Email: email})
	ErrIf(t, err)
	newP4 := "sdf23@348DFsdf"
	err = us.ChangePassword(ChangePasswordParams{Email: email, ResetToken: token + "ADSF", NewPassword: newP4})
	assert.Equal(t, ErrInvalidResetToken, err)
	uc, err = us.SignIn(SignInParams{Username: u.Email, Password: newP2})
	ErrIf(t, err)
	assert.Equal(t, email, uc.Email)
}
