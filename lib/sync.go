package lib

import (
	"github.com/pubnub/go/messaging"
	"sync"
	"time"
)

var syncInterval = time.Minute * 10

type TimeSyncer struct {
	pubnub *messaging.Pubnub

	start time.Time
	timetoken time.Time

	synced bool
	syncResults chan bool

	sync.RWMutex
}

var instance *TimeSyncer
var once sync.Once

func StartTimeSync(pubnub *messaging.Pubnub) *TimeSyncer {
	once.Do(func() {
		instance = &TimeSyncer{
			pubnub: pubnub,
			syncResults: make(chan bool, 1),
		}

		instance.syncTime()
	})

	return instance
}

func (syncer *TimeSyncer) AwaitSynced() bool {
	if syncer.synced {
		return true
	}

	syncResult := <- syncer.syncResults

	if syncResult {
		syncer.synced = true
	}

	return syncResult
}

func (syncer *TimeSyncer) SyncedTime() time.Time {
	syncer.RLock()
	timeChange := time.Now().Sub(syncer.start)
	clientTime := syncer.timetoken.Add(timeChange)
	syncer.RUnlock()

	return clientTime
}

func (syncer *TimeSyncer) syncTime() {
	startTime := time.Now()

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)

	go func() {
		syncer.syncResults <- syncer.handleTimeToken(successChannel, errorChannel, startTime)
	}()
	go syncer.pubnub.GetTime(successChannel, errorChannel)

	time.AfterFunc(syncInterval, syncer.syncTime) // Sync time every syncInterval to prevent clock drift
}

func (syncer *TimeSyncer) handleTimeToken(successChannel, errorChannel chan []byte, startTime time.Time) bool {
	for {
		select {
		case success, ok := <-successChannel:
			if !ok {
				break
			}
			if string(success) != "[]" {
				logger.Debugf("Successfully got timetoken: %v", string(success))
			}

			results, err := jsonToArray(success)
			if err != nil || len(results) == 0 {
				return false
			}

			syncer.updateSyncState(int64(results[0].(float64)), startTime)
			return true
		case failure, ok := <-errorChannel:
			if !ok {
				break
			}
			if string(failure) != "[]" {
				logger.Errorf("Failed to get timetoken: %v", string(failure))
			}

			return false
		case <- time.After(time.Duration(messaging.GetNonSubscribeTimeout()) * time.Second):
			logger.Errorf("Timeout out waiting for time token!")
			return false
		}
	}

	return false
}

func (syncer *TimeSyncer) updateSyncState(timetoken int64, startTime time.Time) {
	syncer.Lock()
	durationToken := time.Duration(timetoken/10000) * time.Millisecond

	syncer.timetoken = time.Unix(0, durationToken.Nanoseconds())
	syncer.start = startTime
	syncer.Unlock()
}