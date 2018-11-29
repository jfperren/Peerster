package gossiper

import (
    "github.com/jfperren/Peerster/common"
    "strings"
    "sync"
    "time"
)

type SpamDetector struct {

    lastSeen   	 map[string]int64
    lock	 	 *sync.RWMutex

}

func NewSpamDetector() *SpamDetector {
    return &SpamDetector{
        make(map[string]int64),
        &sync.RWMutex{},
    }
}

func searchRequestID(request *common.SearchRequest) string {
    return request.Origin + ":" + strings.Join(request.Keywords, common.SearchKeywordSeparator)
}

func (sd *SpamDetector) shouldProcessSearchRequest(request *common.SearchRequest) bool {

    now := time.Now().UnixNano()

    sd.lock.RLock()

    prev, found := sd.lastSeen[searchRequestID(request)]

    sd.lock.RUnlock()

    if found && (now - prev) < common.SearchRequestTimeThreshold {
        // Do not process
        return false
    }

    sd.lock.Lock()
    defer sd.lock.Unlock()

    sd.lastSeen[searchRequestID(request)] = now

    return true
}