package gossiper

import (
    "encoding/hex"
    "fmt"
    "github.com/jfperren/Peerster/common"
    "regexp"
    "strings"
    "sync"
    "time"
)

// This contains fields and functions related to the decentralized search
// algorithm of Peerster.
type SearchEngine struct {

    activeSearches  map[string]*ActiveSearch    // All searches that did not complete
    fileMaps        map[string]*FileMap         // Keeps track of file chunks location
    results         []*common.SearchResult      // All results received
    lock            *sync.RWMutex               // Synchronize access
}

// Represents a search request that has not yet completed.
type ActiveSearch struct {

    id          string                  // Unique ID to reference the search
    keywords    []string                // Keywords used
    matches     []*common.SearchResult  // Current matches found so far
}

// Contains, for each file in the results, a mapping to the location of their chunks.
type FileMap struct {

    FileName        string                      // Name of the file
    FileHash        []byte                      // MetaHash of the file
    chunkCount      uint64                      // Count of chunks
    chunkMap        map[uint64]map[string]bool  // For each chunk number, list of seeds
}

// Initialize search engine
func NewSearchEngine() *SearchEngine {
    return &SearchEngine{
        activeSearches: make(map[string]*ActiveSearch),
        results: make([]*common.SearchResult, 0),
        fileMaps: make(map[string]*FileMap),
        lock: &sync.RWMutex{},
    }
}

// Check if a given fileMap contains at least one seed per chunk
func (fileMap *FileMap) isComplete() bool {

    for i := uint64(1); i <= fileMap.chunkCount; i++ {
        if len(fileMap.chunkMap[i]) == 0 {
            return false
        }
    }

    return true
}

// Random peer from which to download the metaFile of a given file.
func (fileMap *FileMap) peerForMetafile(counter int) (string, bool) {
    peer, found := fileMap.peerForChunk(1, counter)
    return peer, found
}

// Random peer from which to download a given chunk of a file.
func (fileMap *FileMap) peerForChunk(chunkId uint64, counter int) (string, bool) {

    potentialPeers, found := fileMap.chunkMap[chunkId]

    if !found { return "", false }

    index := counter % len(potentialPeers)

    i := 0

    for peer, _ := range potentialPeers {

        if i == index {
            return peer, true
        }

        i++
    }

    return "", false
}


// Get the fileMap for a given meta-hash
func (se *SearchEngine) FileMap(hash []byte) (*FileMap, bool) {

    se.lock.RLock()
    defer se.lock.RUnlock()

    fileMap, found := se.fileMaps[hex.EncodeToString(hash)]

    return fileMap, found
}

// Register a new search as active search so that we can keep track of its results
func (se *SearchEngine) createNewActiveSearch(keywords []string) string {

    searchRequest := &ActiveSearch{
        id: fmt.Sprintf("%v:%v", time.Now().UnixNano(), strings.Join(keywords,common.SearchKeywordSeparator)),
        keywords: keywords,
        matches: make([]*common.SearchResult, 0),
    }

    se.lock.Lock()
    defer se.lock.Unlock()

    se.activeSearches[searchRequest.id] = searchRequest

    return searchRequest.id
}

// Process results from a current or previous search.
func (se *SearchEngine) StoreResults(results []*common.SearchResult, origin string) {

    se.lock.Lock()
    defer se.lock.Unlock()

    for _, result := range results {

        common.LogMatch(*result, origin)

        for _, search := range se.activeSearches {

            if Match(result.FileName, search.keywords) {
                search.matches = append(search.matches, result)
            }
        }

        fileMap, found := se.fileMaps[hex.EncodeToString(result.MetafileHash)]

        if !found {
            fileMap = &FileMap{
                result.FileName,
                result.MetafileHash,
                result.ChunkCount,
                make(map[uint64]map[string]bool),
            }
            se.fileMaps[hex.EncodeToString(result.MetafileHash)] = fileMap
        }

        fillChunkMap(fileMap.chunkMap, result, origin)
        se.results = append(se.results, result)
    }

    for searchId, _ := range se.activeSearches {

        if se.hasCompleted(searchId, false) {
            common.LogSearchFinished()
            delete(se.activeSearches, searchId)
        }
    }
}

