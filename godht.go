package main

const (201 202 203 204

type interface dhtnode {
Query() os.Error
Error(int dht_error) os.Error
Respond(tld [2]byte payload []byte) os.Error
QPing() os.Error
QFindNode(tid [2]byte, nodeid [20]byte, target [20]byte) os.Error
QGetPeers
QAnnouncePeer(tid [2]byte,
RPing(tid [2]byte, nodeid [20]byte) os.Error
RFindNode
RGetPeers
RAnnouncePeer
sender

type struct node {
	pendingQueries map[string] [2]byte;
	nextTID [2]byte;
	nodeID [20]byte;
	knownPeers
	net.UDPConn socket;
}


func sender(outgoing chan []byte){
	for {
		msg := <-outgoing;
		this.getNextTID
		this.socket
	}
}

func getNextTID(){
	
}
