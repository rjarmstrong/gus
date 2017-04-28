package gus

import "regexp"

var (
	Rgx_ValidPasswordChars = regexp.MustCompile(`^[a-z0-9A-Z $&+:=?@#|^*%!-]+$`)
	Rgx_OneLower           = regexp.MustCompile(`[a-z]+`)
	Rgx_OneUpper           = regexp.MustCompile(`[A-Z]+`)
	Rgx_OneSpecial         = regexp.MustCompile(`[$&+,:;=?@#|'<>.^*()%!-]+`)
	Rgx_OneNumeric         = regexp.MustCompile(`\d+`)
	Rgx_PasswordLength     = regexp.MustCompile(`^.{10,30}$`)
)

type Validator interface {
	Validate() error
}

// CustomValidator should be embedded in any struct which implements Validator,
// this allows consumers to override the validation.
type CustomValidator func() error

func (f CustomValidator) Validate() error {
	return f()
}

func TestStr(input string, rgx ...*regexp.Regexp) bool {
	for _, v := range rgx {
		if match := v.Find([]byte(input)); len(match) == 0 {
			return false
		}
	}
	return true
}

func ValidatePassword(in string) bool {
	return TestStr(in, Rgx_ValidPasswordChars) && TestStr(in, Rgx_OneLower, Rgx_OneNumeric, Rgx_OneUpper, Rgx_OneSpecial, Rgx_PasswordLength)
}
