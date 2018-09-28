
# Build client and gossiper
go build
cd client
go build
cd ..

# Rename Peerster into gossiper
mv ./Peerster ./gossiper
