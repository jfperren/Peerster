package gossiper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jfperren/Peerster/common"
	"os"
	"sync"
	"time"
)

// File System handles the low-level complexity of chunking (scanning) existing files, storing file hashes,
// deciding which hashes are missing in a file, etc...
type FileSystem struct {
	sharedPath   string
	downloadPath string
	chunks       map[string]*Chunk
	metaFiles    map[string]*MetaFile
	lock	 	 *sync.RWMutex
}

type MetaFile struct {
	Name 	string
	Size 	int
	Hash 	[]byte
	Data 	[]byte
}

type Chunk struct {
	hash []byte
	data []byte
}

const FileNotFound = 1
const FileTooBig = 2

// Errors thrown by the File System type
type FileSystemError struct {
	filename string
	flag     int
}

func (e *FileSystemError) Error() string {
	switch e.flag {
	case FileNotFound:
		return "File not found: " + e.filename
	case FileTooBig:
		return "File is too big: " + e.filename + " (max. 2Mb)"
	default:
		return "Unexpected error"
	}
}

// Create a new File System
func NewFileSystem(sharedPath, downloadPath string) *FileSystem {

	return &FileSystem{
		sharedPath: sharedPath,
		downloadPath: downloadPath,
		chunks: make(map[string]*Chunk),
		metaFiles: make(map[string]*MetaFile),
		lock: &sync.RWMutex{},
	}
}

// Create a new metafile
func NewMetaFile(name string, data []byte) *MetaFile {
	hash := sha256.Sum256(data)
	return &MetaFile{name, 0, hash[:], data}
}

// Create a new chunk
func NewChunk(data []byte) *Chunk {
	hash := sha256.Sum256(data)
	return &Chunk{hash[:], data}
}

// Get meta file related to a hash
func (fs *FileSystem) getMetaFile(hash []byte) (*MetaFile, bool) {
	v, found := fs.metaFiles[hex.EncodeToString(hash)]
	return v, found
}

// Get chunk related to a hash
func (fs *FileSystem) getChunk(hash []byte) (*Chunk, bool) {
	v, found := fs.chunks[hex.EncodeToString(hash)]
	return v, found
}

// Number of chunks related to a metafile
func (metaFile *MetaFile) countOfChunks() int {
	return len(metaFile.Data) / sha256.Size
}

// Returns the i-th hash in the metafile
func (metaFile *MetaFile) hashAt(index int) []byte {
	return metaFile.Data[index*sha256.Size : (index+1)*sha256.Size]
}

// Size in bytes of a chunk
func (chunk *Chunk) size() int {
	return len(chunk.data)
}

// Return true if a chunk is the last expected one in a file
func (chunk *Chunk) isLastIn(metaFile *MetaFile) bool {
	lastHash := metaFile.hashAt(metaFile.countOfChunks() - 1)
	return bytes.Compare(lastHash, chunk.hash) == 0
}

// Store a meta file into the file system
func (fs *FileSystem) storeMetaFile(metaFile *MetaFile) {

	key := hex.EncodeToString(metaFile.Hash)

	fs.metaFiles[key] = metaFile

}

// Store a chunk into the file system
func (fs *FileSystem) storeChunk(chunk Chunk)  {

	key := hex.EncodeToString(chunk.hash)

	fs.chunks[key] = &chunk

}

// Take a data reply and store its information in the file system.
func (fs *FileSystem) processDataReply(name string, metaHash []byte, reply *common.DataReply) {

	metaFile, found := fs.getMetaFile(metaHash)

	if !found {

		metaFile = NewMetaFile(name, reply.Data)
		fs.storeMetaFile(metaFile)

	} else {

		chunk := NewChunk(reply.Data)

		fs.storeChunk(*chunk)

		metaFile.Size += chunk.size()

		if chunk.isLastIn(metaFile) {
			fs.reconstructFile(metaFile.Hash)
		}
	}
}

// Reconstruct a file using all the chunks downloaded.
func (fs *FileSystem) reconstructFile(metaHash []byte) {

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

	common.LogReconstructed(metaFile.Name)

	filePath := fs.downloadPath + metaFile.Name
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Print(err)
	}
	file.Write(data)

}

// Check what is the next missing hash related to a file.
//
// - First return value is the next hash to be downloaded (or [] if completed)
// - Second return value is the next chunk Id (or -1 if metafile is still needed, -2 if completed)
// - Third return value returns true if the download is completed.
//
func (fs *FileSystem) downloadStatus(metaHash []byte) ([]byte, int, bool) {

	metaFile, found := fs.getMetaFile(metaHash)

	if !found {
		return metaHash, common.MetaHashChunkId, false
	}

	for i := 0; i < metaFile.countOfChunks(); i++ {

		hash := metaFile.hashAt(i)

		_, found := fs.getChunk(hash)

		if !found {
			return hash, i + 1, false
		}
	}

	return make([]byte, 0), common.NoChunkId, true
}

