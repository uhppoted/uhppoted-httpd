package system

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/events"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
	"github.com/uhppoted/uhppoted-httpd/system/grule"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/system/logs"
	"github.com/uhppoted/uhppoted-httpd/types"
)

var channels = struct {
	events chan types.EventsList
}{
	events: make(chan types.EventsList),
}

var sys = system{
	interfaces: struct {
		interfaces.Interfaces
		file string
		tag  string
	}{
		Interfaces: interfaces.NewInterfaces(channels.events),
		tag:        "interfaces",
	},

	controllers: struct {
		controllers.Controllers
		file string
		tag  string
	}{
		Controllers: controllers.NewControllers(),
		tag:         "controllers",
	},

	doors: struct {
		doors.Doors
		file string
		tag  string
	}{
		Doors: doors.NewDoors(),
		tag:   "doors",
	},

	cards: struct {
		cards.Cards
		file string
		tag  string
	}{
		Cards: cards.NewCards(),
		tag:   "cards",
	},

	groups: struct {
		groups.Groups
		file string
		tag  string
	}{
		Groups: groups.NewGroups(),
		tag:    "groups",
	},

	events: struct {
		events.Events
		file string
		tag  string
	}{
		Events: events.NewEvents(),
		tag:    "events",
	},

	logs: struct {
		logs.Logs
		file string
		tag  string
	}{
		Logs: logs.NewLogs(),
		tag:  "logs",
	},

	taskQ:     NewTaskQ(),
	retention: 6 * time.Hour,
}

type system struct {
	sync.RWMutex
	conf string

	interfaces struct {
		interfaces.Interfaces
		file string
		tag  string
	}

	controllers struct {
		controllers.Controllers
		file string
		tag  string
	}

	doors struct {
		doors.Doors
		file string
		tag  string
	}

	cards struct {
		cards.Cards
		file string
		tag  string
	}

	groups struct {
		groups.Groups
		file string
		tag  string
	}

	events struct {
		events.Events
		file string
		tag  string
	}

	logs struct {
		logs.Logs
		file string
		tag  string
	}

	rules     grule.Rules
	taskQ     TaskQ
	retention time.Duration // time after which 'deleted' items are permanently removed
	trail     trail
	debug     bool
}

type trail struct {
	trail audit.AuditTrail
}

func (t trail) Write(records ...audit.AuditRecord) {
	t.trail.Write(records...)
	sys.logs.Received(records...)

	if err := save(sys.logs.file, sys.logs.tag, &sys.logs); err != nil {
		warn(err)
	}
}

type object struct {
	OID   catalog.OID `json:"OID"`
	Value string      `json:"value"`
}

type serializable interface {
	Load(blob json.RawMessage) error
	Save() (json.RawMessage, error)
	Print()
}

func Init(cfg config.Config, conf string, debug bool) error {
	sys.interfaces.file = cfg.HTTPD.System.Interfaces
	sys.controllers.file = cfg.HTTPD.System.Controllers
	sys.doors.file = cfg.HTTPD.System.Doors
	sys.cards.file = cfg.HTTPD.System.Cards
	sys.groups.file = cfg.HTTPD.System.Groups
	sys.events.file = cfg.HTTPD.System.Events
	sys.logs.file = cfg.HTTPD.System.Logs

	if err := load(sys.interfaces.file, sys.interfaces.tag, &sys.interfaces); err != nil {
		return err
	}

	if err := load(sys.controllers.file, sys.controllers.tag, &sys.controllers); err != nil {
		return err
	}

	if err := load(sys.doors.file, sys.doors.tag, &sys.doors); err != nil {
		return err
	}

	if err := load(sys.cards.file, sys.cards.tag, &sys.cards); err != nil {
		return err
	}

	if err := load(sys.groups.file, sys.groups.tag, &sys.groups); err != nil {
		return err
	}

	if err := load(sys.events.file, sys.events.tag, &sys.events); err != nil {
		return err
	}

	if err := load(sys.logs.file, sys.logs.tag, &sys.logs); err != nil {
		return err
	}

	kb := ast.NewKnowledgeLibrary()
	if err := builder.NewRuleBuilder(kb).BuildRuleFromResource("acl", "0.0.0", pkg.NewFileResource(cfg.HTTPD.DB.Rules.ACL)); err != nil {
		log.Fatal(fmt.Errorf("Error loading ACL ruleset (%v)", err))
	}

	rules, err := grule.NewGrule(kb)
	if err != nil {
		log.Fatal(fmt.Errorf("Error initialising ACL ruleset (%v)", err))
	}

	sys.conf = conf
	sys.rules = rules
	sys.retention = cfg.HTTPD.Retention
	sys.trail = trail{
		trail: audit.MakeTrail(),
	}
	sys.debug = debug

	sys.controllers.Init(sys.interfaces.Interfaces)

	controllers.SetWindows(cfg.HTTPD.System.Windows.Ok,
		cfg.HTTPD.System.Windows.Uncertain,
		cfg.HTTPD.System.Windows.Systime,
		cfg.HTTPD.System.Windows.CacheExpiry)

	//	sys.controllers.Print()
	//	sys.doors.Print()
	//	sys.groups.Print()
	//	sys.cards.Print()
	//	sys.events.Print()
	//	sys.logs.Print()

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

func Schema() interface{} {
	return catalog.GetSchema()
}

func System() interface{} {
	sys.RLock()

	defer sys.RUnlock()

	objects := []interface{}{}
	objects = append(objects, sys.interfaces.AsObjects()...)
	objects = append(objects, sys.controllers.AsObjects()...)
	objects = append(objects, sys.doors.AsObjects()...)
	objects = append(objects, sys.cards.AsObjects()...)
	objects = append(objects, sys.groups.AsObjects()...)

	return struct {
		Objects []interface{} `json:"objects"`
	}{
		Objects: objects,
	}
}

func Logs(start, count int) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	return sys.logs.AsObjects(start, count)
}

