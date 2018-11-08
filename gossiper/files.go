package gossiper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jfperren/Peerster/common"
	"os"
)

type MetaFile struct {
	name  string
	size  int
	hash  []byte
	data  []byte
}

type Chunk struct {
	hash  []byte
	data  []byte
}

func NewMetaFile(name string, data []byte) *MetaFile {
	hash := sha256.Sum256(data)
	return &MetaFile{name, 0, hash[:] , data}
}

func NewChunk(data []byte) *Chunk {
	hash := sha256.Sum256(data)
	return &Chunk{hash[:], data}
}

type FileSystem struct {
	chunks map[string]*Chunk
	metaFiles map[string]*MetaFile
}

func (fs FileSystem) getMetaFile(hash []byte) (*MetaFile, bool) {
	v, found := fs.metaFiles[hex.EncodeToString(hash)]
	return v, found
}

func (fs FileSystem) getChunk(hash []byte) (*Chunk, bool) {
	v, found := fs.chunks[hex.EncodeToString(hash)]
	return v, found
}

func (metaFile *MetaFile) countOfChunks() int {
	return len(metaFile.data) / sha256.Size
}

func (metaFile *MetaFile) hashAt(index int) []byte {
	return metaFile.data[index * common.FileChunkSize: (index + 1) * common.FileChunkSize]
}

func (chunk *Chunk) size() int {
	return len(chunk.data)
}

func (chunk *Chunk) isLastIn(metaFile *MetaFile) bool {
	lastHash := metaFile.hashAt(metaFile.countOfChunks())
	return bytes.Compare(lastHash, chunk.hash) == 0
}

func (fs FileSystem) storeMetaFile(metaFile *MetaFile) bool {

	key := hex.EncodeToString(metaFile.hash)

	_, found := fs.metaFiles[key]

	if found {
		return false
	}

	fs.metaFiles[key] = metaFile

	go metaFile.save()

	return true
}

func (fs FileSystem) storeChunk(chunk *Chunk) bool {

	key := hex.EncodeToString(chunk.hash)

	metaFile, found := fs.metaFiles[key]

	if found {
		return false
	}

	metaFile.size += chunk.size()
	fs.chunks[key] = chunk

	go chunk.save()

	if chunk.isLastIn(metaFile) {
		fs.reconstructFile(metaFile.hash)
	}

	return true
}


func (fs FileSystem) saveDataReply(name string, reply *common.DataReply) bool {

	metaFile, found := fs.getMetaFile(reply.HashValue)

	if !found {

		metaFile = NewMetaFile(name, reply.Data)
		return fs.storeMetaFile(metaFile)

	} else {

		chunk := NewChunk(reply.Data)
		return fs.storeChunk(chunk)
	}
}

func (fs FileSystem) reconstructFile(metaHash []byte) {

	data := make([]byte, 0)

	metaFile, found := fs.getMetaFile(metaHash)

	if !found {
		panic("File not found")
	}

	for i := 0; i < metaFile.countOfChunks(); i++ {

		hash := metaFile.hashAt(i)
		chunk, found := fs.getChunk(hash)

		if !found {
			panic("Chunk not found")
		}

		data = append(data, chunk.data...)
	}

	filePath := common.SharedFilesDir + metaFile.name
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { fmt.Print(err) }
	file.Write(data)

}

func (fs FileSystem) downloadStatus(metaHash []byte) ([]byte, bool) {

	metaFile, found := fs.getMetaFile(metaHash)

	if !found {
		return metaHash, false
	}

	for i := 0; i < metaFile.countOfChunks(); i++ {

		hash := metaFile.hashAt(i)

		_, found := fs.getChunk(hash)

		if !found {
			return hash, false
		}
	}

	return make([]byte, 0), true
}

func (chunk *Chunk) save() {

	chunkPath := common.SharedFilesDir + hex.EncodeToString(chunk.hash) + common.ChunkFileSuffix
	chunkFile, err := os.OpenFile(chunkPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { fmt.Print(err) }
	chunkFile.Write(chunk.data)

}

func (meta *MetaFile) save() {

	metaPath := common.SharedFilesDir + meta.name + common.MetaFileSuffix
	metaFile, err := os.OpenFile(metaPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { fmt.Print(err) }
	metaFile.Write(meta.data)

}

// Prepares a file so as to make it available to send on the network. Chunk + computes hash
func (fs FileSystem) ScanFile(fileName string) {

	// Open file for reading
	filePath := common.SharedFilesDir + fileName
	file, err := os.Open(filePath)
	if err != nil { fmt.Print(err) }

	buff := make([]byte, common.FileChunkSize)
	size := 0
	hashes := make([]byte, 0)

	for i := 0; true; i++ {

		count, err := file.Read(buff);
		if err != nil { fmt.Print(err) }

		// Compute hash of chunk
		hash := sha256.Sum256(buff)

		// Append hash and increase length
		hashes = append(hashes, hash[:]...)
		size += count

		// Write chunk into separate file
		chunk := &Chunk{hash[:], buff}
		chunk.save()

		common.DebugScanChunk(i, hash[:])

		if count < common.FileChunkSize {
			break
		}
	}

	// Calculate hash of hashes
	metaHash := sha256.Sum256(hashes)

	// Create and open meta file
	metaFile := &MetaFile{fileName, size,metaHash[:],hashes}
	metaFile.save()

	common.DebugScanFile(fileName, size, metaHash[:])

}