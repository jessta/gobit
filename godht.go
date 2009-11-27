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


type struct request {
	rtid [2]byte;
	rtype int;
}

type struct node {
	pending map[string] *request;
	nextTID [2]byte;
	nodeID [20]byte;
	knownPeers
}
