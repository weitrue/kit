package crons

import (
	"fmt"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

const (
	Second = "Second"
	Minute = "Minute"
)

type Job interface {
	cron.Job
	Frequency() string
	SetEntityId(entityId cron.EntryID)
}

type CronManage struct {
	cron *cron.Cron
}

func NewCronManage(mode string, logger *zap.Logger) (*CronManage, error) {
	var c *cron.Cron
	switch mode {
	case Second:
		c = cron.New(cron.WithSeconds(), cron.WithLogger(NewLogger(logger)))
	case Minute:
		c = cron.New(cron.WithLogger(NewLogger(logger)))
	default:
		return nil, errors.New(fmt.Sprintf("err mode, unsupported mode %s", mode))
	}

	return &CronManage{
		cron: c,
	}, nil
}

func (m *CronManage) Register(job ...Job) (int, error) {
	msg := ""
	count := 0
	for i, j := range job {
		entityId, err := m.cron.AddFunc(j.Frequency(), j.Run)
		if err != nil {
			msg += fmt.Sprintf("%d:%v ", i, err)
		}
		j.SetEntityId(entityId)
	}

	if len(msg) > 0 {
		return len(job) - count, errors.New(fmt.Sprintf("Register exsit err, err:%s", msg))
	}

	return len(job) - count, nil
}

func (m *CronManage) RegisterFunc(frequency string, cmd func()) (int, error) {
	count, err := m.cron.AddFunc(frequency, cmd)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (m *CronManage) Start() {
	m.cron.Start()
}

func (m *CronManage) Stop() {
	m.cron.Stop()
}

func (m *CronManage) Entries() []cron.Entry {
	return m.cron.Entries()
}

func (m *CronManage) RemoveEntity(entityId cron.EntryID) {
	entries := m.Entries()
	for _, v := range entries {
		if v.ID == entityId {
			m.cron.Remove(entityId)
		}
	}
}
