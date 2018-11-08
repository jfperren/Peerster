package gossiper

import (
	"crypto/sha256"
	"fmt"
	"github.com/jfperren/Peerster/common"
	"os"
	"strconv"
)

// Prepares a file so as to make it available to send on the network. Chunk + computes hash
func ScanFile(fileName string) string {

	// Open file for reading
	filePath := common.SharedFilesDir + fileName
	file, err := os.Open(filePath)
	if err != nil { fmt.Print(err) }


	buff := make([]byte, common.FileChunkSize)
	length := 0
	hashes := make([]byte, 0)

	for i := 0; true; i++ {

		count, err := file.Read(buff);
		if err != nil { fmt.Print(err) }

		// Compute hash of chunk
		hash := sha256.Sum256(buff)

		// Append hash and increase length
		hashes = append(hashes, hash[:]...)
		length += count

		// Write chunk into separate file
		chunkPath := common.SharedFilesDir + fileName + common.ChunkFileSuffix + strconv.Itoa(i)
		chunkFile, err := os.OpenFile(chunkPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil { fmt.Print(err) }
		chunkFile.Write(buff)

		common.DebugScanChunk(i, hash[:])

		if count < common.FileChunkSize {
			break
		}
	}

	// Calculate hash of hashes
	metaHash := sha256.Sum256(hashes)

	// Create and open meta file
	metaPath := common.SharedFilesDir + fileName + common.MetaFileSuffix
	metaFile, err := os.OpenFile(metaPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { fmt.Print(err) }

	common.DebugScanFile(fileName, length, metaHash[:])

	// Write list of hashes in meta file
	metaFile.Write(hashes)

	return "Hello"
}