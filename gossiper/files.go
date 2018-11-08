package gossiper

import (
	"crypto/sha256"
	"fmt"
	"github.com/jfperren/Peerster/common"
	"os"
)

// Prepares a file so as to make it available to send on the network. Chunk + computes hash
func ScanFile(fileName string) {

	filePath := common.SharedFilesDir + fileName
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Print(err)
	}

	metaPath := common.SharedFilesDir + fileName + common.MetaFileSuffix
	meta, err := os.OpenFile(metaPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	buff := make([]byte, common.FileChunkSize)

	for i := 0; true; i++ {

		count, err := file.Read(buff);

		if err != nil {
			fmt.Print(err)
		}

		hash := sha256.Sum256(buff)

		common.DebugScanChunk(i, hash)

		meta.Write(hash[:])

		if count < common.FileChunkSize {
			break
		}
	}
}