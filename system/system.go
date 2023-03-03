package system

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/events"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
	"github.com/uhppoted/uhppoted-httpd/system/grule"
	"github.com/uhppoted/uhppoted-httpd/system/history"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/system/logs"
	"github.com/uhppoted/uhppoted-httpd/system/users"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Tag string

const (
	TagInterfaces  Tag = "interfaces"
	TagControllers Tag = "controllers"
	TagDoors       Tag = "doors"
	TagCards       Tag = "cards"
	TagGroups      Tag = "groups"
	TagEvents      Tag = "events"
	TagLogs        Tag = "logs"
	TagUsers       Tag = "users"
	TagHistory     Tag = "history"
)

var channels = struct {
	events chan types.EventsList
}{
	events: make(chan types.EventsList),
}

var sys = system{
	interfaces:  interfaces.NewInterfaces(channels.events),
	controllers: controllers.NewControllers(),
	doors:       doors.NewDoors(),
	cards:       cards.NewCards(),
	groups:      groups.NewGroups(),
	events:      events.NewEvents(),
	logs:        logs.NewLogs(),
	users:       users.NewUsers(),
	history:     history.NewHistory(),

	mode:      types.Normal,
	withPIN:   false,
	taskQ:     NewTaskQ(),
	retention: 6 * time.Hour,
}

type system struct {
	sync.RWMutex
	conf string

	interfaces  interfaces.Interfaces
	controllers controllers.Controllers
	doors       doors.Doors
	cards       cards.Cards
	groups      groups.Groups
	events      events.Events
	logs        logs.Logs
	users       users.Users
	history     history.History

	files     map[Tag]string
	rules     grule.Rules
	taskQ     TaskQ
	retention time.Duration // time after which 'deleted' items are permanently removed
	trail     trail
	mode      types.RunMode
	withPIN   bool
	debug     bool
}

type trail struct {
	trail audit.AuditTrail
}

func (t trail) Write(records ...audit.AuditRecord) {
	t.trail.Write(records...)
	sys.logs.Received(records...)
	sys.history.Received(records...)

	if err := save(TagLogs, &sys.logs); err != nil {
		warnf("%v", err)
	}

	if err := save(TagHistory, &sys.history); err != nil {
		warnf("%v", err)
	}
}

type object struct {
	OID   schema.OID `json:"OID"`
	Value string     `json:"value"`
}

type serializable interface {
	Load(blob json.RawMessage) error
	Save() (json.RawMessage, error)
	Print()
}

func Init(cfg config.Config, conf string, mode types.RunMode, debug bool) error {
	catalog.Init(memdb.NewCatalog())

	sys.mode = mode
	sys.withPIN = cfg.HTTPD.WithPIN

	sys.files = map[Tag]string{
		TagInterfaces:  cfg.HTTPD.System.Interfaces,
		TagControllers: cfg.HTTPD.System.Controllers,
		TagDoors:       cfg.HTTPD.System.Doors,
		TagCards:       cfg.HTTPD.System.Cards,
		TagGroups:      cfg.HTTPD.System.Groups,
		TagEvents:      cfg.HTTPD.System.Events,
		TagLogs:        cfg.HTTPD.System.Logs,
		TagUsers:       cfg.HTTPD.System.Users,
		TagHistory:     cfg.HTTPD.System.History,
	}

	list := subsystems()
	for _, v := range list {
		if err := load(v.tag, v.serializable); err != nil {
			log.Errorf("Unable to load %v from %v (%v)", v.tag, sys.files[v.tag], err)
			return err
		}
	}

	kb := ast.NewKnowledgeLibrary()
	if err := builder.NewRuleBuilder(kb).BuildRuleFromResource("acl", "0.0.0", pkg.NewFileResource(cfg.HTTPD.DB.Rules.ACL)); err != nil {
		log.Fatalf("Error loading ACL ruleset (%v)", err)
	}

	rules, err := grule.NewGrule(kb)
	if err != nil {
		log.Fatalf("Error initialising ACL ruleset (%v)", err)
	}

	sys.debug = debug
	sys.conf = conf
	sys.rules = rules
	sys.retention = cfg.HTTPD.Retention
	sys.trail = trail{
		trail: audit.MakeTrail(),
	}

	controllers.SetWindows(cfg.HTTPD.System.Windows.Ok,
		cfg.HTTPD.System.Windows.Uncertain,
		cfg.HTTPD.System.Windows.Systime,
		cfg.HTTPD.System.Windows.CacheExpiry)

	// for _, v := range list {
	// 	v.Print()
	// }

	go func() {
		time.Sleep(2500 * time.Millisecond)
		sys.refresh()

		c := time.Tick(cfg.HTTPD.System.Refresh)
		for _ = range c {
			sys.refresh()
		}
	}()

	go func(ch <-chan types.EventsList) {
		for v := range ch {
			AppendEvents(v)
		}
	}(channels.events)

	return nil
}

