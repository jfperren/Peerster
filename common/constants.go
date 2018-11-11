package common

import "time"

const StatusBufferSize = 5
const StatusTimeout = 1 * time.Second
const DownloadTimeout = 5 * time.Second
const AntiEntropyDT = 1 * time.Second
const InitialId = uint32(1)
const NoRouteRumor = time.Duration(0)
const InitialHopLimit = 10
const FileChunkSize = 8192
const SocketBufferSize = 2 * FileChunkSize
const SharedFilesDir = "_SharedFiles/"
const DownloadDir = "_Downloads/"
const MaxDownloadRequests = 10
const MetaHashChunkId = -1
const NoChunkId = -2
