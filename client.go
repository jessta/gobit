/*It is (49+len(pstr)) bytes long.
handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
pstrlen: string length of <pstr>, as a single raw byte
pstr: string identifier of the protocol
reserved: eight (8) reserved bytes. All current implementations use all zeroes. Each bit in these bytes can be used to change the behavior of the protocol. An email from Bram suggests that trailing bits should be used first, so that leading bits may be used to change the meaning of trailing bits.
info_hash: 20-byte SHA1 hash of the info key in the metainfo file. This is the same info_hash that is transmitted in tracker requests.
peer_id: 20-byte string used as a unique ID for the client. This is usually the same peer_id that is transmitted in tracker requests (but not always e.g. an anonymity option in Azureus).

pstrlen = 19, and pstr = "BitTorrent protocol"
peer_id = -AZ2060-
*/

const keepalive =0;
const (
choke = iota;
unchoke; 
interested;
not interested;
have;
bitfield;
request;
piece;
port;
)


/*
keep-alive: <len=0000>
The keep-alive message is a message with zero bytes, specified with the length prefix set to zero. There is no message ID and no payload. Peers may close a connection if they receive no messages (keep-alive or any other message) for a certain period of time, so a keep-alive message must be sent to maintain the connection alive if no command have been sent for a given amount of time. This amount of time is generally two minutes.

choke: <len=0001><id=0>
The choke message is fixed-length and has no payload.

unchoke: <len=0001><id=1>
The unchoke message is fixed-length and has no payload.

interested: <len=0001><id=2>
The interested message is fixed-length and has no payload.

not interested: <len=0001><id=3>
The not interested message is fixed-length and has no payload.
have: <len=0005><id=4><piece index>

bitfield: <len=0001+X><id=5><bitfield>

request: <len=0013><id=6><index><begin><length>
The request message is fixed length, and is used to request a block. The payload contains the following information:
index: integer specifying the zero-based piece index
begin: integer specifying the zero-based byte offset within the piece
length: integer specifying the requested length.

piece: <len=0009+X><id=7><index><begin><block>
The piece message is variable length, where X is the length of the block. The payload contains the following information:
index: integer specifying the zero-based piece index
begin: integer specifying the zero-based byte offset within the piece
block: block of data, which is a subset of the piece specified by index.

cancel: <len=0013><id=8><index><begin><length>
The cancel message is fixed length, and is used to cancel block requests. The payload is identical to that of the "request" message. It is typically used during "End Game" (see the Algorithms section below).

port: <len=0003><id=9><listen-port>
The port message is sent by newer versions of the Mainline that implements a DHT tracker. The listen port is the port this peer's DHT node is listening on. This peer should be inserted in the local routing table (if DHT tracker is supported).
*/
type handshake struct {
	pstrlen int8;
	pstr string;
	reserved [8]byte;
	info_hash [20]byte;
	peer_id [20]byte;
}

type message struct {
	length int32;
	msg_id  int8;
	payload *[]byte;
}

type bttracker interface {
	Request();
}

type btpeer interface {
	KeepAlive();
	Choke();
	Unchoke();
	Interested();
	NotInterested();
	Have();
	BitField();
	Request();
	Piece();
	Cancel();
	Port();
}

type client struct {
	bool am_unchoked;
	bool am_interested;
	bool peer_unchoked;
	bool peer_interested;
	net.Conn conn;
}

type torrent struct {
	clients []client;
	trackers []tracker;
}

/*piece: <len=0009+X><id=7><index><begin><block>*/
func (client *client) Piece(int32 index, int32 begin, block []byte) os.Error{
	var msg  [13]byte;
	msg[0:4] = 5;
	msg[5] = ; 
	msg[6:9] = piece;
	n,err := client.conn.Write(msg);
	return err;

}

/*cancel: <len=0013><id=8><index><begin><length>*/
func (client *client) Have(int32 index, int32 begin, int32 length,int32 piece) os.Error{
	msg[0:4] = 5;
	msg[5] = have; 
	msg[6:9] = piece;
	n,err := client.conn.Write(msg);
	return err;
}

/*port: <len=0003><id=9><listen-port>*/
func (client *client) Have(int16 port) os.Error{
	msg[0:4] = 5;
	msg[5] = have; 
	msg[6:9] = piece;
	n,err := client.conn.Write(msg);
	return err;
}
func (client *client) Have(int32 index) os.Error{
	var msg [9]byte;
	msg[0:4] = 5;
	msg[5] = have; 
	msg[6:9] = piece;
	n,err := client.conn.Write(msg);
	return err;

}
func (client *client) Request(int32 index,int32 begin, int32 length) os.Error{
	var msg [17]byte;
	msg[0:4] = 13;
	msg[5] = request; 
	msg[6:9] = index;
	msg[10:13] = begin;
	msg[14:17] = length;
	n,err := client.conn.Write(msg);
	return err;
}

func (client *client) Choke() os.Error{
	return msgNoPayLoad(choke);
}

func (client *client) Unchoke() os.Error{
	return msgNoPayLoad(unchoke);
}

func (client *client) Interested() os.Error{
	return msgNoPayLoad(interested);
}

func (client *client) Uninterested() os.Error{
	return msgNoPayLoad(uninterested);
}

func (client *client) Choke() os.Error{
	return msgNoPayLoad(choke);
}

func (client *client) msgNoPayload(int8 id) os.Error{
	var msg [5]byte;
	msg[0:4] = 1;
	msg[5] = id;
	n,err := client.conn.Write(msg);
	return err;
}

func (client *client) KeepAlive() os.Error{
	var length [4]byte = 0;
	n,err := client.conn.Write(length);
	return err;
}


