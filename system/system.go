package system

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/config"
)

var sys = system{
	controllers: controllers.NewControllerSet(),
	doors:       doors.NewDoors(),
	cards:       cards.NewCards(),
	groups:      groups.NewGroups(),
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
	rules       cards.IRules
	audit       audit.Trail
	taskQ       TaskQ
	retention   time.Duration // time after which 'deleted' items are permanently removed
}

type object catalog.Object

func Init(cfg config.Config, conf string, permissions cards.IRules, trail audit.Trail) error {
	if err := sys.doors.Load(cfg.HTTPD.System.Doors); err != nil {
		return err
	}

	if err := sys.controllers.Load(cfg.HTTPD.System.Controllers); err != nil {
		return err
	}

	if err := sys.groups.Load(cfg.HTTPD.System.Groups); err != nil {
		return err
	}

	if err := sys.cards.Load(cfg.HTTPD.System.Cards); err != nil {
		return err
	}

	sys.conf = conf
	sys.rules = permissions
	sys.audit = trail
	sys.retention = cfg.HTTPD.Retention

	controllers.SetAuditTrail(trail)
	doors.SetAuditTrail(trail)
	cards.SetAuditTrail(trail)
	cards.SetAuditTrail(trail)
	groups.SetAuditTrail(trail)

	controllers.SetWindows(cfg.HTTPD.System.Windows.Ok,
		cfg.HTTPD.System.Windows.Uncertain,
		cfg.HTTPD.System.Windows.Systime,
		cfg.HTTPD.System.Windows.CacheExpiry)

	sys.controllers.Print()
	sys.doors.Print()
	sys.groups.Print()
	sys.cards.Print()

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

func Groups() interface{} {
	type group struct {
		OID   string
		Name  string
		Index uint32
	}

	list := []group{}
	for k, v := range sys.groups.Groups {
		if v.IsValid() && !v.IsDeleted() {
			list = append(list, group{
				OID:   fmt.Sprintf("%v", k),
				Name:  sys.groups.Groups[k].Name,
				Index: sys.groups.Groups[k].Index,
			})
		}
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].Index < list[j].Index })

	return list
}

func (s *system) refresh() {
	if s == nil {
		return
	}

	sys.taskQ.Add(Task{
		f: s.controllers.Refresh,
	})

	sys.taskQ.Add(Task{
		f: s.sweep,
	})

	sys.taskQ.Add(Task{
		f: func() {
			CompareACL(s.rules)
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
			UpdateACL(s.rules)
		},
	})
}

func (s *system) sweep() {
	cutoff := time.Now().Add(-s.retention)
	log.Printf("INFO  Sweeping all items invalidated before %v", cutoff.Format("2006-01-02 15:04:05"))

	s.controllers.Sweep(s.retention)
	s.doors.Sweep(s.retention)
}

func (s *system) log(op string, info interface{}, auth auth.OpAuth) {
	if s.audit != nil {
		uid := ""
		if auth != nil {
			uid = auth.UID()
		}

		s.audit.Write(audit.LogEntry{
			UID:       uid,
			Module:    "system",
			Operation: op,
			Info:      info,
		})
	}
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
