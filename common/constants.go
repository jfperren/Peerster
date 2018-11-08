package common

import "time"

const StatusBufferSize = 5
const StatusTimeout = 1 * time.Second
const AntiEntropyDT = 1 * time.Second
const InitialId = uint32(1)
const DontSendRouteRumor = time.Duration(0)
const InitialHopLimit = 10
const FileChunkSize = 8196
const SharedFilesDir = "_SharedFiles/"
const MetaFileSuffix = ".meta"