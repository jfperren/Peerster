package common

import "time"

const StatusBufferSize = 5
const StatusTimeout = 1 * time.Second
const DownloadTimeout = 5 * time.Second
const AntiEntropyDT = 1 * time.Second
const InitialId = uint32(1)
const NoRouteRumor = time.Duration(0)
const InitialHopLimit = 10
const FileChunkSize = 3800//8192
const SocketBufferSize = /*2*/ 4 * FileChunkSize
const SharedFilesDir = "_SharedFiles/"
const DownloadDir = "_Downloads/"
const MaxDownloadRequests = 10
const MetaHashChunkId = -1
const NoChunkId = -2
const DefaultSearchBudget = 2
const MaxSearchBudget = 32
const SearchKeywordSeparator = ","
const SearchRequestTimeThreshold = int64(500 * time.Millisecond)
const SearchRequestBudgetIncreaseDT = 1 * time.Second
const SearchNoBudget = uint64(0)
const SearchRequestResultsThreshold = 2
const SearchTimeout = 1 * time.Second
const TransactionHopLimit = 10
const BlockHopLimit = 20
const FileNameSeparator = ","
const InitialMiningSleepTime = 5 * time.Second
const MiningSleepTimeFactor = 2
const CryptoKeySize = 4096 // bits
const OnionPayloadSize = 11104//12288  
const OnionHeaderSize = SocketBufferSize - OnionPayloadSize
const OnionSubHeaderCount = 16
const OnionSubHeaderSize = OnionHeaderSize / OnionSubHeaderCount
const SignOnly = 1
const CypherIfPossible = 2
