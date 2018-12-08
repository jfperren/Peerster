package gossiper

import (
    "bytes"
    "crypto/rand"
    "github.com/jfperren/Peerster/common"
    "sync"
    "time"
)


//
//  DATA STRUCTURES
//

// Internal representation of the current state of the block chain.
type BlockChain struct {

    Pending     []common.TxPublish          // Current transactions to be included in next block
    Files       map[string]*common.File     // Current state of the chain: mapping of name to file

    Blocks      map[[32]byte]*common.Block  // All chain blocks, mapped by hash
    Length      map[[32]byte]int            // Length of chain at each block

    IsNew       bool                        // If true, we accept any block from other peers (only true at init)
    MinedBlocks chan *common.Block          // Channel that publishes found blocks to be broadcasted
    MiningTime  int64                       // Timestamp of when the gossiper start mining a new block

    Latest      [32]byte                    // Current hash on the longest chain

    lock        *sync.RWMutex               // Mutex to synchronize access to the chain
}

//
//  CONSTRUCTORS
//


func NewBlockChain() *BlockChain {

    return &BlockChain{
        Pending:     make([]common.TxPublish, 0),
        Files:       make(map[string]*common.File),
        Blocks:      make(map[[32]byte]*common.Block),
        Length:      make(map[[32]byte]int),
        IsNew:       true,
        MinedBlocks: make(chan *common.Block, 2),
        MiningTime:  0,
        lock:        &sync.RWMutex{},
    }
}

func NewTransaction(metaFile *MetaFile) *common.TxPublish {
    return &common.TxPublish{
        File: common.File{
            Name: metaFile.Name,
            Size: int64(metaFile.Size),
            MetafileHash: metaFile.Hash,
        },
        HopLimit: common.TransactionHopLimit,
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

        common.DebugBroadcastBlock(publish.Block.Hash())

        gossiper.broadcastToNeighbors(publish.Packed())
    }
}

//
//  UPDATE FUNCTIONS
//

// Atomically test and append transaction
func (bc *BlockChain) TryAddFile(candidate *common.TxPublish) bool {

    bc.lock.Lock()
    defer bc.lock.Unlock()

    _, found := bc.Files[candidate.File.Name]

    if found {
        common.DebugIgnoreTransactionAlreadyInChain(candidate)
        return false
    }

    for _, otherCandidates := range bc.Pending {
        if candidate.File.Name == otherCandidates.File.Name {
            common.DebugIgnoreTransactionAlreadyCandidate(candidate)
            return false
        }
    }

    bc.Pending = append(bc.Pending, *candidate)
    common.DebugAddCandidateTransaction(candidate)

    return true
}

// Atomically test and append block
func (bc *BlockChain) TryAddBlock(candidate *common.Block) bool {

    bc.lock.Lock()
    defer bc.lock.Unlock()

    hash := candidate.Hash()

    if !(hash[0] == 0 && hash[1] == 0) {
        common.DebugIgnoreBlockIsNotValid(candidate)
        return false
    }

    _, found := bc.Blocks[hash]

    if found {
        common.DebugIgnoreBlockAlreadyPresent(candidate)
        return false
    }

    if !bc.IsConsistent(candidate) {
        common.DebugIgnoreBlockInconsistent(candidate)
        return false
    }

    // All good, store block

    bc.Blocks[hash] = candidate
    bc.Length[hash] = bc.Length[candidate.PrevHash] + 1


    if bc.IsNew || bytes.Compare(candidate.PrevHash[:], bc.Latest[:]) == 0 {

        // Append on the longest chain

        bc.Latest = hash

        for _, transaction := range candidate.Transactions {
            bc.Files[transaction.File.Name] = &transaction.File
        }

        bc.updatePendingTransactions()

        common.LogChain(bc.allBlocks())

    } else if bc.Length[hash] > bc.Length[bc.Latest] {

        // Need to Rollback

        latest, found := bc.Blocks[bc.Latest]

        if !found {
            bc.IsNew = false
            return false
        }

        _, _, currentChain, newChain := bc.FirstCommonAncestor(latest, candidate)

        bc.rollback(currentChain)
        bc.fastForward(newChain)
        bc.updatePendingTransactions()

        common.LogForkLongerRewind(currentChain)
        common.DebugChainLength(bc.Length[hash])

    } else {

        // We already stored it, just log
        common.LogShorterFork(candidate)
        common.DebugChainLength(bc.Length[hash])
    }

    bc.IsNew = false

    return true

}

