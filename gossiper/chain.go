package gossiper

import (
    "bytes"
    "crypto/rand"
    "encoding/hex"
    "github.com/jfperren/Peerster/common"
    "sync"
)

type BlockChain struct {

    Candidates []common.TxPublish

    Files       map[string]*common.File
    Blocks      map[string]*common.Block
    Latest      [32]byte
    IsNew       bool
    MinedBlocks chan *common.Block

    lock        *sync.RWMutex
}

func NewBlockChain() *BlockChain {

    return &BlockChain{
        Candidates:     make([]common.TxPublish, 0),
        Files:          make(map[string]*common.File),
        Blocks:         make(map[string]*common.Block),
        IsNew:          true,
        MinedBlocks:    make(chan *common.Block, 2),
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
        return false
    }

    for _, candidate := range bc.Candidates {
        if candidate.File.Name == candidate.File.Name {
            return false
        }
    }

    bc.Candidates = append(bc.Candidates, *candidate)

    return true
}

// Atomically test and append block
func (bc *BlockChain) tryAddBlock(candidate *common.Block) bool {

    bc.lock.Lock()
    defer bc.lock.Unlock()

    hash := candidate.Hash()

    if !(hash[0] == 0 && hash[1] == 0) {
        return false
    }

    _, found := bc.Blocks[hex.EncodeToString(hash[:])]

    if found {
        return false
    }

    if bytes.Compare(candidate.PrevHash[:], bc.Latest[:]) != 0 && !bc.IsNew {
        return false
    }

    bc.Blocks[hex.EncodeToString(hash[:])] = candidate
    bc.Latest = hash
    bc.IsNew = true
    common.LogChain(bc.Latest, bc.Blocks)

    return true
}

func (bc *BlockChain) mine() {

    for {

        var nonce [32]byte

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

        if hash[0] == 0 && hash[1] == 0 && hash[2] < 50 {

            if bc.tryAddBlock(candidate) {

                bc.MinedBlocks <- candidate
            }
        }
    }
}
}