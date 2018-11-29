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
    downloadNoHash = iota
    downloadNoName
    downloadNoDest
    downloadInvalidHash
    searchNoKeywords
)

func (e *CommandError) Error() string {
    switch e.flag {
    case downloadNoHash:        return "Cannot request a file without giving a hash"
    case downloadNoName:        return "Cannot request a file without giving a name"
    case downloadNoDest:        return "Cannot request a file without specifying the destination node"
    case downloadInvalidHash:   return "Error decoding hash specified in 'request'"
    case searchNoKeywords:      return "Cannot search without providing keywords"
    default:                    return "Unexpected error"
    }
}

// --
// -- CONSTRUCTORS
// --

func NewMessageCommand(content *string) (*Command, *CommandError) {

    privateMessageCommand := &MessageCommand{*content}
    return &Command{privateMessageCommand, nil, nil, nil, nil}, nil
}

func NewPrivateMessageCommand(content, destination *string) (*Command, *CommandError) {

    privateMessageCommand := &PrivateMessageCommand{*content, *destination}
    return &Command{nil, privateMessageCommand, nil, nil, nil}, nil
}

func NewUploadCommand(file *string) (*Command, *CommandError) {

    uploadCommand := &UploadCommand{*file}
    return &Command{nil, nil, uploadCommand, nil, nil}, nil
}

func NewDownloadCommand(request, file, dest *string) (*Command, *CommandError) {

    if *request == "" {
        return nil, &CommandError{downloadNoHash}
    }

    if *file == "" {
        return nil, &CommandError{downloadNoName}
    }

    if *dest == "" {
        return nil, &CommandError{downloadNoDest}
    }

    hash, err := hex.DecodeString(*request)

    if err != nil {
        return nil, &CommandError{downloadInvalidHash}
    }

    downloadCommand := &DownloadCommand{*file, *dest, hash}
    return &Command{nil, nil, nil, downloadCommand, nil}, nil
}

func NewSearchCommand(query *string, budget uint64) (*Command, *CommandError) {

    if *query == "" {
        return nil, &CommandError{searchNoKeywords}
    }

    keywords := strings.Split(*query, SearchKeywordSeparator)
    var searchCommand *SearchCommand

    searchCommand = &SearchCommand{budget, keywords}
    return &Command{nil, nil, nil, nil, searchCommand}, nil
}

// --
// -- CONVENIENCE
// --

func (command *Command) IsValid() bool {
    return boolCount(command.Message != nil)+boolCount(command.PrivateMessage != nil)+
        boolCount(command.Upload != nil)+boolCount(command.Download != nil)+
        boolCount(command.Search != nil) == 1
}
