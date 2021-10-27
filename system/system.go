package system

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	doors:       doors.NewDoors(),
	cards:       cards.NewCards(),
	groups:      groups.NewGroups(),
	events:      *events.NewEvents(),
	logs:        logs.NewLogs(),
	taskQ:       NewTaskQ(),
	retention:   6 * time.Hour,
}

type system struct {
	sync.RWMutex
	conf        string
	controllers controllers.ControllerSet
	doors       doors.Doors
	cards       cards.Cards
	groups      groups.Groups
	events      events.Events
	logs        logs.Logs
	rules       grule.Rules
	taskQ       TaskQ
	retention   time.Duration // time after which 'deleted' items are permanently removed
	callback    callback
	trail       trail
}

type trail struct {
	trail audit.AuditTrail
}

func (t trail) Write(records ...audit.AuditRecord) {
	t.trail.Write(records...)
	sys.logs.Received(records...)
}

type callback struct {
}

func Init(cfg config.Config, conf string) error {
	if err := sys.doors.Load(cfg.HTTPD.System.Doors); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	}

	if err := sys.controllers.Load(cfg.HTTPD.System.Controllers); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	}

	if err := sys.groups.Load(cfg.HTTPD.System.Groups); err != nil {
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

	if err := sys.events.Load(cfg.HTTPD.System.Events); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
	}

	if err := sys.logs.Load(cfg.HTTPD.System.Logs); err != nil {
		if os.IsNotExist(err) {
			warn(err)
		} else {
			return err
		}
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

	controllers.SetWindows(cfg.HTTPD.System.Windows.Ok,
		cfg.HTTPD.System.Windows.Uncertain,
		cfg.HTTPD.System.Windows.Systime,
		cfg.HTTPD.System.Windows.CacheExpiry)

	//	sys.controllers.Print()
	//	sys.doors.Print()
	//	sys.groups.Print()
	//	sys.cards.Print()
	//	sys.events.Print()

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

func System() interface{} {
	sys.RLock()

	defer sys.RUnlock()

	objects := []interface{}{}
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

func Events(start, count int) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	return sys.events.AsObjects(start, count)
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
			s.controllers.Refresh(&s.callback)
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
}

func unpack(m map[string]interface{}) ([]catalog.Object, error) {
	f := func(err error) error {
		return types.BadRequest(fmt.Errorf("Invalid request (%v)", err), fmt.Errorf("Error unpacking 'post' request (%w)", err))
	}

	o := struct {
		Objects []catalog.Object `json:"objects"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, f(err)
	}

	log.Printf("DEBUG %v", fmt.Sprintf("UNPACK %s\n", string(blob)))

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, f(err)
	}

	return o.Objects, nil
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
