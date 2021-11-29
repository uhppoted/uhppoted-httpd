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

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/events"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
	"github.com/uhppoted/uhppoted-httpd/system/grule"
	"github.com/uhppoted/uhppoted-httpd/system/logs"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/config"
)

var sys = system{
	controllers: controllers.NewControllerSet(),
	doors: struct {
		doors doors.Doors
		file  string
		tag   string
	}{
		doors: doors.NewDoors(),
		tag:   "doors",
	},
	cards: cards.NewCards(),
	groups: struct {
		groups groups.Groups
		file   string
		tag    string
	}{
		groups: groups.NewGroups(),
		tag:    "groups",
	},
	events: struct {
		events events.Events
		file   string
		tag    string
	}{
		events: events.NewEvents(),
		tag:    "events",
	},

	logs: struct {
		logs logs.Logs
		file string
		tag  string
	}{
		logs: logs.NewLogs(),
		tag:  "logs",
	},

	taskQ:     NewTaskQ(),
	retention: 6 * time.Hour,
}

type system struct {
	sync.RWMutex
	conf        string
	controllers controllers.ControllerSet
	doors       struct {
		doors doors.Doors
		file  string
		tag   string
	}
	cards  cards.Cards
	groups struct {
		groups groups.Groups
		file   string
		tag    string
	}
	events struct {
		events events.Events
		file   string
		tag    string
	}
	logs struct {
		logs logs.Logs
		file string
		tag  string
	}
	rules     grule.Rules
	taskQ     TaskQ
	retention time.Duration // time after which 'deleted' items are permanently removed
	callback  callback
	trail     trail
	debug     bool
}

type trail struct {
	trail audit.AuditTrail
}

func (t trail) Write(records ...audit.AuditRecord) {
	t.trail.Write(records...)
	sys.logs.logs.Received(records...)

	if bytes, err := sys.logs.logs.Save(); err != nil {
		warn(err)
	} else if err := save(sys.logs.file, sys.logs.tag, bytes); err != nil {
		warn(err)
	}
}

type callback struct {
}

type object struct {
	OID   catalog.OID `json:"OID"`
	Value string      `json:"value"`
}

func Init(cfg config.Config, conf string, debug bool) error {
	sys.doors.file = cfg.HTTPD.System.Doors
	sys.groups.file = cfg.HTTPD.System.Groups
	sys.events.file = cfg.HTTPD.System.Events
	sys.logs.file = cfg.HTTPD.System.Logs

	if blob, err := load(sys.doors.file); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	} else if err := sys.doors.doors.Load(blob[sys.doors.tag]); err != nil {
		return err
	}

	if err := sys.controllers.Load(cfg.HTTPD.System.Controllers); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	}

	if err := sys.cards.Load(cfg.HTTPD.System.Cards); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	}

	if blob, err := load(sys.groups.file); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	} else if err := sys.groups.groups.Load(blob[sys.groups.tag]); err != nil {
		return err
	}

	if blob, err := load(sys.events.file); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	} else if err := sys.events.events.Load(blob[sys.events.tag]); err != nil {
		return err
	}

	if blob, err := load(sys.logs.file); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	} else if err := sys.logs.logs.Load(blob[sys.logs.tag]); err != nil {
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

	return nil
}

func Schema() interface{} {
	return catalog.GetSchema()
}

func System() interface{} {
	sys.RLock()

	defer sys.RUnlock()

	objects := []interface{}{}
	objects = append(objects, sys.controllers.AsObjects()...)
	objects = append(objects, sys.doors.doors.AsObjects()...)
	objects = append(objects, sys.cards.AsObjects()...)
	objects = append(objects, sys.groups.groups.AsObjects()...)

	return struct {
		Objects []interface{} `json:"objects"`
	}{
		Objects: objects,
	}
}

func Logs(start, count int) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	return sys.logs.logs.AsObjects(start, count)
}

func (s *system) refresh() {
	if s == nil {
		return
	}

	sys.taskQ.Add(Task{
		f: func() {
			if objects := s.controllers.Refresh(&s.callback); objects != nil {
				catalog.PutL(objects)
			}
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
			if objects := sys.controllers.Sync(); objects != nil {
				catalog.PutL(objects)
			}

			UpdateACL()
		},
	})
}

func (s *system) sweep() {
	cutoff := time.Now().Add(-s.retention)
	log.Printf("INFO  Sweeping all items invalidated before %v", cutoff.Format("2006-01-02 15:04:05"))

	s.controllers.Sweep(s.retention)
	s.doors.doors.Sweep(s.retention)
	s.cards.Sweep(s.retention)
	s.groups.groups.Sweep(s.retention)
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

func load(file string) (map[string]json.RawMessage, error) {
	blob := map[string]json.RawMessage{}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &blob); err != nil {
		return nil, err
	}

	return blob, nil
}

func save(file string, tag string, bytes json.RawMessage) error {
	if file == "" {
		return nil
	}

	serializable := map[string]json.RawMessage{
		tag: bytes,
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
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
