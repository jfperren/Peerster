package tests

import (
    "crypto/rand"
    "github.com/jfperren/Peerster/common"
    "github.com/jfperren/Peerster/gossiper"
    "testing"
)

func TestGenesisBlock(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock(make([]common.TxPublish, 0))

    ok := chain.TryAddBlock(b0)

    if !ok {
        t.Errorf("Chain should not reject a valid block")
    }

    if chain.Latest != b0.Hash() {
        t.Errorf("Chain did not update its status after receiving first block")
    }

    if chain.Length[b0.Hash()] != 1 {
        t.Errorf("Chain should have depth of 1 after adding b0, instead has %v", chain.Length[b0.Hash()])
    }
}

func TestNoUpdateStateWhenOtherGenesisBlock(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock(make([]common.TxPublish, 0))
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1 := newValidBlock(b0.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    otherGenesis := newGenesisBlock(make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(otherGenesis)
    if !ok { t.Errorf("Could not add block otherGenesis to the chain") }

    if chain.Latest != b1.Hash() {
        t.Errorf("Chain should not update its state when adding a new genesis block")
    }
}

func TestNoUpdateStateWhenShorterChain(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0_0 := newGenesisBlock(make([]common.TxPublish, 0))
    ok := chain.TryAddBlock(b0_0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1_0 := newValidBlock(b0_0.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b1_0)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b0_1 := newGenesisBlock(make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b0_1)
    if !ok { t.Errorf("Could not add block otherGenesis to the chain") }

    b1_1 := newValidBlock(b0_1.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b1_1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    if chain.Latest != b1_0.Hash() {
        t.Errorf("Chain should not update its state when adding a new genesis block")
    }
}

func TestRejectInvalidBlock(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newRandomBlock(make([]common.TxPublish, 0))

    ok := chain.TryAddBlock(b0)

    if ok {
        t.Errorf("Chain should reject wrong blocks")
    }
}

func TestAncestorCommon(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock(make([]common.TxPublish, 0))
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1 := newValidBlock(b0.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b2_0 := newValidBlock(b1.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b2_0)
    if !ok { t.Errorf("Could not add block b2_0 to the chain") }

    b2_1 := newValidBlock(b1.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b2_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    b3_0 := newValidBlock(b2_0.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b3_0)
    if !ok { t.Errorf("Could not add block b3_0 to the chain") }

    ancestor, hasCommonAncestor, currentChain, newChain := chain.FirstCommonAncestor(b3_0, b2_1)

    if !hasCommonAncestor {
        t.Errorf("b3_0 and b2_1 should have a common ancestor")
    }

    if ancestor.Hash() != b1.Hash() {
        t.Errorf("Common ancestor should be b1, instead is %v", ancestor)
    }

    if len(newChain) != 1 {
        t.Errorf("Length of new chain should be 1, instead is %v", len(newChain))
    }

    if newChain[0].Hash() != b2_1.Hash() {
        t.Errorf("new chain should be [b2_1], instead is %v", newChain)
    }

    if len(currentChain) != 2 {
        t.Errorf("Length of current chain should be 2, instead is %v", len(currentChain))
    }

    if currentChain[0].Hash() != b3_0.Hash() || currentChain[1].Hash() != b2_0.Hash() {
        t.Errorf("new chain should be [b3_0, b2_0], instead is %v", currentChain)
    }
}

func TestAncestorSeparate(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0_0 := newGenesisBlock(make([]common.TxPublish, 0))
    ok := chain.TryAddBlock(b0_0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1_0 := newValidBlock(b0_0.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b1_0)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b2_0 := newValidBlock(b1_0.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b2_0)
    if !ok { t.Errorf("Could not add block b2_0 to the chain") }

    b0_1 := newGenesisBlock(make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b0_1)
    if !ok { t.Errorf("Could not add block b0_1 to the chain") }

    b1_1 := newValidBlock(b0_1.Hash(), make([]common.TxPublish, 0))
    ok = chain.TryAddBlock(b1_1)
    if !ok { t.Errorf("Could not add block b1_1 to the chain") }

    _, hasCommonAncestor, currentChain, newChain := chain.FirstCommonAncestor(b2_0, b1_1)

    if hasCommonAncestor {
        t.Errorf("b2_0 and b1_1 should not have a common ancestor")
    }

    if len(newChain) != 2 {
        t.Errorf("Length of new chain should be 2, instead is %v", len(newChain))
    }

    if newChain[0].Hash() != b1_1.Hash() || newChain[1].Hash() != b0_1.Hash()  {
        t.Errorf("new chain should be [b1_1, b0_1], instead is %v", newChain)
    }

    if len(currentChain) != 3 {
        t.Errorf("Length of current chain should be 4, instead is %v", len(currentChain))
    }

    if currentChain[0].Hash() != b2_0.Hash() || currentChain[1].Hash() != b1_0.Hash() || currentChain[2].Hash() != b0_0.Hash()  {
        t.Errorf("currentChain chain should be [b2_0, b1_0, b0_0], instead is %v", currentChain)
    }
}


func newGenesisBlock(transactions []common.TxPublish) *common.Block {

    var initialHash [32]byte

    return newValidBlock(initialHash, transactions)
}

func newRandomBlock(transactions []common.TxPublish) *common.Block {

    var nonce [32]byte

    rand.Read(nonce[:])

    return &common.Block {
        Nonce: nonce,
        Transactions: transactions,
    }
}

func newValidBlock(prevHash [32]byte, transactions []common.TxPublish) *common.Block {
    for {

        var nonce [32]byte

        _, err := rand.Read(nonce[:])

        if err != nil {
            continue
        }

        candidate := &common.Block {
            prevHash,
            nonce,
            transactions,
        }

        hash := candidate.Hash()

        if hash[0] == 0 && hash[1] == 0 {
            return candidate
        }
    }
}