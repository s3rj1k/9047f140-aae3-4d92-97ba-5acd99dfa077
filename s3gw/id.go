package s3gw

import (
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

var idRegExp *regexp.Regexp

func init() {
	idRegExp = regexp.MustCompile(`^[a-zA-Z0-9]{1,32}$`)
}

func IsValidID(id string) bool {
	return idRegExp.MatchString(id)
}

func GetID(r *http.Request, w http.ResponseWriter) string {
	vars := mux.Vars(r)
	if vars == nil {
		vars = make(map[string]string)
	}

	id := vars["id"]

	if !IsValidID(id) {
		return ""
	}

	return id
}
