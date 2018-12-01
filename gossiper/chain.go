package gossiper

import (
    "bytes"
    "encoding/hex"
    "github.com/jfperren/Peerster/common"
    "sync"
    "crypto/rand"
    "time"
)

type BlockChain struct {

    Candidates []common.TxPublish

    Files       map[string]*common.File
    Blocks      map[string]*common.Block
    Latest      [32]byte
    IsNew       bool
    MinedBlocks chan *common.Block
    MiningTime  int64

    lock        *sync.RWMutex
}

func NewBlockChain() *BlockChain {

    return &BlockChain{
        Candidates:     make([]common.TxPublish, 0),
        Files:          make(map[string]*common.File),
        Blocks:         make(map[string]*common.Block),
        IsNew:          true,
        MinedBlocks:    make(chan *common.Block, 2),
        MiningTime:     0,
        lock:           &sync.RWMutex{},
    }
}

func NewTransaction(metaFile *MetaFile) *common.TxPublish {
    return &common.TxPublish{
        common.File{
            metaFile.Name,
            int64(metaFile.Size),
            metaFile.Hash,
        },
        common.TransactionHopLimit,
    }
}

func (gossiper *Gossiper) waitForNewBlocks() {

    for {


        // Wait for blocks
        block := <- gossiper.BlockChain.MinedBlocks


        publish := &common.BlockPublish{
            Block:    *block,
            HopLimit: common.BlockHopLimit,
        }

        gossiper.broadcastToNeighbors(publish.Packed())
        common.LogFoundBlock(block)
    }
}

// Atomically test and append transaction
func (bc *BlockChain) tryAddFile(candidate *common.TxPublish) bool {

    bc.lock.Lock()
    defer bc.lock.Unlock()

    _, found := bc.Files[candidate.File.Name]

    if found {
        common.DebugIgnoreTransactionAlreadyInChain(candidate)
        return false
    }

    for _, otherCandidates := range bc.Candidates {
        if candidate.File.Name == otherCandidates.File.Name {
            common.DebugIgnoreTransactionAlreadyCandidate(candidate)
            return false
        }
    }

    bc.Candidates = append(bc.Candidates, *candidate)
    common.DebugAddCandidateTransaction(candidate)

    return true
}

// Atomically test and append block
func (bc *BlockChain) tryAddBlock(candidate *common.Block) bool {

    bc.lock.Lock()
    defer bc.lock.Unlock()

    hash := candidate.Hash()

    if !(hash[0] == 0 && hash[1] == 0) {
        common.DebugIgnoreBlockIsNotValid(candidate)
        return false
    }

    _, found := bc.Blocks[hex.EncodeToString(hash[:])]

    if found {
        common.DebugIgnoreBlockAlreadyPresent(candidate)
        return false
    }

    if !bc.IsNew && bytes.Compare(candidate.PrevHash[:], bc.Latest[:]) != 0 {
        common.DebugIgnoreBlockPrevDoesntMatch(candidate, bc.Latest)
        return false
    }

    bc.Blocks[hex.EncodeToString(hash[:])] = candidate
    bc.Latest = hash
    bc.IsNew = false

    for _, transaction := range candidate.Transactions {
        bc.Files[transaction.File.Name] = &transaction.File
    }

    newCandidates := make([]common.TxPublish, 0)

    for _, transaction := range bc.Candidates {
        _, found := bc.Files[transaction.File.Name]

        if !found {
            newCandidates = append(newCandidates, transaction)
        }
    }

    bc.Candidates = newCandidates

    common.LogChain(bc.allBlocks())

    return true
}

func (bc *BlockChain) mine() {

    for {

        var nonce [32]byte
        bc.MiningTime = time.Now().UnixNano()

        _, err := rand.Read(nonce[:])

        if err != nil {
            continue
        }

        bc.lock.RLock()

        candidate := &common.Block {
            bc.Latest,
            nonce,
            bc.Candidates,
        }

        bc.lock.RUnlock()

        hash := candidate.Hash()

        if hash[0] == 0 && hash[1] == 0 {

            if bc.tryAddBlock(candidate) {

                nanoSeconds := time.Duration(2 * (time.Now().UnixNano() - bc.MiningTime))
                time.Sleep(nanoSeconds * time.Nanosecond)

                bc.MinedBlocks <- candidate
            }
        }
    }
}

func (bc *BlockChain) allBlocks() []*common.Block {

    hash := bc.Latest
    allBlocks := make([]*common.Block, 0)

    for {
        block, found := bc.Blocks[hex.EncodeToString(hash[:])]

        if !found {
            return allBlocks
        }

        allBlocks = append(allBlocks, block)
        hash = block.PrevHash
    }
}