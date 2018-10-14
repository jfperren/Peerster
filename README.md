# Peerster




## Building

Instead of manually running `go build` in both `/` and `client/` folders, the whole project can be built by typing
```bash
scripts/build.sh
```

*Note - Even though the assignment mentionned something about renaming the `Peerster` exec into `gossiper`, the files provided did not do it. As a result, we will not do it here either and run `Peerster` directy.*

## Running

#### Via Command-Line

To run a gossiper node in CLI mode, use
```
./Peerster -gossipAddr=127.0.0.1:5000 -UIPort=8080 -name="Alice" -peers=127.0.0.1:5001
```

Then, you can send messages to the node directly via the command line `client` executable. To do this, use

```
client/client -UIPort=8080 "Your message here"
```

#### Via GUI

In order to interact with the gossiper via the GUI, you will need to run the `Peerster` executable with the `-server` mode. For instance,

```
./Peerster -gossipAddr=127.0.0.1:5000 -UIPort=8080 -name="Alice" -peers=127.0.0.1:5001 -server
```

This will start the web server and serve the GUI on `UIPort`. Therefore, simply connect to the GUI by accessing `localhost:8080` in your browser.

*Note - As the web server will be connecting to the `UIPort` directly, it is no longer possible to interact with the gossiper through the command line `client` executable.*

Alternatively, you can use the `run.sh` script that creates a simple 4-node ring. For each node, use the following command where `N` should be replaced by a value from 1 to 4.

```
scripts/run.sh N
```
