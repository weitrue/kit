package crons

import (
	"go.uber.org/zap"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/weitrue/kit/logger/xzap"
)

var logger *zap.Logger

type TestJob struct{}

func (TestJob) Frequency() string {
	return "* * * * * *" // Every second
}

func (TestJob) Run() {
	log.Println("Test Crontab", time.Now().Unix())
}

func TestCronJob(t *testing.T) {
	CronManger, err := NewCronManage(Second, logger)
	assert.Nil(t, err)
	t1 := TestJob{}
	_, err = CronManger.RegisterFunc(t1.Frequency(), t1.Run)
	assert.Nil(t, err)

	go CronManger.Start()
	log.Println("Cron started success")
	time.Sleep(10 * time.Second)
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