// Prepares a file so as to make it available to send on the network. Chunk + computes hash
func (fs *FileSystem) ScanFile(fileName string) (*MetaFile, error) {

	// Open file for reading
	filePath := fs.sharedPath + fileName
	file, err := os.Open(filePath)

	if err != nil {
		common.DebugFileNotFound(fileName)
		return nil, &FileSystemError{fileName, FileNotFound}
	}

	size := 0
	hashes := make([]byte, 0)
	chunks := make([]Chunk, 0)

	for i := 0; true; i++ {

		buff := make([]byte, common.FileChunkSize)
		count, err := file.Read(buff)
		if err != nil {
			fmt.Print(err)
		}

		// Compute hash of chunk
		hash := sha256.Sum256(buff)

		// Append hash and increase length
		hashes = append(hashes, hash[:]...)
		size += count

		chunk := Chunk{hash[:], buff}
		chunks = append(chunks, chunk)

		if len(hashes) > common.FileChunkSize {
			common.DebugFileTooBig(fileName)
			return nil, &FileSystemError{fileName, FileTooBig}
		}

		if count < common.FileChunkSize {
			break
		}
	}

	// Calculate hash of hashes
	metaHash := sha256.Sum256(hashes)

	// Create and open meta file
	metaFile := &MetaFile{fileName, size, metaHash[:], hashes}
	fs.storeMetaFile(metaFile)

	// Store all chunks
	for i, chunk := range chunks {
		common.DebugScanChunk(i, chunk.hash)
		fs.storeChunk(chunk)
	}

	common.DebugScanFile(fileName, size, metaHash[:])

	return metaFile, nil
}

func (fs *FileSystem) getFileWithName(filename string) *MetaFile {

	for _, metaFile := range fs.metaFiles {
		if metaFile.Name == filename {
			return metaFile
		}
	}

	return nil
}

//
//  GOSSIPER METHODS
//

// Start a new download process. As long as there are missing chunks related to the metahash provided, it will
// continue to download the remaining ones. If there is a timeout or the response does not match the data requested,
// it will restart evert 5 seconds, up to a total of 10 tries before stopping.
//
//  - name: Name of the file as it will appear in the file system later on
//  - metaHash: Hash of the requested file
//  - peer: Name of the peer from which we want to download the file
//  - counter: Set to 0 for new downloads, it will increase up to 10 until timing out.
//
func (gossiper *Gossiper) StartDownload(name string, metaHash []byte, peer string, counter int) {

	if counter > common.MaxDownloadRequests {
		// Probably print some stuff
		common.DebugDownloadTimeout(name, metaHash, peer)
		return
	}

	nextHash, chunkId, completed := gossiper.FileSystem.downloadStatus(metaHash)

	if completed {
		common.DebugDownloadCompleted(name, metaHash, peer)
		return
	}

	if peer == "" {

		fileMap, found := gossiper.SearchEngine.FileMap(metaHash)

		if !found {
			common.DebugDownloadUnknownFile(metaHash)
			return
		}

		if chunkId == common.MetaHashChunkId {
			peer, found = fileMap.peerForMetafile(counter)
		} else {
			peer, found = fileMap.peerForChunk(uint64(chunkId), counter)
		}

		if !found {
			common.DebugNoKnownOwnerForFile(metaHash)
			return
		}
	}

	common.DebugStartDownload(name, nextHash, peer)

	go func() {

		ticker := time.NewTicker(common.DownloadTimeout)
		defer ticker.Stop()

		select {
		case packet := <-gossiper.Dispatcher.dataReplies(nextHash):

			reply := packet.DataReply

			if reply == nil {
				// Error, we expect only a data reply from this
				return
			}

			if !reply.VerifyHash(nextHash) {
				// Unexpected hash or incorrect data, retry
				common.DebugCorruptedDataReply(nextHash, reply)
				go gossiper.StartDownload(name, metaHash, peer, counter+1)
				return
			}

			gossiper.FileSystem.processDataReply(name, metaHash, reply)

			// At this point, the download is successful, so we can log it

			if chunkId == common.MetaHashChunkId {
				common.LogDownloadingMetafile(name, packet.DataReply.Origin)
			} else {
				common.LogDownloadingChunk(name, chunkId, packet.DataReply.Origin)
			}

			// Then we start downloading the next chunk
			go gossiper.StartDownload(name, metaHash, peer, 0)

		case <-ticker.C: // Timeout
			go gossiper.StartDownload(name, metaHash, peer, counter+1)
		}

		gossiper.Dispatcher.stopWaitingOnDataReply(nextHash)
	}()

	request := gossiper.GenerateDataRequest(peer, nextHash)
	gossiper.sendToNode(request.Packed(), request.Destination, nil)

}