package redis

import (
	"fmt"

	"hypertube_storage/parser/env"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

const (
	fileSliceIndexes       = "slices"
)

type manager struct {
	conn	*redis.Client
}

var Manager manager

func (m *manager) GetSliceIndexesKey(fileName string) string {
	return fmt.Sprintf("%s:%s", fileSliceIndexes, fileName)
}

func (m *manager) GetSliceIndexesForFile(fileName string) (slices []int64) {
	err := m.conn.Sort(m.GetSliceIndexesKey(fileName), &redis.Sort{Order: "ASC"}).ScanSlice(&slices)
	if err != nil {
		logrus.Errorf("Error GetSliceIndexesForFile: %v", err)
	}

	return slices
}

func (m *manager) PubPriorityByteIdx(fileId, fileName string, idx int64) {
	if res := m.conn.Publish(fmt.Sprintf("%s:%s", fileId, fileName), idx); res.Err() != nil {
		logrus.Errorf("Error publish priority: %v", res.Err())
	}
	logrus.Debugf("Published new priority=%v for %v:%v", idx, fileId, fileName)
}

func (m *manager) InitConnection() {
	m.conn = redis.NewClient(&redis.Options{
		Addr: env.GetParser().GetRedisDbAddr(),
		Password: env.GetParser().GetRedisDbPasswd(),
		DB: 0,
	})

	if err := m.conn.Ping().Err(); err != nil {
		logrus.Fatalf("Error pinging redis: %v", err)
	}
}

func (m *manager) CloseConnection() {
	if err := m.conn.Close(); err != nil {
		logrus.Errorf("Error closing redis conn: %v", err)
	}
}



