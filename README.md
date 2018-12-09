# Peerster


## Building

Instead of manually running `go build` in both `/` and `client/` folders, the whole project can be built by typing
```
scripts/build.sh
```

*Note - Even though the assignment mentionned something about renaming the `Peerster` exec into `gossiper`, the files provided with the assignment handout did not. As a result, we will not do it here either and run `Peerster` directy. If the executable **must** be named `gossiper`, it is possible to uncomment the lines in `build.sh`*.

## Running

#### Via Command-Line

To run a gossiper node in CLI mode, use the following command:
```
./Peerster -gossipAddr=127.0.0.1:5002 -UIPort=8082 -name="Charlie" -peers=127.0.0.1:5000 [-rtimer 5] [-verbose] [-separatefs]
```

Here, `rtimer` is the number of seconds between route rumors, `verbose` allows to display additional information (it is useful for debugging but might clutter the log) and `separatefs` allows the node to use its own subfolder of the `_Download` and `_SharedFiles` folder (Note: the folder is created using the `name` attribute).


Alternatively, you can simply use the following command (also works with `bob` instead of `alice`):
```
scripts/run_alice.sh
```

When the gossiper is running, you can send messages to the node directly via the command line `client` executable. To do this, use

```
// 1. In order to send a gossip message
client/client -UIPort=8082 -msg="Your message here"

// 2. In order to send a direct message
client/client -UIPort=8082 -msg="Your message here" -dest="Bob"

// 3. In order to upload a file (file must be in the correct _SharedFiles subfolder)
client/client -UIPort=8082 -file="File.txt"

// 4. In order to download a file from someone else
client/client -UIPort=8082 -request=<hash of file> -file="File.txt" -dest="Bob"
```

#### Using the GUI

In order to interact with the gossiper via the GUI, you will need to run the `Peerster` executable with the `-server` mode. For instance,

```
./Peerster -gossipAddr=127.0.0.1:5000 -UIPort=8080 -name="Alice" -peers=127.0.0.1:5001 -server [-rtimer 1] [-separatefs] [-sharedfs]
```

This will start the web server and serve the GUI on `UIPort`. Therefore, simply connect to the GUI by accessing `localhost:8080` in your browser.

*Note - As the web server will be connecting to the `UIPort` directly, it is no longer possible to interact with the gossiper through the command line `client` executable.*

Alternatively, there are also two pre-written scripts to start two nodes that communicate with each other (and with Charlie from the `run.sh` script!). 

```
scripts/run_alice_server.sh
scripts/run_bob_server.sh
```

## Testing

The two basic testing scripts, as well as more advanced tests for routing, private messages and file sharing are available in the `tests/sh/` folder:

```
tests/sh/test_chain.sh
tests/sh/test_download.sh
tests/sh/test_files.sh
tests/sh/test_private.sh
tests/sh/test_download.sh
tests/sh/test_routing.sh
tests/sh/test_rumors.sh
tests/sh/test_search.sh
tests/sh/test_simple.sh
```

There are also some unit tests written in go that can be run while inside the `tests/go/` folder using `go test -v`. Finally, it is possible to easily run everything as one big test suite using `scripts/test.sh`.

## Notes about Implementation of HW3

Here are relevant details for whoever is reading / testing the code of Homework 3.

- The node contininuously mines, even when there is no transaction. The usual mining time on my laptop is about 0.1s so the chain grows very rapidly. I did not make any change to that in order to avoid failing automatic tests and/or breaking compatibility with other nodes.
- Because it was unclear in the assignment guidelines, I decided to keep blocks that have an unknown parent. The main reason is this ensures that a node joining the network late has still a chance to somewhat converge to the longest chain, as otherwise it will simply always discard new blocks being mined on the main chain.
- Following the previous point, it is possible to rewind and fast-forward on two completely separate chains.
- When downloading without a destination, the node chooses a new random peer for each new chunk (in the list of peers that have this chunk).
- When performing a file search, results can match any search request which has not yet completed (and will match as many as possible).
- When receiving a file search result that has our node as destination but does not correspond to an active search (i.e. not finished), we discard this search result. 
- Following the point above, we also discard / do not log search results which have already matched with all active searches. This ensures that a result received many times over the course of one expanding search (at each iteration) is not logged more than once.

## Notes about Implementation of HW2

Here are relevant details for whoever is reading / testing the code of Homework 2.

- Be careful about the `separatefs` flag explained above, make sure that you don't use it if you use the `_Download` and `_SharedFiles` folders directly. The tests use subfolders as it is easier to keep track of who owns what that way.
- In this implementation, the gossiper does not write the chunks /metafiles on the disk (my first version did it but I removed it). This is deliberate, as it added some overhead to write / read / cleanup etc... and is not necessary. Since the files will never go above 2Mb, I think it's fine to keep them in memory for now. Maybe I will write it on disk later again if we decide to allow larger files with a more complex (e.g. hierarchical) chunking algorithm.
- Since I used a lot more logs for debugging, I added the `verbose` tag which you can set to false or remove if you strictly need to see the output as required in the assignment. 
- The server / front-end could probably be optimized (for instance with web sockets) but since it is not the focus of this assignment I decided to leave it like this for now. It still does the job nicely though.
- When using the GUI, you can click on the usernames to send direct messages. Also you can upload / download files using the button-links in blue.
