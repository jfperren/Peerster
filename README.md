# Peerster


## Building

Instead of manually running `go build` in both `/` and `client/` folders, the whole project can be built by typing
```
scripts/build.sh
```

*Note - Even though the assignment mentionned something about renaming the `Peerster` exec into `gossiper`, the files provided with the assignment handout did not. As a result, we will not do it here either and run `Peerster` directy. If the executable **must** be named `gossiper`, it is possible to uncomment the lines in `build.sh`.

## Running

#### Via Command-Line

To run a gossiper node in CLI mode, use
```
./Peerster -gossipAddr=127.0.0.1:5002 -UIPort=8082 -name="Charlie" -peers=127.0.0.1:5000
```

Alternatively, you can simply use
```
scripts/run.sh
```

When the gossiper is running, you can send messages to the node directly via the command line `client` executable. To do this, use

```
client/client -UIPort=8082 -msg="Your message here"
```

#### Using the GUI

In order to interact with the gossiper via the GUI, you will need to run the `Peerster` executable with the `-server` mode. For instance,

```
./Peerster -gossipAddr=127.0.0.1:5000 -UIPort=8080 -name="Alice" -peers=127.0.0.1:5001 -server
```

This will start the web server and serve the GUI on `UIPort`. Therefore, simply connect to the GUI by accessing `localhost:8080` in your browser.

*Note - As the web server will be connecting to the `UIPort` directly, it is no longer possible to interact with the gossiper through the command line `client` executable.*

Alternatively, there are also two pre-written scripts to start two nodes that communicate with each other (and with Charlie from the `run.sh` script!). 

```
scripts/run_server.sh
scripts/run_server_2.sh
```

## Testing

The two testing scripts are available in the `scripts` folder as well.

```
scripts/test_1_ring.sh
scripts/test_2_ring.sh
```
