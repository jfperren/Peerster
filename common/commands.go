package common

import (
    "encoding/hex"
    "strings"
)

// --
// -- DATA STRUCTURES
// --

type Command struct {
    Message         *MessageCommand
    PrivateMessage  *PrivateMessageCommand
    Upload          *UploadCommand
    Download        *DownloadCommand
    Search          *SearchCommand
}

type MessageCommand struct {
    Content     string
}

type PrivateMessageCommand struct {
    Content     string
    Destination string
}

type UploadCommand struct {
    FileName    string
}

type DownloadCommand struct {
    FileName    string
    Destination string
    Hash        []byte
}

type SearchCommand struct {
    Budget      uint64
    Keywords    []string
}

// --
// -- ERRORS
// --

type CommandError struct {
    flag     int
}

const (
    messageNoContent = iota

    privateMessageNoContent
    privateMessageNoDest

    uploadNoName

    downloadNoHash
    downloadNoName
    downloadInvalidHash

    searchNoKeywords
)

func (e *CommandError) Error() string {
    switch e.flag {

    case messageNoContent:          return "Cannot send a message without content"

    case privateMessageNoContent:   return "Cannot send a private message without content"
    case privateMessageNoDest:      return "Cannot send a private message without destination"

    case uploadNoName:              return "Cannot upload a file without a name"

    case downloadNoHash:            return "Cannot request a file without giving a hash"
    case downloadNoName:            return "Cannot request a file without giving a name"
    case downloadInvalidHash:       return "Error decoding hash specified in 'request'"

    case searchNoKeywords:          return "Cannot search without providing keywords"
    default:                        return "Unexpected error"
    }
}

// --
// -- CONSTRUCTORS
// --

func NewMessageCommand(content string) (*Command, *CommandError) {

    if content == "" {
        return nil, &CommandError{messageNoContent}
    }

    privateMessageCommand := &MessageCommand{content}
    return &Command{Message: privateMessageCommand}, nil
}

func NewPrivateMessageCommand(content, destination string) (*Command, *CommandError) {

    if content == "" {
        return nil, &CommandError{privateMessageNoContent}
    }

    if destination == "" {
        return nil, &CommandError{privateMessageNoDest}
    }

    privateMessageCommand := &PrivateMessageCommand{content, destination}
    return &Command{PrivateMessage: privateMessageCommand}, nil
}

func NewUploadCommand(file string) (*Command, *CommandError) {

    uploadCommand := &UploadCommand{file}
    return &Command{Upload: uploadCommand}, nil
}

func NewDownloadCommand(request, file, dest string) (*Command, *CommandError) {

    if request == "" {
        return nil, &CommandError{downloadNoHash}
    }

    if file == "" {
        return nil, &CommandError{downloadNoName}
    }

    hash, err := hex.DecodeString(request)

    if err != nil {
        return nil, &CommandError{downloadInvalidHash}
    }

    downloadCommand := &DownloadCommand{file, dest, hash}
    return &Command{Download: downloadCommand}, nil
}

func NewSearchCommand(query *string, budget uint64) (*Command, *CommandError) {

    if *query == "" {
        return nil, &CommandError{searchNoKeywords}
    }

    keywords := strings.Split(*query, SearchKeywordSeparator)
    var searchCommand *SearchCommand

    searchCommand = &SearchCommand{budget, keywords}
    return &Command{Search: searchCommand}, nil
}

// --
// -- CONVENIENCE
// --

func (command *Command) IsValid() bool {
    return boolCount(command.Message != nil)+boolCount(command.PrivateMessage != nil)+
        boolCount(command.Upload != nil)+boolCount(command.Download != nil)+
        boolCount(command.Search != nil) == 1
}
