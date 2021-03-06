package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"torrentClient/parser/env"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

const (
	fileSliceIndexes       = "slices"
)

type manager struct {
	conn	*redis.Client
}

type PriorityUpdateMsg struct {
	TorrentId	string
	FileName string
	ByteIdx	int64
}

var Manager manager

func (m *manager) GetSliceIndexesKey(fileName string) string {
	return fmt.Sprintf("%s:%s", fileSliceIndexes, fileName)
}

func (m *manager) AddSliceIndexForFile(fileName string, sliceByteIdx ...int64) {
	for _, idx := range sliceByteIdx {
		_, err := m.conn.SAdd(m.GetSliceIndexesKey(fileName), idx).Result()
		if err != nil {
			logrus.Errorf("Error AddSliceIndexForFile: %v", err)
		} else {
			logrus.Debugf("Added slices for file %v: %v", fileName, idx)
		}
	}
}

func (m *manager) DeleteSliceIndexesSet(fileName string) {
	if _, err := m.conn.Del(m.GetSliceIndexesKey(fileName)).Result(); err != nil {
		logrus.Errorf("Error deleting key: %v", err)
	}
}

func (m *manager) GetLoadPriorityUpdatesChan(ctx context.Context, fileId string) chan PriorityUpdateMsg {
	sub := m.conn.PSubscribe(fmt.Sprintf("%s:*", fileId))

	updatesChan := make(chan PriorityUpdateMsg, 100)

	go func() {
		for {
			select {
			case <- ctx.Done():
				close(updatesChan)
				return
			case msg := <- sub.Channel():
				logrus.Debugf("Got priority msg '%v' in chan '%v'", msg.Payload, msg.Channel)
				fileName := strings.Split(msg.Channel, ":")[1]
				if byteIdx, err := strconv.Atoi(msg.Payload); err != nil {
					logrus.Errorf("Error parsing msg payload for byte index: %v", err)
				} else {
					updatesChan <- PriorityUpdateMsg{FileName: fileName, TorrentId: fileId, ByteIdx: int64(byteIdx)}
				}
			}
		}
	}()

	return updatesChan
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



