package gus

import (
	"database/sql"
	"time"
	"github.com/asaskevich/govalidator"
	"golang.org/x/crypto/bcrypt"
	"fmt"
)

const (
	ERR_STRING_EMAIL_CONSTRAINT string = "UNIQUE constraint failed: users.email"
)

var (
	ErrEmailTaken        error = ErrInvalid("That email is taken.")
	ErrEmailInvalid      error = ErrInvalid("'email' invalid.")
	ErrEmailRequired     error = ErrInvalid("'email' required.")
	ErrPasswordRequired  error = ErrInvalid("'password' required.")
	ErrInvalidResetToken error = ErrInvalid("Invalid reset token.")
	ErrPasswordInvalid   error = ErrInvalid(
		"'new_password' must contain: 1 Upper, 1 Lower, 1 Number and 8 Chars",
		"OR any alphanumeric with a minimum of 15 chars.")
	ResetTokenExpirySeconds int64
	ResetTokenExpiryKey     = "RESET_TOKEN_EXPIRY"
)

type Role int64

type UserOpts struct {
	MaxAuthAttempts     int           // Maximum amount of times a user can attempt to login with a given username.
	AttemptLockDuration time.Duration // Duration which the user will be locked out if MaxAuthAttempts has been exceeded.
	PassGen             PasswordGen   // A function used to generate passwords and reset tokens
	PassGenLength       int           // When a random password is generated when a user is created by another user
	// (as opposed to registered) this is the length of the generated password length.
}

type User struct {
	Id        int64 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	OrgId     int64 `json:"org_id"`
	Updated   time.Time `json:"updated"`
	Created   time.Time `json:"created"`
	Role      Role `json:"role"`
	Suspended bool `json:"suspended"`
}

type UserWithClaims struct {
	*User
	*Claims
}

type Claims struct {
	Role         Role `json:"role"`
	OrgId        int64 `json:"org_id"`
	OrgSuspended bool `json:"org_suspended"`
}

type UserWithToken struct {
	User
	Token string `json:"token"`
}

func NewUsers(db *sql.DB, opt UserOpts) *Users {
	if opt.AttemptLockDuration == 0 {
		opt.AttemptLockDuration = time.Duration(5) * time.Minute
	}
	if opt.PassGen == nil {
		opt.PassGen = RandStringBytesMaskImprSrc
	}
	if opt.PassGenLength == 0 {
		opt.PassGenLength = 15
	}
	return &Users{
		db:        db,
		Suspender: NewSuspender("users", db),
		UserOpts:  opt,
	}
}

type Users struct {
	db *sql.DB
	*Suspender
	UserOpts
}

func NewCreateUserParams() CreateUserParams {
	return CreateUserParams{}
}

type CreateUserParams struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	OrgId     int64 `json:"org_id"`
	Role      Role `json:"role"`
	CustomValidator `json:"-"`
}

func (va *CreateUserParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if !govalidator.IsEmail(va.Email) {
		return ErrEmailRequired
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Create returns a user, random password and [error]
func (us *Users) Create(p CreateUserParams) (*User, string, error) {
	stmt, err := us.db.Prepare("INSERT INTO users(email, first_name, last_name, phone, password_hash, org_id, updated, created, deleted, role, suspended) values(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return nil, "", err
	}
	u := &User{Email: p.Email, FirstName: p.FirstName, LastName: p.LastName, Phone: p.Phone,
		OrgId:    p.OrgId, Created: time.Now(), Updated: time.Now(), Role: p.Role, Suspended: false}

	password := us.UserOpts.PassGen(us.UserOpts.PassGenLength)
	hash, err := hashPassword(password)
	if err != nil {
		return nil, "", err
	}

	res, err := stmt.Exec(u.Email, u.FirstName, u.LastName, u.Phone, hash, u.OrgId, u.Updated, u.Created, 0, u.Role, u.Suspended)
	if err != nil {
		if err.Error() == ERR_STRING_EMAIL_CONSTRAINT {
			return nil, "", ErrEmailTaken
		}
		return nil, "", err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, "", err
	}
	u.Id = id
	return u, password, nil
}

func (us *Users) Get(id int64) (*User, error) {
	stmt, err := us.db.Prepare("SELECT id, email, first_name, last_name, phone, org_id, created, updated, role, suspended from users WHERE id =  ? AND deleted = 0 LIMIT 1")
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(id)
	var u User
	err = CheckNotFound(row.Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.Phone, &u.OrgId,
		&u.Created, &u.Updated, &u.Role, &u.Suspended))
	if err != nil {
		return nil, err
	}
	return &u, err
}

type SignInParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	CustomValidator `json:"-"`
}

