package s3gw

import (
	"unicode"
	"unicode/utf8"
)

func CapitalizeErrorString(err error) string {
	if err == nil {
		return ""
	}

	errorMsg := err.Error()
	if errorMsg == "" {
		return ""
	}

	r, size := utf8.DecodeRuneInString(errorMsg)
	if r == utf8.RuneError {
		return errorMsg // return original string if invalid UTF-8
	}

	return string(unicode.ToUpper(r)) + errorMsg[size:]
}
