package interfaces

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LAN struct {
	OID              catalog.OID
	Name             string
	BindAddress      core.BindAddr
	BroadcastAddress core.BroadcastAddr
	ListenAddress    core.ListenAddr
	Debug            bool

	created      types.DateTime
	deleted      *types.DateTime
	unconfigured bool
}

var created = time.Now()

func (l *LAN) IsValid() bool {
	if l != nil {
		if strings.TrimSpace(l.Name) != "" {
			return true
		}
	}

	return false
}

func (l *LAN) IsDeleted() bool {
	if l != nil && l.deleted != nil {
		return true
	}

	return false
}

func (l LAN) String() string {
	return fmt.Sprintf("%v", l.Name)
}

func (l LAN) serialize() ([]byte, error) {
	record := struct {
		OID              catalog.OID        `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.DateTime     `json:"created,omitempty"`
	}{
		OID:              l.OID,
		Name:             l.Name,
		BindAddress:      l.BindAddress,
		BroadcastAddress: l.BroadcastAddress,
		ListenAddress:    l.ListenAddress,
		Created:          types.DateTime(l.created),
	}

	return json.MarshalIndent(record, "", "  ")
}

func (l *LAN) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)
	datetime := types.DateTime(created)

	record := struct {
		OID              catalog.OID        `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          *types.DateTime    `json:"created,omitempty"`
	}{
		Created: &datetime,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	l.OID = record.OID
	l.Name = record.Name
	l.BindAddress = record.BindAddress
	l.BroadcastAddress = record.BroadcastAddress
	l.ListenAddress = record.ListenAddress
	l.created = *record.Created
	l.unconfigured = false

	return nil
}