func (va *SignInParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if govalidator.IsNull(va.Password) {
		return ErrPasswordRequired
	}
	if govalidator.IsNull(va.Email) {
		return ErrEmailRequired
	}
	if !govalidator.IsEmail(va.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// GetByEmail returns a user and password hash
func (us *Users) GetByEmail(email string) (*UserWithClaims, string, error) {
	stmt, err := us.db.Prepare("SELECT u.password_hash, u.id, u.email, u.first_name, u.last_name, u.phone, u.org_id, u.created, u.updated, u.role, u.suspended, COALESCE(o.suspended, 0) from users u left join orgs o on u.org_id = o.id WHERE u.email = ? AND u.deleted = 0 LIMIT 1")
	if err != nil {
		return nil, "", err
	}
	row := stmt.QueryRow(email)
	var u User
	var passwordHash string
	var orgSuspended bool
	err = CheckNotFound(row.Scan(&passwordHash, &u.Id, &u.Email, &u.FirstName, &u.LastName, &u.Phone,
		&u.OrgId, &u.Created, &u.Updated, &u.Role, &u.Suspended, &orgSuspended))
	if err != nil {
		return nil, "", err
	}
	c := &UserWithClaims{User: &u, Claims: &Claims{OrgId: u.OrgId, Role: u.Role, OrgSuspended: orgSuspended}}
	return c, passwordHash, err
}

func (us *Users) Authenticate(p SignInParams) (*UserWithClaims, error) {
	if us.isLocked(p.Email) {
		return nil, &RateLimitExceededError{Messages: []string{"Too many sign-in attempts try again later."}}
	}
	u, hash, err := us.GetByEmail(p.Email)
	if err != nil {
		_, ok := err.(*NotFoundError)
		if ok {
			return nil, ErrNotAuth
		}
		return nil, err
	}
	Debug(fmt.Sprintf("CLAIMS: %+v", *u.Claims))
	if u.Suspended || u.OrgSuspended {
		Debug("FAILED ATTEMPT:", us.isLocked(p.Email))
		return nil, ErrNotAuth
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(p.Password))
	if err != nil {
		return nil, ErrNotAuth
	}
	return u, nil
}

func (us *Users) isLocked(username string) bool {
	stmt, err := us.db.Prepare("INSERT into password_attempts (username, created) values (?, ?)")
	_, err = stmt.Exec(username, time.Now().Unix())
	if err != nil {
		LogErr(err)
		// Lock the account regardless
		return true
	}

	since := time.Now().Unix() - int64(us.AttemptLockDuration/time.Second)
	Debug("SINCE:", since)
	row := us.db.QueryRow("SELECT COUNT(username) FROM password_attempts WHERE created > ? AND username = ?", since, username)
	var count int
	err = row.Scan(&count)
	if err != nil {
		LogErr(err)
		// Lock the account regardless
		return true
	}
	Debug("ATTEMPTS:", count, "max:", us.MaxAuthAttempts)
	return count > us.MaxAuthAttempts
}

type UpdateUserParams struct {
	Id        *int64 `json:"id"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Role      *Role `json:"role"`
	CustomValidator `json:"-"`
}

func (va *UpdateUserParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if *va.Email != "" && !govalidator.IsEmail(*va.Email) {
		return ErrInvalid("'email' invalid.")
	}
	return nil
}

func (us *Users) Update(p UpdateUserParams) error {
	u, err := us.Get(*p.Id)
	if err != nil {
		return err
	}
	ApplyUpdates(u, p)
	stmt, err := us.db.Prepare("UPDATE users SET first_name = ?, last_name = ?, email = ?, phone = ?, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	err = CheckUpdated(stmt.Exec(u.FirstName, u.LastName, u.Email, u.Phone, time.Now(), u.Id))
	if err != nil && err.Error() == ERR_STRING_EMAIL_CONSTRAINT {
		return ErrEmailTaken
	}
	return err
}

func (us *Users) Delete(id int64) error {
	stmt, err := us.db.Prepare("UPDATE users SET deleted = 1, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	return CheckUpdated(stmt.Exec(time.Now(), id))
}

type ListUsersParams struct {
	OrgId int64 `json:"org_id"`
	ListArgs
	CustomValidator `json:"-"`
}

func (va *ListUsersParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	return nil
}

func (us *Users) List(p ListUsersParams) ([]*User, error) {
	q := "SELECT id, email, first_name, last_name, phone, org_id, created, updated, role from users WHERE deleted = 0"
	args := []interface{}{}
	if p.OrgId > 0 {
		q += " AND org_id = ?"
		args = append(args, p.OrgId)
	}
	rows, err := GetRows(us.db, q, p.ListArgs, args...)
	if err != nil {
		return nil, err
	}
	users := []*User{}
	for rows.Next() {
		u := &User{}
		rows.Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.Phone, &u.OrgId, &u.Created, &u.Updated, &u.Role)
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

type ResetPasswordParams struct {
	Email string `json:"email"`
	CustomValidator `json:"-"`
}

func (va *ResetPasswordParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if govalidator.IsNull(va.Email) {
		return ErrEmailRequired
	}
	if !govalidator.IsEmail(va.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (us *Users) ResetPassword(p ResetPasswordParams) (string, error) {
	u, _, err := us.GetByEmail(p.Email)
	if err != nil {
		return "", err
	}
	token := us.PassGen(256)
	stmt, err := us.db.Prepare("INSERT into password_resets (user_id, email, reset_token, created, deleted) values (?, ?, ?, ?, ?)")
	_, err = stmt.Exec(u.Id, u.Email, token, time.Now(), 0)
	if err != nil {
		return "", err
	}
	return token, nil
}

type ChangePasswordParams struct {
	Email            string `json:"email"`
	ExistingPassword string `json:"existing_password"`
	NewPassword      string `json:"new_password"`
	ResetToken       string `json:"reset_token"`
	CustomValidator `json:"-"`
}

func (va *ChangePasswordParams) Validate() error {
	if va.CustomValidator != nil {
		return va.CustomValidator()
	}
	if govalidator.IsNull(va.Email) {
		return ErrEmailRequired
	}
	if !govalidator.IsEmail(va.Email) {
		return ErrEmailInvalid
	}
	if govalidator.IsNull(va.ExistingPassword) && govalidator.IsNull(va.ResetToken) {
		return ErrInvalid("'existing_password' or 'reset_token' required.")
	}
	if govalidator.IsNull(va.NewPassword) {
		return ErrInvalid("'new_password' is required.")
	}
	if !ValidatePassword(va.NewPassword) {
		return ErrPasswordInvalid
	}
	return nil
}

func (us *Users) ChangePassword(p ChangePasswordParams) error {
	if p.ExistingPassword != "" {
		_, err := us.Authenticate(SignInParams{Email: p.Email, Password: p.ExistingPassword})
		if err != nil {
			return err
		}
	}
	if p.ResetToken != "" {
		stmt, err := us.db.Prepare(
			"SELECT reset_token FROM password_resets where email = ? and created > ? and deleted = 0 " +
				"ORDER BY created DESC LIMIT 1")
		exp := time.Now().Add(-time.Second * time.Duration(ResetTokenExpirySeconds))
		row := stmt.QueryRow(p.Email, exp)
		var resetToken string
		err = CheckNotFound(row.Scan(&resetToken))
		if err != nil {
			return err
		}
		if resetToken != p.ResetToken {
			return ErrInvalidResetToken
		}
	}
	hash, err := hashPassword(p.NewPassword)
	if err != nil {
		return err
	}
	stmt, err := us.db.Prepare("UPDATE users SET password_hash = ?, updated = ? WHERE email = ? AND deleted = 0")
	err = CheckNotFound(err)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(hash, time.Now(), p.Email)
	return nil
}
