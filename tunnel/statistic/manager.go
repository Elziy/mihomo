package statistic

import (
	"os"
	"time"

	"github.com/metacubex/mihomo/common/atomic"
	C "github.com/metacubex/mihomo/constant"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/shirou/gopsutil/v3/process"
)

var DefaultManager *Manager

func init() {
	DefaultManager = &Manager{
		connections:         xsync.NewMapOf[string, Tracker](),
		uploadTemp:          atomic.NewInt64(0),
		downloadTemp:        atomic.NewInt64(0),
		uploadBlip:          atomic.NewInt64(0),
		downloadBlip:        atomic.NewInt64(0),
		uploadTotal:         atomic.NewInt64(0),
		downloadTotal:       atomic.NewInt64(0),
		directUploadTotal:   atomic.NewInt64(0),
		directDownloadTotal: atomic.NewInt64(0),
		proxyUploadTotal:    atomic.NewInt64(0),
		proxyDownloadTotal:  atomic.NewInt64(0),
		process:             &process.Process{Pid: int32(os.Getpid())},
		userStatistic:       xsync.NewMapOf[string, *UserStatistic](),
		ruleStatistic:       xsync.NewMapOf[string, *RuleStatistic](),
	}

	go DefaultManager.handle()
}

type Manager struct {
	connections         *xsync.MapOf[string, Tracker]
	uploadTemp          atomic.Int64
	downloadTemp        atomic.Int64
	uploadBlip          atomic.Int64
	downloadBlip        atomic.Int64
	uploadTotal         atomic.Int64
	downloadTotal       atomic.Int64
	directUploadTotal   atomic.Int64
	directDownloadTotal atomic.Int64
	proxyUploadTotal    atomic.Int64
	proxyDownloadTotal  atomic.Int64
	process             *process.Process
	userStatistic       *xsync.MapOf[string, *UserStatistic]
	ruleStatistic       *xsync.MapOf[string, *RuleStatistic]
	memory              uint64
}

func (m *Manager) Join(c Tracker) {
	m.connections.Store(c.ID(), c)
}

func (m *Manager) Leave(c Tracker) {
	m.connections.Delete(c.ID())
}

func (m *Manager) Get(id string) (c Tracker) {
	if value, ok := m.connections.Load(id); ok {
		c = value
	}
	return
}

func (m *Manager) Range(f func(c Tracker) bool) {
	m.connections.Range(func(key string, value Tracker) bool {
		return f(value)
	})
}

func (m *Manager) PushUploaded(size int64) {
	m.uploadTemp.Add(size)
	m.uploadTotal.Add(size)
}

func (m *Manager) PushUserUploaded(user string, size int64, isProxy bool) {
	if !C.UserStatistic {
		return
	}
	value, ok := m.userStatistic.Load(user)
	if ok {
		value.AddUpload(size, isProxy)
	} else {
		userStat := &UserStatistic{
			User:                user,
			DirectUploadTotal:   atomic.NewInt64(0),
			DirectDownloadTotal: atomic.NewInt64(0),
			ProxyUploadTotal:    atomic.NewInt64(0),
			ProxyDownloadTotal:  atomic.NewInt64(0),
		}
		userStat.AddUpload(size, isProxy)
		m.userStatistic.Store(user, userStat)
	}
}

func (m *Manager) PushRuleUploaded(rule string, size int64) {
	if !C.RuleStatistic || rule == "()" {
		return
	}
	value, ok := m.ruleStatistic.Load(rule)
	if ok {
		value.AddUpload(size)
	} else {
		ruleStat := &RuleStatistic{
			Rule:          rule,
			UploadTotal:   atomic.NewInt64(0),
			DownloadTotal: atomic.NewInt64(0),
		}
		ruleStat.AddUpload(size)
		m.ruleStatistic.Store(rule, ruleStat)
	}
}

func (m *Manager) PushDownloaded(size int64) {
	m.downloadTemp.Add(size)
	m.downloadTotal.Add(size)
}

func (m *Manager) PushUserDownloaded(user string, size int64, isProxy bool) {
	if !C.UserStatistic {
		return
	}
	value, ok := m.userStatistic.Load(user)
	if ok {
		value.AddDownload(size, isProxy)
	} else {
		userStat := &UserStatistic{
			User:                user,
			DirectUploadTotal:   atomic.NewInt64(0),
			DirectDownloadTotal: atomic.NewInt64(0),
			ProxyUploadTotal:    atomic.NewInt64(0),
			ProxyDownloadTotal:  atomic.NewInt64(0),
		}
		userStat.AddDownload(size, isProxy)
		m.userStatistic.Store(user, userStat)
	}
}

