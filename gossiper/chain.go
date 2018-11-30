package gossiper

import (
    "bytes"
    "crypto/rand"
    "encoding/hex"
    "github.com/jfperren/Peerster/common"
    "sync"
)

type BlockChain struct {

    candidates []common.TxPublish

    files       map[string]*common.File
    blocks      map[string]*common.Block
    previous    [32]byte
    isNew       bool
    minedBlocks chan *common.Block

    lock        *sync.RWMutex
}

func NewBlockChain() *BlockChain {

    return &BlockChain{
        candidates:     make([]common.TxPublish, 0),
        files:          make(map[string]*common.File),
        blocks:         make(map[string]*common.Block),
        isNew:          true,
        lock:           &sync.RWMutex{},
    }
}

func (gossiper *Gossiper) waitForNewBlocks() {

    go func() {

        for {

            // Wait for blocks
            block := <- gossiper.BlockChain.minedBlocks

            publish := &common.BlockPublish{
                Block:    *block,
                HopLimit: common.BlockHopLimit,
            }

            gossiper.broadcastToNeighbors(publish.Packed())
        }
    }()
}

// Atomically test and append transaction
func (bc *BlockChain) tryAddTransaction(candidate *common.TxPublish) bool {

    bc.lock.Lock()
    defer bc.lock.Unlock()

    _, found := bc.files[candidate.File.Name]

    if found {
        return false
    }

    for _, candidate := range bc.candidates {
        if candidate.File.Name == candidate.File.Name {
            return false
        }
    }

    bc.candidates = append(bc.candidates, *candidate)

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

    _, found := bc.blocks[hex.EncodeToString(hash[:])]

    if found {
        return false
    }

    if bytes.Compare(candidate.PrevHash[:], bc.previous[:]) != 0 && !bc.isNew {
        return false
    }

    bc.blocks[hex.EncodeToString(hash[:])] = candidate
    bc.previous = hash
    bc.isNew = true

    return true
}

func (bc *BlockChain) startMining() {

    go func() {

        for {

            var nonce [32]byte

            _, err := rand.Read(nonce[:])

            if err != nil {
                continue
            }

            bc.lock.RLock()

            candidate := &common.Block {
                bc.previous,
                nonce,
                bc.candidates,
            }

            bc.lock.RUnlock()

            hash := candidate.Hash()

            if hash[0] == 0 && hash[1] == 0 {

                if bc.tryAddBlock(candidate) {

                    bc.minedBlocks <- candidate
                }
            }
        }
    }()
}