package groups

import (
	"encoding/json"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type Group struct {
	OID  catalog.OID `json:"OID"`
	Name string      `json:"name"`

	created time.Time
	deleted *time.Time
}

const GroupName = catalog.GroupName
const GroupCreated = catalog.GroupCreated

var created = time.Now()

func (g Group) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID `json:"OID"`
		Name    string      `json:"name,omitempty"`
		Created string      `json:"created"`
	}{
		OID:     g.OID,
		Name:    g.Name,
		Created: g.created.Format("2006-01-02 15:04:05"),
	}

	return json.Marshal(record)
}

func (g *Group) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     string `json:"OID"`
		Name    string `json:"name,omitempty"`
		Created string `json:"created"`
	}{}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	g.OID = catalog.OID(record.OID)
	g.Name = record.Name
	g.created = created

	if t, err := time.Parse("2006-01-02 15:04:05", record.Created); err == nil {
		g.created = t
	}

	return nil
}
