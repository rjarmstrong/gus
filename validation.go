package gus

import "regexp"

var (
	Rgx_ValidPasswordChars     = regexp.MustCompile(`^[a-z0-9A-Z]+$`)
	Rgx_OneLower               = regexp.MustCompile(`[a-z]+`)
	Rgx_OneUpper               = regexp.MustCompile(`[A-Z]+`)
	Rgx_OneNumeric             = regexp.MustCompile(`\d+`)
	Rgx_PasswordLength         = regexp.MustCompile(`^.{8,30}$`)
	Rgx_PasswordExtendedLength = regexp.MustCompile(`^.{15,30}$`)
)

type Validator interface {
	Validate() error
}

func TestStr(input string, rgx ... *regexp.Regexp) bool {
	for _, v := range rgx {
		if match := v.Find([]byte(input)); len(match) == 0 {
			return false
		}
	}
	return true
}

func ValidatePassword(in string) bool {
	return TestStr(in, Rgx_ValidPasswordChars)  && (TestStr(in, Rgx_OneLower, Rgx_OneNumeric, Rgx_OneUpper, Rgx_PasswordLength) ||
		TestStr(in, Rgx_PasswordExtendedLength))
}