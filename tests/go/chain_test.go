package tests

import (
    "crypto/rand"
    "github.com/jfperren/Peerster/common"
    "github.com/jfperren/Peerster/gossiper"
    "testing"
)

func TestGenesisBlock(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock(make([]string, 0))

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

    b0 := newGenesisBlock(make([]string, 0))
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1 := newValidBlock(b0.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    otherGenesis := newGenesisBlock(make([]string, 0))
    ok = chain.TryAddBlock(otherGenesis)
    if !ok { t.Errorf("Could not add block otherGenesis to the chain") }

    if chain.Latest != b1.Hash() {
        t.Errorf("Chain should not update its state when adding a new genesis block")
    }
}

func TestNoUpdateStateWhenShorterChain(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0_0 := newGenesisBlock(make([]string, 0))
    ok := chain.TryAddBlock(b0_0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1_0 := newValidBlock(b0_0.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b1_0)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b0_1 := newGenesisBlock(make([]string, 0))
    ok = chain.TryAddBlock(b0_1)
    if !ok { t.Errorf("Could not add block otherGenesis to the chain") }

    b1_1 := newValidBlock(b0_1.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b1_1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    if chain.Latest != b1_0.Hash() {
        t.Errorf("Chain should not update its state when adding a new genesis block")
    }
}

func TestRejectInvalidBlock(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newRandomBlock(make([]string, 0))

    ok := chain.TryAddBlock(b0)

    if ok {
        t.Errorf("Chain should reject wrong blocks")
    }
}

func TestAncestorCommon(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock(make([]string, 0))
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1 := newValidBlock(b0.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b2_0 := newValidBlock(b1.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b2_0)
    if !ok { t.Errorf("Could not add block b2_0 to the chain") }

    b2_1 := newValidBlock(b1.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b2_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    b3_0 := newValidBlock(b2_0.Hash(), make([]string, 0))
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

    b0_0 := newGenesisBlock(make([]string, 0))
    ok := chain.TryAddBlock(b0_0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1_0 := newValidBlock(b0_0.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b1_0)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b2_0 := newValidBlock(b1_0.Hash(), make([]string, 0))
    ok = chain.TryAddBlock(b2_0)
    if !ok { t.Errorf("Could not add block b2_0 to the chain") }

    b0_1 := newGenesisBlock(make([]string, 0))
    ok = chain.TryAddBlock(b0_1)
    if !ok { t.Errorf("Could not add block b0_1 to the chain") }

    b1_1 := newValidBlock(b0_1.Hash(), make([]string, 0))
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

func TestRejectInconsistentBlocks(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock([]string{"hello.txt"})
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1 := newValidBlock(b0.Hash(), []string{"image.png","zip.zip"})
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b2_0 := newValidBlock(b1.Hash(), []string{"hello.txt","message.txt", "presentation.pdf"})
    ok = chain.TryAddBlock(b2_0)
    if ok { t.Errorf("Should not add block b2_0 to the chain") }

    b2_1 := newValidBlock(b1.Hash(), []string{"message.txt"})
    ok = chain.TryAddBlock(b2_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    b1_1 := newValidBlock(b0.Hash(), []string{"hello.txt"})
    ok = chain.TryAddBlock(b1_1)
    if ok { t.Errorf("Should not add block b1_1 to the chain") }

    if chain.Latest != b2_1.Hash() {
        t.Errorf("There was an error updating the chain")
    }
}

func TestFilesAreUpdatedCorrectly(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock([]string{"hello.txt"})
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    _, found := chain.Files["hello.txt"]

    if !found {
        t.Errorf("hello.txt was not added properly to the chain")
    }

    b1 := newValidBlock(b0.Hash(), []string{"image.png","zip.zip"})
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    _, found = chain.Files["image.png"]

    if !found {
        t.Errorf("image.png was not added properly to the chain")
    }

    _, found = chain.Files["zip.zip"]

    if !found {
        t.Errorf("zip.zip was not added properly to the chain")
    }

    b2_0 := newValidBlock(b1.Hash(), []string{"hello.txt","message.txt", "presentation.pdf"})
    ok = chain.TryAddBlock(b2_0)
    if ok { t.Errorf("Should not add block b2_0 to the chain") }

    _, found = chain.Files["message.txt"]

    if found {
        t.Errorf("message.txt should not be added to the chain")
    }

    _, found = chain.Files["presentation.pdf"]

    if found {
        t.Errorf("presentation.pdf should not be added to the chain")
    }

    b2_1 := newValidBlock(b1.Hash(), []string{"message.txt"})
    ok = chain.TryAddBlock(b2_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    _, found = chain.Files["message.txt"]

    if !found {
        t.Errorf("message.txt was not added properly to the chain")
    }

    b1_1 := newValidBlock(b0.Hash(), []string{"hello.txt"})
    ok = chain.TryAddBlock(b1_1)
    if ok { t.Errorf("Should not add block b1_1 to the chain") }
}

func TestRollbackSameChain(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0 := newGenesisBlock([]string{"hello.txt"})
    ok := chain.TryAddBlock(b0)
    if !ok { t.Errorf("Could not add block b0 to the chain") }

    b1 := newValidBlock(b0.Hash(), []string{"image.png","zip.zip"})
    ok = chain.TryAddBlock(b1)
    if !ok { t.Errorf("Could not add block b1 to the chain") }

    b2_0 := newValidBlock(b1.Hash(), []string{"message.txt, DSE.pdf"})
    ok = chain.TryAddBlock(b2_0)
    if !ok { t.Errorf("Could not add block b2_0 to the chain") }

    b2_1 := newValidBlock(b1.Hash(), []string{"presentation.pdf"})
    ok = chain.TryAddBlock(b2_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    if chain.Latest != b2_0.Hash() {
        t.Errorf("Should not rollback when chains are the same size")
    }

    _, found := chain.Files["presentation.pdf"]

    if found {
        t.Errorf("Should not update state with smaller chains")
    }

    b3_1 := newValidBlock(b2_1.Hash(), []string{"DSE.pdf"})
    ok = chain.TryAddBlock(b3_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    if chain.Latest != b3_1.Hash() {
        t.Errorf("Should rollback when find a longer chain")
    }

    _, found = chain.Files["message.txt"]

    if found {
        t.Errorf("Should re-write state when performing rollback")
    }

    _, found = chain.Files["presentation.pdf"]

    if !found {
        t.Errorf("Did not fast-forward correctly: missing presentation.pdf")
    }

    _, found = chain.Files["DSE.pdf"]

    if !found {
        t.Errorf("Did not fast-forward correctly: missing DSE.pdf")
    }
}

func TestRollbackDifferentChain(t *testing.T) {

    chain := gossiper.NewBlockChain()

    b0_0 := newGenesisBlock([]string{"hello.txt"})
    ok := chain.TryAddBlock(b0_0)
    if !ok { t.Errorf("Could not add block b0_0 to the chain") }

    b0_1 := newGenesisBlock([]string{"message.pdf", "DSE.pdf"})
    ok = chain.TryAddBlock(b0_1)
    if !ok { t.Errorf("Could not add block b0_1 to the chain") }

    if chain.Latest != b0_0.Hash() {
        t.Errorf("Should not rollback when chains are the same size")
    }

    _, found := chain.Files["message.pdf"]

    if found {
        t.Errorf("Should not update state with smaller chains")
    }

    _, found = chain.Files["DSE.pdf"]

    if found {
        t.Errorf("Should not update state with smaller chains")
    }

    b1_0 := newValidBlock(b0_0.Hash(), []string{"message.pdf"})
    ok = chain.TryAddBlock(b1_0)
    if !ok { t.Errorf("Could not add block b1_0 to the chain") }

    _, found = chain.Files["message.pdf"]

    if !found {
        t.Errorf("Did not update file properly: missing message.pdf")
    }

    b1_1 := newValidBlock(b0_1.Hash(), []string{})
    ok = chain.TryAddBlock(b1_1)
    if !ok { t.Errorf("Could not add block b1_1 to the chain") }

    if chain.Latest != b1_0.Hash() {
        t.Errorf("Should not rollback when chains are the same size")
    }

    b2_1 := newValidBlock(b1_1.Hash(), []string{"image.png"})
    ok = chain.TryAddBlock(b2_1)
    if !ok { t.Errorf("Could not add block b2_1 to the chain") }

    if chain.Latest != b2_1.Hash() {
        t.Errorf("Should rollback when find a longer chain")
    }

    _, found = chain.Files["hello.txt"]

    if found {
        t.Errorf("Should re-write state when performing rollback")
    }

    _, found = chain.Files["image.png"]

    if !found {
        t.Errorf("Did not update file properly: missing image.png")
    }

    _, found = chain.Files["message.pdf"]

    if !found {
        t.Errorf("Did not update file properly: missing message.pdf")
    }

    _, found = chain.Files["DSE.pdf"]

    if !found {
        t.Errorf("Did not update file properly: missing DSE.pdf")
    }
}


func newGenesisBlock(files []string) *common.Block {

    var initialHash [32]byte

    return newValidBlock(initialHash, files)
}

func newRandomBlock(files []string) *common.Block {

    var nonce [32]byte

    rand.Read(nonce[:])

    return &common.Block {
        Nonce: nonce,
        Transactions: newTransactions(files),
    }
}

func newValidBlock(prevHash [32]byte, files []string) *common.Block {
    for {

        var nonce [32]byte

        _, err := rand.Read(nonce[:])

        if err != nil {
            continue
        }

        candidate := &common.Block {
            PrevHash: prevHash,
            Nonce: nonce,
            Transactions: newTransactions(files),
        }

        hash := candidate.Hash()

        if hash[0] == 0 && hash[1] == 0 && hash[2] < common.MiningDifficulty {
            return candidate
        }
    }
}

func newTransaction(filename string) common.TxPublish {

    return common.TxPublish{
        File: common.File{
            Name: filename,
            Size: 10,
        },
        HopLimit: 10,
    }
}

func newTransactions(filenames []string) []common.TxPublish {

    transactions := make([]common.TxPublish, 0)

    for _, filename := range filenames {
        transactions = append(transactions, newTransaction(filename))
    }

    return transactions
}