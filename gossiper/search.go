package gossiper

import (
    "bytes"
    "encoding/hex"
    "github.com/jfperren/Peerster/common"
    "regexp"
    "sync"
    "time"
)

type SearchEngine struct {

    history      []*SearchLog
    fileMaps     map[string]*FileMap
    lock	 	 *sync.RWMutex
}

type SearchLog struct {

    result      *common.SearchResult
    origin      string
    timestamp   int64
}

type FileMap struct {

    FileName        string
    FileHash        []byte
    chunkCount      uint64
    chunkMap        map[string]uint64

}

func NewSearchEngine() *SearchEngine {
    return &SearchEngine{
        make([]*SearchLog, 0),
        make(map[string]*FileMap),
        &sync.RWMutex{},
    }
}

//
//
//

func (se *SearchEngine) FileMap(hash []byte) (*FileMap, bool) {

    se.lock.RLock()
    defer se.lock.RUnlock()

    fileMap, found := se.fileMaps[hex.EncodeToString(hash)]

    return fileMap, found
}

func (fileMap *FileMap) peerForChunk(chunkId uint64, counter int) (string, bool) {

    for peer, highestChunk := range fileMap.chunkMap {

        if highestChunk == fileMap.chunkCount - 1 {
            return peer, true
        }
    }

    return "", false
}

/// Process results from a current or previous search.
func (se *SearchEngine) StoreResults(results []*common.SearchResult, origin string) {

    se.lock.Lock()
    defer se.lock.Unlock()

    for _, result := range results {

        common.LogMatch(result.FileName, origin, result.MetafileHash, result.ChunkMap)
        log := &SearchLog{result, origin, time.Now().Unix()}
        se.history = append(se.history, log)

        fileMap, found := se.fileMaps[hex.EncodeToString(result.MetafileHash)]

        if !found {
            fileMap = &FileMap{
                result.FileName,
                result.MetafileHash,
                result.ChunkCount,
                make(map[string]uint64),
            }
            se.fileMaps[hex.EncodeToString(result.MetafileHash)] = fileMap
        }

        fileMap.chunkMap[origin] = result.ChunkMap[len(result.ChunkMap) - 1]
    }
}

//
//
//

func (fs *FileSystem) Search(keywords []string) []*common.SearchResult {

    results := make([]*common.SearchResult, 0)

    for _, metaFile := range fs.metaFiles {
        if Match(metaFile.Name, keywords) {
            results = append(results, fs.newSearchResult(metaFile))
        }
    }

    common.DebugSearchResults(keywords, results)

    return results
}

func (fs *FileSystem) newSearchResult(metaFile *MetaFile) *common.SearchResult {
    return &common.SearchResult{
        metaFile.Name,
        metaFile.Hash,
        fs.chunkMap(metaFile),
        uint64(metaFile.countOfChunks()),
    }
}

func (fs *FileSystem) chunkMap(metaFile *MetaFile) []uint64 {

    chunkMap := make([]uint64, 0)

    for i := 0; i < metaFile.countOfChunks(); i++ {

        hash := metaFile.hashAt(i)
        _, found  := fs.getChunk(hash)

        if found {
            chunkMap = append(chunkMap, uint64(i))
        }
    }

    return chunkMap
}

func Match(name string, keywords []string) bool {

    for _, keyword := range keywords {

        match, _ := regexp.MatchString(keyword, name)

        if match {
            return true
        }
    }

    return false
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

        if result.ChunkMap[len(result.ChunkMap) - 1] == result.ChunkCount - 1 {
            count++
        }
    }

    common.DebugSearchStatus(count, keywords)

    return count
}

func (se *SearchEngine) buildFileMap(fileHash []byte) (bool, *FileMap) {

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

    if budget > common.MaxSearchBudget && increasing {
        common.DebugSearchTimeout(keywords)
        return
    }

    if gossiper.SearchEngine.shouldStopSearch(keywords, timestamp) {
        common.LogSearchFinished()
        return
    }

    common.DebugStartSearch(keywords, budget, increasing)

    request := gossiper.newSearchRequest(budget, keywords)
    gossiper.broadcastSearchRequest(request)

    if !increasing {
        return
    }

    time.Sleep(common.SearchRequestBudgetIncreaseDT)

    gossiper.ringSearchInternal(keywords, budget * 2, timestamp, true)
}

// Forward a Search request
func (gossiper *Gossiper) forwardSearchRequest(searchRequest *common.SearchRequest, origin string) {


    searchRequest.Budget--
    budgets := common.SplitBudget(searchRequest.Budget, len(gossiper.Router.Peers) - 1)

    i := 0

    for _, peer := range gossiper.Router.Peers {

        if origin == peer {
            continue
        }

        request := common.CopySearchRequest(searchRequest, budgets[i])
        common.DebugForwardSearchRequest(request, peer)
        go gossiper.sendToNeighbor(peer, request.Packed())

        i++
    }
}

func (gossiper *Gossiper) broadcastSearchRequest(searchRequest *common.SearchRequest) {
    gossiper.broadcastToNeighbors(searchRequest.Packed())
}