//
//  ADDITIONAL FILE SYSTEM FUNCTIONS
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
        FileName:       metaFile.Name,
        MetafileHash:   metaFile.Hash,
        ChunkMap:       fs.chunkMap(metaFile),
        ChunkCount:     uint64(metaFile.countOfChunks()),
    }
}

func (fs *FileSystem) chunkMap(metaFile *MetaFile) []uint64 {

    chunkMap := make([]uint64, 0)

    for i := 1; i <= metaFile.countOfChunks(); i++ {

        hash := metaFile.hashAt(i - 1)
        _, found  := fs.getChunk(hash)

        if found {
            chunkMap = append(chunkMap, uint64(i))
        }
    }

    return chunkMap
}

// Check if a search is completed by checking the number of matches
func (se *SearchEngine) hasCompleted(searchId string, lock bool) bool {

    if lock {
        se.lock.RLock()
        defer se.lock.RUnlock()
    }

    search, found := se.activeSearches[searchId]

    if !found {
        return true
    }

    chunkMaps := make(map[string]map[uint64]bool)
    sizes := make(map[string]uint64)

    for _, result := range search.matches {

        chunkMap, found := chunkMaps[result.FileName]

        if !found {
            chunkMap = make(map[uint64]bool)
            chunkMaps[result.FileName] = chunkMap
        }

        for _, chunkId := range result.ChunkMap {
            chunkMap[chunkId] = true
        }

        sizes[result.FileName] = result.ChunkCount
    }

    count := 0

    FILE_LOOP:
    for name, chunkMap := range chunkMaps {

        size, found := sizes[name]

        if !found {
            continue FILE_LOOP
        }

        for i := uint64(1); i <= size; i++ {

            _, found := chunkMap[i]

            if !found {
                continue FILE_LOOP
            }
        }

        count++
    }

    return count >= 2
}

//
//  OTHER HELPERS
//

// Compare filename and keyword for potential match
func Match(name string, keywords []string) bool {

    for _, keyword := range keywords {

        match, _ := regexp.MatchString(keyword, name)

        if match {
            return true
        }
    }

    return false
}

// Fill a given chunkMap with the contents of a search result
func fillChunkMap(chunkMap map[uint64]map[string]bool, result *common.SearchResult, origin string) {

    for _, chunk := range result.ChunkMap {

        seeds, found := chunkMap[chunk]

        if !found {
            seeds = make(map[string]bool)
            chunkMap[chunk] = seeds
        }

        seeds[origin] = true
    }
}

//
//  GOSSIPER FUNCTIONS
//

func (gossiper *Gossiper) RingSearch(keywords []string, budget uint64) {

    timestamp := time.Now().Unix()

    searchId := gossiper.SearchEngine.createNewActiveSearch(keywords)

    if budget == common.SearchNoBudget {
        gossiper.ringSearchInternal(searchId, keywords, common.DefaultSearchBudget, timestamp, true)
    } else {
        gossiper.ringSearchInternal(searchId, keywords, budget, timestamp, false)
    }
}

func (gossiper *Gossiper) newSearchRequest(budget uint64, keywords []string) *common.SearchRequest {
    return &common.SearchRequest{
        Origin: gossiper.Name,
        Budget:	budget,
        Keywords: keywords,
    }
}

func (gossiper *Gossiper) ringSearchInternal(searchId string, keywords []string, budget uint64, timestamp int64, increasing bool) {

    if budget > common.MaxSearchBudget && increasing {
        common.DebugSearchTimeout(keywords)
        return
    }

    if gossiper.SearchEngine.hasCompleted(searchId, true) {
        return
    }

    common.DebugStartSearch(keywords, budget, increasing)

    request := gossiper.newSearchRequest(budget, keywords)
    gossiper.broadcastSearchRequest(request)

    if !increasing {
        return
    }

    time.Sleep(common.SearchRequestBudgetIncreaseDT)

    gossiper.ringSearchInternal(searchId, keywords, budget * 2, timestamp, true)
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

// Forward the search
func (gossiper *Gossiper) broadcastSearchRequest(searchRequest *common.SearchRequest) {
    gossiper.broadcastToNeighbors(searchRequest.Packed())
}
