package listener

import (
	"github.com/jeffpierce/cassabon/config"
	"github.com/jeffpierce/cassabon/pearson"
)

type indexedLine struct {
	peerIndex int
	statLine  string
}

// peerList contains an ordered list of Cassabon peers.
type peerList struct {
	target   chan indexedLine // Channel for forwarding a stat line to a Cassabon peer
	hostPort string           // Host:port on which the local server is listening
	peers    []string         // Host:port information for all Cassabon peers (inclusive)
}

// isInitialized indicates whether the structure has ever been updated.
func (pl *peerList) isInitialized() bool {
	return pl.hostPort != ""
}

// start records the current peer list and starts the forwarder goroutine.
func (pl *peerList) start(hostPort string, peers []string) {

	// Create the channel on which stats to forward are received.
	pl.target = make(chan indexedLine, 1)

	// Record the current set of peers.
	pl.hostPort = hostPort
	pl.peers = make([]string, len(peers))
	for i, v := range peers {
		pl.peers[i] = v
	}

	// Start the forwarder goroutine.
	config.G.OnReload2WG.Add(1)
	go pl.run()
}

// isEqual indicates whether the given new configuration is equal to the current.
func (pl *peerList) isEqual(hostPort string, peers []string) bool {
	if pl.hostPort != hostPort {
		return false
	}
	if len(pl.peers) != len(peers) {
		return false
	}
	for i, v := range pl.peers {
		if peers[i] != v {
			return false
		}
	}
	return true
}

// ownerOf determines which host owns a particular stats path.
func (pl *peerList) ownerOf(statPath string) (int, bool) {
	peerIndex := int(pearson.Hash8(statPath)) % len(pl.peers)
	if pl.hostPort == pl.peers[peerIndex] {
		config.G.Log.System.LogInfo("Mine! %-30s %d %s", statPath, peerIndex, pl.peers[peerIndex])
		return peerIndex, true
	} else {
		//config.G.Log.System.LogInfo("      %-30s %d %s", statPath, peerIndex, pl.peers[peerIndex])
		return peerIndex, false
	}
}

// run listens for stat lines on a channel and sends them to the appropriate Cassabon peer.
func (pl *peerList) run() {

	defer close(pl.target)

	for {
		select {
		case <-config.G.OnReload2:
			config.G.Log.System.LogDebug("peerList::run received QUIT message")
			config.G.OnReload2WG.Done()
			return
		case il := <-pl.target:
			if pl.hostPort != pl.peers[il.peerIndex] {
				config.G.Log.System.LogInfo("Forwarding to %d %s: \"%s\"",
					il.peerIndex, pl.peers[il.peerIndex], il.statLine)
			}
		}
	}
}