func (m *Manager) PushRuleDownloaded(rule string, size int64) {
	if !C.RuleStatistic || rule == "()" {
		return
	}
	value, ok := m.ruleStatistic.Load(rule)
	if ok {
		value.AddDownload(size)
	} else {
		ruleStat := &RuleStatistic{
			Rule:          rule,
			UploadTotal:   atomic.NewInt64(0),
			DownloadTotal: atomic.NewInt64(0),
		}
		ruleStat.AddDownload(size)
		m.ruleStatistic.Store(rule, ruleStat)
	}
}

func (m *Manager) TrafficStatistic() *TrafficStatistic {
	var userStatistic []*UserStatistic
	var ruleStatistic []*RuleStatistic
	var directUploadTotal, directDownloadTotal, proxyUploadTotal, proxyDownloadTotal int64
	m.userStatistic.Range(func(key string, value *UserStatistic) bool {
		directUploadTotal += value.DirectUploadTotal.Load()
		directDownloadTotal += value.DirectDownloadTotal.Load()
		proxyUploadTotal += value.ProxyUploadTotal.Load()
		proxyDownloadTotal += value.ProxyDownloadTotal.Load()
		userStatistic = append(userStatistic, value)
		return true
	})
	m.ruleStatistic.Range(func(key string, value *RuleStatistic) bool {
		ruleStatistic = append(ruleStatistic, value)
		return true
	})
	m.directUploadTotal.Store(directUploadTotal)
	m.directDownloadTotal.Store(directDownloadTotal)
	m.proxyUploadTotal.Store(proxyUploadTotal)
	m.proxyDownloadTotal.Store(proxyDownloadTotal)
	return &TrafficStatistic{
		UserStatistic: userStatistic,
		RuleStatistic: ruleStatistic,
	}
}

func (m *Manager) Now() (up int64, down int64) {
	return m.uploadBlip.Load(), m.downloadBlip.Load()
}

func (m *Manager) Memory() uint64 {
	m.updateMemory()
	return m.memory
}

func (m *Manager) Snapshot() *Snapshot {
	var connections []*TrackerInfo
	m.Range(func(c Tracker) bool {
		connections = append(connections, c.Info())
		return true
	})
	return &Snapshot{
		UploadTotal:         m.uploadTotal.Load(),
		DownloadTotal:       m.downloadTotal.Load(),
		DirectUploadTotal:   m.directUploadTotal.Load(),
		DirectDownloadTotal: m.directDownloadTotal.Load(),
		ProxyUploadTotal:    m.proxyUploadTotal.Load(),
		ProxyDownloadTotal:  m.proxyDownloadTotal.Load(),
		Connections:         connections,
		Memory:              m.memory,
	}
}

func (m *Manager) updateMemory() {
	stat, err := m.process.MemoryInfo()
	if err != nil {
		return
	}
	m.memory = stat.RSS
}

func (m *Manager) ResetStatistic() {
	m.uploadTemp.Store(0)
	m.uploadBlip.Store(0)
	m.uploadTotal.Store(0)
	m.downloadTemp.Store(0)
	m.downloadBlip.Store(0)
	m.downloadTotal.Store(0)
	m.userStatistic = xsync.NewMapOf[string, *UserStatistic]()
	m.ruleStatistic = xsync.NewMapOf[string, *RuleStatistic]()
}

func (m *Manager) handle() {
	ticker := time.NewTicker(time.Second)

	for range ticker.C {
		m.uploadBlip.Store(m.uploadTemp.Load())
		m.uploadTemp.Store(0)
		m.downloadBlip.Store(m.downloadTemp.Load())
		m.downloadTemp.Store(0)
	}
}

type Snapshot struct {
	DownloadTotal       int64          `json:"downloadTotal"`
	UploadTotal         int64          `json:"uploadTotal"`
	Connections         []*TrackerInfo `json:"connections"`
	Memory              uint64         `json:"memory"`
	DirectUploadTotal   int64          `json:"directUploadTotal"`
	DirectDownloadTotal int64          `json:"directDownloadTotal"`
	ProxyUploadTotal    int64          `json:"proxyUploadTotal"`
	ProxyDownloadTotal  int64          `json:"proxyDownloadTotal"`
}

type TrafficStatistic struct {
	UserStatistic []*UserStatistic `json:"userStatistic"`
	RuleStatistic []*RuleStatistic `json:"ruleStatistic"`
}
