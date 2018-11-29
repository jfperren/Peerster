package gossiper

import (
    "bytes"
    "github.com/jfperren/Peerster/common"
    "regexp"
    "sync"
    "time"
)

type SearchEngine struct {

    lastSeen   	 map[string]int64
    history      []*SearchLog

    lock	 	 *sync.RWMutex
}

type SearchLog struct {

    result      *common.SearchResult
    origin      string
    timestamp   int64
}

type FileMap struct {

    FileName     string
    FileHash    []byte
    chunkCount  uint64
    origins     map[string]uint64

}

func NewSearchEngine() *SearchEngine {
    return &SearchEngine{
        make(map[string]int64),
        make([]*SearchLog, 0),
        &sync.RWMutex{},
    }
}

func (se *SearchEngine) shouldProcessRequest(request *common.SearchRequest) bool {

    now := time.Now().Unix()

    se.lock.RLock()

    prev, found := se.lastSeen[request.Key()]

    se.lock.RUnlock()

    if found && (now - prev) < common.SearchRequestTimeThreshold {
        // Do not process
        return false
    }

    se.lock.Lock()
    defer se.lock.Unlock()

    se.lastSeen[request.Key()] = now

    return true
}

//
//
//

func (fs *FileSystem) StartSearch(request *common.SearchRequest) {

    fs.lock.Lock()
    defer fs.lock.Unlock()


}

//
//
//

/// Process results from a current or previous search.
func (se *SearchEngine) StoreResults(results []*common.SearchResult, origin string) {

    se.lock.Lock()
    defer se.lock.Unlock()

    for _, result := range results {

        log := &SearchLog{result, origin, time.Now().Unix()}
        se.history = append(se.history, log)

    }
}

//
//
//

func (fs *FileSystem) Search(keywords []string) []*common.SearchResult {

    results := make([]*common.SearchResult, 0)

    for _, metaFile := range fs.metaFiles {
        if match(metaFile.Name, keywords) {
            results = append(results, fs.newSearchResult(metaFile))
        }
    }

    return results
}

func match(name string, keywords []string) bool {

    for _, keyword := range keywords {

        match, _ := regexp.MatchString("*" + keyword + "*", name)

        if match {
            return true
        }
    }

    return false
}

func (fs *FileSystem) newSearchResult(metaFile *MetaFile) *common.SearchResult {
    return &common.SearchResult{
        metaFile.Name,
        metaFile.Hash,
        make([]uint64, 0),
        10,
    }
}

//
//
//

func (se *SearchEngine) shouldStopSearch(keywords []string, since int64) bool {
    return se.countOfResults(keywords, since) >= 2
}

func (se *SearchEngine) countOfResults(keywords []string, since int64) int {

    se.lock.RLock()
    defer se.lock.RUnlock()

    count := 0

    for i := len(se.history) - 1; i >= 0; i-- {

        result := se.history[i].result

        if result.ChunkMap[len(result.ChunkMap) - 1] == result.ChunkCount {
            count++
        }
    }

    return count
}

func (se *SearchEngine) getFileMap(fileHash []byte) (bool, *FileMap) {

    origins := make(map[string]uint64)
    found := false

    var count uint64
    var fileName string

    for _, log := range se.history {

        if bytes.Compare(log.result.MetafileHash, fileHash) == 0 {

            count = log.result.ChunkCount
            fileName = log.result.FileName

            origins[log.origin] = log.result.ChunkCount

            if log.result.ChunkCount == log.result.ChunkMap[len(log.result.ChunkMap) - 1] {
                found = true
            }
        }
    }

    if !found {
        return false, nil
    }

    return true, &FileMap{fileName, fileHash, count, origins}
}

// --
// -- Locations
// --

// func (fs *FileSystem)

func (gossiper *Gossiper) RingSearch(keywords []string, budget uint64) {

    timestamp := time.Now().Unix()

    if budget == common.SearchNoBudget {
        gossiper.ringSearchInternal(keywords, common.DefaultSearchBudget, timestamp, true)
    } else {
        gossiper.ringSearchInternal(keywords, budget, timestamp, false)
    }
}

func (gossiper *Gossiper) newSearchRequest(budget uint64, keywords []string) *common.SearchRequest {
    return &common.SearchRequest{
        Origin: gossiper.Name,
        Budget:	budget,
        Keywords: keywords,
    }
}

func (gossiper *Gossiper) ringSearchInternal(keywords []string, budget uint64, timestamp int64, increasing bool) {

    if budget > common.MaxSearchBudget {

        if increasing {
            // Timeout
        } else {
            // Error budget too high
        }

        return
    }

    if gossiper.SearchEngine.shouldStopSearch(keywords, timestamp) {
        //
        return
    }

    request := gossiper.newSearchRequest(budget, keywords).Packed()
    gossiper.broadcastToNeighbors(request, false)

    if !increasing {
        return
    }

    time.Sleep(common.SearchRequestBudgetIncreaseDT)

    gossiper.ringSearchInternal(keywords, budget * 2, timestamp, true)
}

// Forward a GossipPacket
func (gossiper *Gossiper) forwardSearchRequest(searchRequest *common.SearchRequest) {

    searchRequest.Budget--;

    budgets := common.SplitBudget(searchRequest.Budget, len(gossiper.Router.Peers))

    for i, budget := range budgets {

        if budget != 0 {
            request := common.CopySearchRequest(searchRequest, budget).Packed()
            peer := gossiper.Router.Peers[i]

            go gossiper.sendToNeighbor(peer, request)
        }
    }
}