func (s *system) refresh() {
	if s == nil {
		return
	}

	f := func(controllers []types.IController) []uint32 {
		list := []uint32{}
		for _, c := range controllers {
			list = append(list, c.ID())
		}

		return list
	}

	controllers := s.controllers.AsIControllers()
	missing := s.events.Missing(2, f(controllers)...) // Fix at most 2 gaps in each controller's event list

	sys.taskQ.Add(Task{
		f: func() {
			found := s.interfaces.Search(controllers)
			s.controllers.Found(found)
		},
	})

	sys.taskQ.Add(Task{
		f: func() {
			s.interfaces.Refresh(controllers)
		},
	})

	sys.taskQ.Add(Task{
		f: func() {
			s.interfaces.GetEvents(controllers, missing)
		},
	})

	sys.taskQ.Add(Task{
		f: s.sweep,
	})

	sys.taskQ.Add(Task{
		f: func() {
			s.compareACL()
		},
	})

	if sys.mode == types.Synchronize {
		sys.taskQ.Add(Task{
			f: func() {
				s.synchronize()
			},
		})
	}
}

func (s *system) synchronize() {
	infof("Checking system synchronization")
	controllers := sys.controllers.AsIControllers()

	unsynchronized := struct {
		datetime bool
		doors    bool
		ACL      bool
	}{}

	for _, c := range controllers {
		if !c.DateTimeOk() {
			warnf("Controller %v date/time out of synch", c.ID())
			unsynchronized.datetime = true
		}

		for _, d := range []uint8{1, 2, 3, 4} {
			if oid, ok := c.Door(d); ok {
				if door, ok := sys.doors.Door(oid); ok {
					if !door.IsOk() {
						warnf("Door '%v' out of synch", door)
						unsynchronized.doors = true
					}
				}
			}
		}
	}

	if acl, err := s.permissions(controllers); err != nil {
		warnf("%v", err)
	} else if diff, err := s.interfaces.CompareACL(controllers, acl, s.withPIN); err != nil {
		warnf("%v", err)
	} else if diff == nil {
		warnf("Invalid ACL diff (%v)", diff)
	} else {
		count := 0
		for _, v := range diff {
			count += len(v.Updated)
			count += len(v.Added)
			count += len(v.Deleted)
		}

		if count > 0 {
			warnf("ACL out of synch")
			unsynchronized.ACL = true
		}
	}

	if unsynchronized.datetime {
		warnf("Resynchronizing all controller date/times")
		SynchronizeDateTime()
	}

	if unsynchronized.doors {
		warnf("Resynchronizing mode and delay for all doors")
		SynchronizeDoors()
	}

	if unsynchronized.ACL {
		warnf("Resynchronizing ACL")
		SynchronizeACL()
	}
}

func SynchronizeACL() error {
	if err := sys.synchronizeACL(); err != nil {
		return err
	}

	sys.compareACL()

	return nil
}

func SynchronizeDateTime() error {
	controllers := sys.controllers.AsIControllers()
	now := time.Now()

	for _, c := range controllers {
		controller := c
		go func() {
			sys.interfaces.SetTime(controller, now)
		}()
	}

	return nil
}

func SynchronizeDoors() error {
	controllers := sys.controllers.AsIControllers()

	for _, c := range controllers {
		controller := c

		for _, d := range []uint8{1, 2, 3, 4} {
			if oid, ok := controller.Door(d); ok {
				if door, ok := sys.doors.Door(oid); ok {
					doorID := d

					go func(id uint8, door doors.Door) {
						sys.interfaces.SetDoor(controller, id, door.Mode(), door.Delay())
					}(doorID, door)
				}
			}
		}
	}

	return nil
}

