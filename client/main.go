package main

import (
	"flag"
	"github.com/dedis/protobuf"
	"github.com/jfperren/Peerster/common"
)

func main() {

	// Define Flags

	uiPort := flag.String("UIPort", "8080", "port for the UI client")
	message := flag.String("msg", "Test message", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message")
	file := flag.String("file", "", "file to be indexed by the gossiper, or filename of the requested file")
	request := flag.String("request", "", "request a chunk or metafile of this hash")
	keywords := flag.String("keywords", "", "comma-separated list of keywords for search")
	budget := flag.Uint64("budget", common.SearchNoBudget, "budget for file search (optional)")

	flag.Parse()

	// Packet to send
	var command *common.Command
	var commandError *common.CommandError

	switch {

	case *keywords != "":

		command, commandError = common.NewSearchCommand(keywords, *budget)

	case *request != "":

		command, commandError = common.NewDownloadCommand(request, file, dest)

	case *file != "":

		command, commandError = common.NewUploadCommand(file)

	case *dest != "":

		command, commandError = common.NewPrivateMessageCommand(message, dest)

	case *message != "":

		command, commandError = common.NewMessageCommand(message)

	default:

		panic("No message specified, unclear instructions.")
	}

	if commandError != nil {
		panic(commandError)
	}

	// Encodes message
	bytes, encodeError := protobuf.Encode(command)
	if encodeError != nil {
		panic(encodeError)
	}

	// Send
	socket := common.NewUDPSocket(":5050")
	socket.Send(bytes, ":"+*uiPort)
	socket.Unbind()

}
