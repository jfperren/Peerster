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
const DefaultSearchBudget = 2
const MaxSearchBudget = 32
const SearchKeywordSeparator = ","
const SearchRequestTimeThreshold = int64(500 * time.Millisecond)
const SearchRequestBudgetIncreaseDT = 1 * time.Second
const UnverifiableMessageRetryDT = 3 * time.Second
const SearchNoBudget = uint64(0)
const SearchRequestResultsThreshold = 2
const SearchTimeout = 1 * time.Second
const TransactionHopLimit = 10
const BlockHopLimit = 20
const FileNameSeparator = ","
const InitialMiningSleepTime = 5 * time.Second
const MiningSleepTimeFactor = 2
const MiningDifficulty = 16
const CryptoKeySize = 4096 // bits
const OnionBufferSize = 2048 // Additional space on top of maximum message size
const OnionPayloadSize = FileChunkSize + OnionBufferSize
const OnionSubHeaderPaddingSize = 256
const OnionSubHeaderSize = 512
const OnionSubHeaderCount = 8
const OnionHeaderSize = OnionSubHeaderSize * OnionSubHeaderCount
const OnionSize = OnionPayloadSize + OnionHeaderSize
const CTRKeySize = 32
const SignOnly = 1
const CypherIfPossible = 2
const MixerNodeBufferSize = 8
const MixerRandomTimeSleepRange = 1000
const NoNextHop = ""
