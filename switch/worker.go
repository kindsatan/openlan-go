package _switch

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/point"
	"github.com/danieldin95/openlan-go/switch/api"
	"github.com/danieldin95/openlan-go/switch/storage"
	"sync"
	"time"
)

type NetworkWorker struct {
	// private
	alias       string
	cfg         config.Network
	newTime     int64
	startTime   int64
	linksLock   sync.RWMutex
	links       map[string]*point.Point
	uuid        string
	initialized bool
	crypt       *config.Crypt
}

func NewNetworkWorker(c config.Network, crypt *config.Crypt) *NetworkWorker {
	w := NetworkWorker{
		alias:       c.Alias,
		cfg:         c,
		newTime:     time.Now().Unix(),
		startTime:   0,
		links:       make(map[string]*point.Point),
		initialized: false,
		crypt:       crypt,
	}

	return &w
}

func (w *NetworkWorker) Initialize() {
	w.initialized = true

	for _, pass := range w.cfg.Password {
		user := models.User{
			Name:     pass.Username + "@" + w.cfg.Name,
			Password: pass.Password,
		}
		storage.User.Add(&user)
	}
	if w.cfg.Subnet.Netmask != "" {
		met := models.Network{
			Name:    w.cfg.Name,
			IpStart: w.cfg.Subnet.Start,
			IpEnd:   w.cfg.Subnet.End,
			Netmask: w.cfg.Subnet.Netmask,
			Routes:  make([]*models.Route, 0, 2),
		}
		for _, rt := range w.cfg.Routes {
			if rt.NextHop == "" {
				libol.Warn("NetworkWorker.Initialize %s no nexthop", rt.Prefix)
				continue
			}
			met.Routes = append(met.Routes, &models.Route{
				Prefix:  rt.Prefix,
				NextHop: rt.NextHop,
			})
		}
		storage.Network.Add(&met)
	}
}

func (w *NetworkWorker) ID() string {
	return w.uuid
}

func (w *NetworkWorker) String() string {
	return w.ID()
}

func (w *NetworkWorker) LoadLinks() {
	if w.cfg.Links != nil {
		for _, lin := range w.cfg.Links {
			lin.Default()
			w.AddLink(lin)
		}
	}
}

func (w *NetworkWorker) Start(v api.Switcher) {
	libol.Info("NetworkWorker.Start: %s", w.cfg.Name)
	if !w.initialized {
		w.Initialize()
	}
	w.uuid = v.UUID()
	w.startTime = time.Now().Unix()
	w.LoadLinks()
}

func (w *NetworkWorker) Stop() {
	libol.Info("NetworkWorker.Close: %s", w.cfg.Name)
	for _, p := range w.links {
		p.Stop()
	}
	w.startTime = 0
}

func (w *NetworkWorker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *NetworkWorker) AddLink(c *config.Point) {
	c.Alias = w.alias
	c.Interface.Bridge = w.cfg.Bridge.Name //Reset bridge name.
	c.RequestAddr = false
	c.Network = w.cfg.Name
	c.Interface.Address = w.cfg.Bridge.Address
	libol.Go(func() {
		p := point.NewPoint(c)
		p.Initialize()
		w.linksLock.Lock()
		w.links[c.Connection] = p
		w.linksLock.Unlock()
		storage.Link.Add(p)
		p.Start()
	})
}

func (w *NetworkWorker) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()

	if p, ok := w.links[addr]; ok {
		p.Stop()
		storage.Link.Del(p.Addr())
		delete(w.links, addr)
	}
}