func (s *system) refresh() {
	if s == nil {
		return
	}

	sys.taskQ.Add(Task{
		f: func() {
			s.controllers.Refresh()
		},
	})

	sys.taskQ.Add(Task{
		f: s.sweep,
	})

	sys.taskQ.Add(Task{
		f: func() {
			CompareACL()
		},
	})
}

func (s *system) updated() {
	//	s.taskQ.Add(Task{
	//		f: func() {
	//			if err := controllers.Export(sys.conf, shadow.Controllers, sys.doors.Doors); err != nil {
	//				warn(err)
	//			}
	//		},
	//	})

	s.taskQ.Add(Task{
		f: func() {
			info("Updating controllers from configuration")
			sys.controllers.Sync()

			UpdateACL()
		},
	})
}

func (s *system) sweep() {
	cutoff := time.Now().Add(-s.retention)
	log.Printf("INFO  Sweeping all items invalidated before %v", cutoff.Format("2006-01-02 15:04:05"))

	s.controllers.Sweep(s.retention)
	s.doors.Sweep(s.retention)
	s.cards.Sweep(s.retention)
	s.groups.Sweep(s.retention)
}

func unpack(m map[string]interface{}) ([]object, error) {
	f := func(err error) error {
		return types.BadRequest(fmt.Errorf("Invalid request (%v)", err), fmt.Errorf("Error unpacking 'post' request (%w)", err))
	}

	o := struct {
		Objects []object `json:"objects"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, f(err)
	}

	if sys.debug {
		log.Printf("DEBUG %v", fmt.Sprintf("UNPACK %s\n", string(blob)))
	}

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, f(err)
	}

	return o.Objects, nil
}

func load(file string, tag string, v serializable) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		warn(err)
		return nil
	}

	blob := map[string]json.RawMessage{}
	if err = json.Unmarshal(bytes, &blob); err != nil {
		return err
	}

	return v.Load(blob[tag])
}

func save(file string, tag string, v serializable) error {
	if file == "" {
		return nil
	}

	bytes, err := v.Save()
	if err != nil {
		return err
	}

	blob := map[string]json.RawMessage{
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

// Returns a deduplicated list of objects, retaining only the the last (i.e. latest) value.
// NOTE: this implementation is horribly inefficient but the list is expected to almost always
//       be tiny since it is the result of a manual edit.
func squoosh(objects []catalog.Object) []catalog.Object {
	keys := map[catalog.OID]struct{}{}
	list := []catalog.Object{}

	for i := len(objects); i > 0; i-- {
		object := objects[i-1]
		oid := object.OID
		if _, ok := keys[oid]; !ok {
			keys[oid] = struct{}{}
			list = append([]catalog.Object{object}, list...)
		}
	}

	return list
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func info(msg string) {
	log.Printf("INFO  %v", msg)
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}

// TODO remove - debugging only
func beep() {
	exec.Command("say", "beep").Run()
}

// TODO remove - debugging only
func beep2() {
	exec.Command("say", "beep beep").Run()
}