func (bc *BlockChain) updatePendingTransactions() {

    newPending := make([]common.TxPublish, 0)

    for _, transaction := range bc.Pending {
        _, found := bc.Files[transaction.File.Name]

        if !found {
            newPending = append(newPending, transaction)
        }
    }

    bc.Pending = newPending
}

func (bc *BlockChain) rollback(chain []*common.Block) {

    for _, block := range chain {
        for _, transaction := range block.Transactions {
            delete(bc.Files, transaction.File.Name)
        }
    }
}

func (bc *BlockChain) fastForward(chain []*common.Block) {

    for _, block := range chain {
        for _, transaction := range block.Transactions {
            bc.Files[transaction.File.Name] = &transaction.File
        }
    }

    bc.Latest = chain[0].Hash()
}

func (bc *BlockChain) mine() {

    bc.MiningTime = time.Now().UnixNano()

    for {

        var nonce [32]byte

        _, err := rand.Read(nonce[:])

        if err != nil {
            continue
        }

        bc.lock.RLock()

        candidate := &common.Block {
            PrevHash: bc.Latest,
            Nonce: nonce,
            Transactions: bc.Pending,
        }

        bc.lock.RUnlock()

        hash := candidate.Hash()

        if hash[0] == 0 && hash[1] == 0 {

            common.LogFoundBlock(hash)

            if bc.TryAddBlock(candidate) {

                var sleepTime time.Duration

                if bytes.Equal(candidate.PrevHash[:], []byte{}) {
                    sleepTime = common.InitialMiningSleepTime
                } else {
                    diff := time.Duration(common.MiningSleepTimeFactor * (time.Now().UnixNano() - bc.MiningTime))
                    sleepTime = time.Duration(diff * time.Nanosecond)
                }

                common.DebugSleep(sleepTime)
                time.Sleep(sleepTime)

                bc.MinedBlocks <- candidate

                bc.MiningTime = time.Now().UnixNano()
            }
        }
    }
}

func (bc *BlockChain) allBlocks() []*common.Block {

    hash := bc.Latest
    allBlocks := make([]*common.Block, 0)

    for {
        block, found := bc.Blocks[hash]

        if !found {
            return allBlocks
        }

        allBlocks = append(allBlocks, block)
        hash = block.PrevHash
    }
}

func (bc *BlockChain) IsConsistent(newBlock *common.Block) bool {

    names := make(map[string]bool)

    for _, transaction := range newBlock.Transactions {

        _, found := names[transaction.File.Name]

        if found {
            return false
        }

        names[transaction.File.Name] = true
    }

    hash := newBlock.PrevHash

    for {
        block, found := bc.Blocks[hash]

        if !found {
            return true
        }

        for _, transaction := range block.Transactions {

            _, found := names[transaction.File.Name]

            if found {
                return false
            }

            names[transaction.File.Name] = true
        }

        hash = block.PrevHash
    }

    return true
}

func (bc *BlockChain) FirstCommonAncestor(current, new *common.Block) (*common.Block, bool, []*common.Block, []*common.Block) {

    // Initialize map with parents seen
    visited := make(map[[32]byte]*common.Block)
    var zero [32]byte

    currentHash := current.Hash()
    newHash := new.Hash()

    var ancestor *common.Block
    hasCommonAncestor := false

    // Will use these to store path to the common ancestor
    currentChain := make([]*common.Block, 0)
    newChain := make([]*common.Block, 0)

    hash := currentHash

    for {

        if bytes.Equal(hash[:], zero[:]){
            break
        }

        block, found := bc.Blocks[hash]

        if !found {
            break
        }

        // Climb up the chain
        visited[hash] = block
        hash = block.PrevHash
    }

    hash = newHash

    for {

        block, found := visited[hash]

        if found {
            // Found common ancestor
            ancestor = block
            hasCommonAncestor = true
            break
        }

        block, found = bc.Blocks[hash]

        if !found {
            // There is no common ancestor. Chains have different genesis block
            ancestor = nil
            break
        }

        // Add block to the list
        newChain = append(newChain, block)

        // Climb up the chain
        visited[hash] = block
        hash = block.PrevHash
    }

    hash = currentHash

    for {

        block, found := bc.Blocks[hash]

        if !found {
            // Finished going up the chain, there was no common ancestor
            break
        }

        if ancestor != nil && ancestor.Hash() == hash {
            // We are at the ancestor
            break
        }

        currentChain = append(currentChain, block)
        hash = block.PrevHash
    }

    return ancestor, hasCommonAncestor, currentChain, newChain
}

