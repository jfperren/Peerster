package common

import (
    "encoding/hex"
    "strings"
)

//
//  DATA STRUCTURES
//

// Aggregate of all other fields, should be used as top-level
// entity for internal communication with client.
type Command struct {
    Message         *MessageCommand
    PrivateMessage  *PrivateMessageCommand
    Upload          *UploadCommand
    Download        *DownloadCommand
    Search          *SearchCommand
}

// A command to send a message or rumor.
type MessageCommand struct {
    Content     string
}

// A command to send a private message to someone.
type PrivateMessageCommand struct {
    Content     string
    Destination string
}

// A command to upload a file
type UploadCommand struct {
    FileName    string
}

// A command to download a file
type DownloadCommand struct {
    FileName    string
    Destination string
    Hash        []byte
}

type SearchCommand struct {
    Budget      uint64
    Keywords    []string
}

//
//  ERRORS
//

type CommandError struct {
    flag     int
}

const (
    invalidCommand = iota

    messageNoContent

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

    case invalidCommand:            return "Command is not valid - Only one command at a time"

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

func InvalidCommandError() error {
    return &CommandError{invalidCommand}
}

//
//  CONSTRUCTORS
//

func NewMessageCommand(content string) (*Command, error) {

    if content == "" {
        return nil, &CommandError{messageNoContent}
    }

    privateMessageCommand := &MessageCommand{content}
    return &Command{Message: privateMessageCommand}, nil
}

func NewPrivateMessageCommand(content, destination string) (*Command, error) {

    if content == "" {
        return nil, &CommandError{privateMessageNoContent}
    }

    if destination == "" {
        return nil, &CommandError{privateMessageNoDest}
    }

    privateMessageCommand := &PrivateMessageCommand{content, destination}
    return &Command{PrivateMessage: privateMessageCommand}, nil
}

func NewUploadCommand(file string) (*Command, error) {

    if file == "" {
        return nil, &CommandError{uploadNoName}
    }

    uploadCommand := &UploadCommand{file}
    return &Command{Upload: uploadCommand}, nil
}

func NewDownloadCommand(request, file, dest string) (*Command, error) {

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

func NewSearchCommand(query string, budget uint64) (*Command, error) {

    if query == "" {
        return nil, &CommandError{searchNoKeywords}
    }

    keywords := strings.Split(query, SearchKeywordSeparator)
    var searchCommand *SearchCommand

    searchCommand = &SearchCommand{budget, keywords}
    return &Command{Search: searchCommand}, nil
}

//
//  SANITY CHECK
//

// Check if a given command is valid (i.e. only contains one non-nil field).
func (command *Command) IsValid() bool {
    return boolCount(command.Message != nil)+boolCount(command.PrivateMessage != nil)+
        boolCount(command.Upload != nil)+boolCount(command.Download != nil)+
        boolCount(command.Search != nil) == 1
}