func (s *system) Update(oid schema.OID, field schema.Suffix, value any) {
	controllers := s.controllers.AsIControllers()

	switch {
	case oid.HasPrefix(schema.CardsOID):
		if card, ok := value.(uint32); ok && card != 0 {
			for _, c := range controllers {
				controller := c
				go func() {
					s.updateCardPermissions(controller, card)
				}()
			}
		}
		return

	case oid.HasPrefix(schema.GroupsOID):
		list := map[schema.OID]cards.Card{}
		for _, c := range s.cards.List() {
			card := c
			for _, g := range c.Groups() {
				if g == oid {
					list[c.OID] = card
				}
			}
		}

		for _, card := range list {
			cardID := card.CardID
			for _, c := range controllers {
				controller := c
				go func() {
					s.updateCardPermissions(controller, cardID)
				}()
			}
		}

		return

	case oid.HasPrefix(schema.ControllersOID) && field == schema.ControllerDateTime:
		for _, c := range controllers {
			if c.OID() == oid {
				controller := c
				go func() {
					s.interfaces.SetTime(controller, value.(time.Time))
				}()
				return
			}
		}

	case oid.HasPrefix(schema.DoorsOID) && field == schema.DoorControl:
		for _, c := range controllers {
			for _, i := range []uint8{1, 2, 3, 4} {
				if d, ok := c.Door(i); ok && d == oid {
					ddoor, _ := sys.doors.Door(oid)
					controller := c
					door := i
					go func() {
						fmt.Printf(">>>>>>>> SetDoorControl - value:%v  configured:%v\n", value.(core.ControlState), ddoor.Mode())
						s.interfaces.SetDoorControl(controller, door, value.(core.ControlState))
					}()
					return
				}
			}
		}

	case oid.HasPrefix(schema.DoorsOID) && field == schema.DoorDelay:
		for _, c := range controllers {
			controller := c
			for _, i := range []uint8{1, 2, 3, 4} {
				door := i
				if d, ok := c.Door(i); ok && d == oid {
					ddoor, _ := sys.doors.Door(oid)

					go func() {
						fmt.Printf(">>>>>>>> SetDoorDelay   - value:%v  configured:%v\n", value.(uint8), ddoor.Delay())
						s.interfaces.SetDoorDelay(controller, door, value.(uint8))
					}()
					return
				}
			}
		}
	}
}

func (s *system) sweep() {
	cutoff := time.Now().Add(-s.retention)

	infof("Sweeping all items invalidated before %v", cutoff.Format("2006-01-02 15:04:05"))

	s.controllers.Sweep(s.retention)
	s.doors.Sweep(s.retention)
	s.cards.Sweep(s.retention)
	s.groups.Sweep(s.retention)
	s.users.Sweep(s.retention)
}

func subsystems() []struct {
	serializable
	tag Tag
} {
	return []struct {
		serializable
		tag Tag
	}{
		{&sys.interfaces, TagInterfaces},
		{&sys.controllers, TagControllers},
		{&sys.doors, TagDoors},
		{&sys.cards, TagCards},
		{&sys.groups, TagGroups},
		{&sys.events, TagEvents},
		{&sys.logs, TagLogs},
		{&sys.users, TagUsers},
		{&sys.history, TagHistory},
	}
}

func unpack(m map[string]interface{}) ([]object, []object, []schema.OID, error) {
	o := struct {
		Created []object     `json:"created"`
		Updated []object     `json:"updated"`
		Deleted []schema.OID `json:"deleted"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		warnf("%v", err)
		return nil, nil, nil, fmt.Errorf("Invalid request (%v)", err)
	}

	if sys.debug {
		log.Debugf("UNPACK %s\n", string(blob))
	}

	if err := json.Unmarshal(blob, &o); err != nil {
		warnf("%v", err)
		return nil, nil, nil, fmt.Errorf("Invalid request (%v)", err)
	}

	return o.Created, o.Updated, o.Deleted, nil
}

func load(tag Tag, v serializable) error {
	if file, ok := sys.files[tag]; !ok || file == "" {
		return nil
	} else {
		bytes, err := os.ReadFile(file)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}

			warnf("%v", err)
			return nil
		}

		blob := map[Tag]json.RawMessage{}
		if err = json.Unmarshal(bytes, &blob); err != nil {
			return err
		}

		return v.Load(blob[tag])
	}
}

func save(tag Tag, v serializable) error {
	var file string
	var ok bool

	if file, ok = sys.files[tag]; !ok || file == "" {
		return nil
	}

	bytes, err := v.Save()
	if err != nil {
		return err
	}

	blob := map[Tag]json.RawMessage{
		tag: bytes,
	}

	b, err := json.MarshalIndent(blob, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", fmt.Sprintf("uhppoted-%v.*", tag))
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), file)
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func infof(format string, args ...any) {
	log.Infof(format, args...)
}

func warnf(format string, args ...any) {
	log.Warnf(format, args...)
}
