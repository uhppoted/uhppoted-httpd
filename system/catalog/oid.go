package catalog

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type OID string
type Suffix string

func Join(oid OID, suffix Suffix) string {
	return regexp.MustCompile(`\.+`).ReplaceAllString(fmt.Sprintf("%v.%v", oid, suffix), ".")
}

func (oid OID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(oid))
}

func (oid *OID) UnmarshalJSON(bytes []byte) error {
	var s string

	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	*oid = OID(s)

	return nil
}

func (s Suffix) Append(v string) Suffix {
	oid := regexp.MustCompile(`\.+`).ReplaceAllString(fmt.Sprintf(".%v.%v", s, v), ".")

	return Suffix(oid)
}
