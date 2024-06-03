package crons

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/weitrue/kit/logger/xzap"
)

var logger *zap.Logger

type TestJob struct {
	entityId cron.EntryID
}

func (j *TestJob) Frequency() string {
	return "@every 5s" // Every second
}

func (j *TestJob) Run() {
	log.Println("Test Crontab", time.Now().Unix())
}

func (j *TestJob) SetEntityId(entityId cron.EntryID) {
	j.entityId = entityId
}

func TestCronJob(t *testing.T) {
	CronManger, err := NewCronManage(Second, logger)
	assert.Nil(t, err)

	go CronManger.Start()
	log.Println("Cron started success")
	t1 := TestJob{}
	_, err = CronManger.Register(&t1)
	assert.Nil(t, err)
	fmt.Print(t1.entityId)
	time.Sleep(16 * time.Second)
	CronManger.RemoveEntity(t1.entityId)
	time.Sleep(30 * time.Second)
}

func init() {
	logger, _ = xzap.SetUp(xzap.Config{
		ServiceName: "cron",
		Mode:        "file",
		Path:        "logs/cron",
		Level:       "info",
		Compress:    false,
		KeepDays:    7,
	})
}
