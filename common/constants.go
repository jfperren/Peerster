package common

import "time"

const StatusBufferSize = 5
const StatusTimeout = 1 * time.Second
const DonwloadTimeout = 5 * time.Second
const AntiEntropyDT = 1 * time.Second
const InitialId = uint32(1)
const DontSendRouteRumor = time.Duration(0)
const InitialHopLimit = 10
const FileChunkSize = 8192
const SocketBufferSize = 2 * FileChunkSize
const SharedFilesDir = "_SharedFiles/"
const DownloadDir = "_Downloads/"
const MetaFileSuffix = ".meta"
const ChunkFileSuffix = ".chunk"
const NewDownload = 0
const MaxDownloadRequests = 10
const MetaHashChunkId = -1
const NoChunkId = -2
const ErrorFileNotFound = 1
const ErrorFileTooBig = 2