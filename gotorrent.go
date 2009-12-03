package gotorrent

import "net"
import "os"
import "bytes"
//import "strings"
import "encoding/binary"
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

const keepalive = 0
const (
	choke	= iota;
	unchoke;
	interested;
	uninterested;
	have;
	bitfield;
	request;
	piece;
	cancel;
	port;
)

const (
	MaxRequestSize		= (1 << 14);
	MaxInRequestSize	= (1 << 17);
	MaxInMsgSize		= MaxInRequestSize + 9;
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
	pstrlen		uint8;
	reserved	[8]byte;
}

type message struct {
	length	uint32;
	msgId	uint8;
	payLoad	[]byte;
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

type Client struct {
	amUnchoked	bool;
	amInterested	bool;
	peerUnchoked	bool;
	peerInterested	bool;
	addr		net.Addr;
	conn		net.Conn;
	torrent		*torrent;
	buffer		[MaxInMsgSize]byte;
	bitfield	*[]byte;
	todo		*chan message;
}

type tracker struct {
	name string;
}

type torrent struct {
	trackers	[]tracker;
	Pstr		string;
	info_hash	[20]byte;
	peer_id		[20]byte;
}

/*handshake: <pstrlen><pstr><reserved><info_hash><peer_id>*/
func (c *Client) HandShake() os.Error {
	var pstrlen uint8 = (uint8)(len(c.torrent.Pstr));
	blank := make([]byte,9);
	msg := make([]byte, 49+pstrlen);
	buffer := bytes.NewBuffer(msg);
	buffer.WriteByte(pstrlen);
	buffer.WriteString(c.torrent.Pstr);
	buffer.Write(blank);
	buffer.Write(&(c.torrent.info_hash));
	buffer.Write(&(c.torrent.peer_id));
	_, err := c.conn.Write(buffer.Bytes());
	return err;
}

/*piece: <len=0009+X><id=7><index><begin><block>*/
func (c *Client) Piece(index uint32, begin uint32, block []byte) os.Error {
	blocksize := (uint32)(len(block));
	msg := make([]byte, 9);
	binary.BigEndian.PutUint32(msg[0:3], blocksize+9);
	binary.BigEndian.PutUint32(msg[5:8], index);
	binary.BigEndian.PutUint32(msg[9:12], begin);
	msg[4] = piece;
	_, err := c.conn.Write(msg);
	if err == nil {
		_, err2 := c.conn.Write(block);
		return err2;
	}
	return err;

}

/*cancel: <len=0013><id=8><index><begin><length>*/
func (c *Client) Cancel(index uint32, begin uint32, length uint32) os.Error {
	msg := make([]byte, 17);
	binary.BigEndian.PutUint32(msg[0:3], 13);
	msg[4] = cancel;
	binary.BigEndian.PutUint32(msg[5:8], index);
	binary.BigEndian.PutUint32(msg[9:12], begin);
	binary.BigEndian.PutUint32(msg[13:16], length);
	_, err := c.conn.Write(msg);
	return err;
}

/*port: <len=0003><id=9><listen-port>*/
func (c *Client) Port(portnum uint16) os.Error {
	msg := make([]byte, 7);
	binary.BigEndian.PutUint32(msg[0:3], 3);
	msg[4] = port;
	binary.BigEndian.PutUint16(msg[5:6], portnum);
	_, err := c.conn.Write(msg);
	return err;
}
func (c *Client) Have(index uint32) os.Error {
	msg := make([]byte, 9);
	binary.BigEndian.PutUint32(msg[0:3], 5);
	binary.BigEndian.PutUint32(msg[5:8], piece);
	msg[4] = have;
	_, err := c.conn.Write(msg);
	return err;

}
func (c *Client) Request(index uint32, begin uint32, length uint32) os.Error {
	msg := make([]byte, 17);
	binary.BigEndian.PutUint32(msg[0:3], 13);
	binary.BigEndian.PutUint32(msg[5:8], index);
	binary.BigEndian.PutUint32(msg[9:12], begin);
	binary.BigEndian.PutUint32(msg[13:16], length);
	msg[4] = request;
	_, err := c.conn.Write(msg);
	return err;
}

func (c *Client) msgNoPayLoad(id uint8) os.Error {
	msg := make([]byte, 5);
	binary.BigEndian.PutUint32(msg[0:3], 1);
	msg[4] = id;
	_, err := c.conn.Write(msg);
	return err;
}

func (c *Client) Unchoke() os.Error	{ return c.msgNoPayLoad(unchoke) }

func (c *Client) Interested() os.Error	{ return c.msgNoPayLoad(interested) }

func (c *Client) Uninterested() os.Error	{ return c.msgNoPayLoad(uninterested) }

func (c *Client) Choke() os.Error	{ return c.msgNoPayLoad(choke) }

func (c *Client) KeepAlive() os.Error {
	msg := make([]byte, 4);
	binary.BigEndian.PutUint32(msg[0:4], 0);
	_, err := c.conn.Write(msg);
	return err;
}
func (c *Client) processMsg(msg *message) (err os.Error){
	err = nil;
		switch msg.msgId {
		case choke:
			c.peerUnchoked = false
		case unchoke:
			c.peerUnchoked = true
		case interested:
			c.peerInterested = true
		case uninterested:
			c.peerInterested = false
		case have:
			err = c.processHave(msg)
		case bitfield:
			err = c.processBitField(msg)
		case request:
			err = c.processRequest(msg)
		case piece:
			err = c.processPiece(msg)
		case cancel:
			err = c.processCancel(msg)
		case port: fallthrough; //ignore for now
		default:
			err = os.NewError("Recieved unknown message");

		}
		return
}
func (c *Client) processBitField(msg *message) (err os.Error) {
	size := (binary.LittleEndian.Uint32(c.buffer[0:3]) - 1);
	bits := make([]byte, size);
	buffer := bytes.NewBuffer(bits);
	buffer.Write(c.buffer[2:(size+1)]);
	tmp := buffer.Bytes();
	c.bitfield = &tmp;
	return;
}

func (c *Client) processHave(msg *message) (err os.Error)	{print("recieved have msg"); return}
func (c *Client) processPiece(msg *message) (err os.Error){ print("recieved piece msg"); return }

func (c *Client) processRequest(msg *message) (err os.Error)	{ print("recieved request msg"); return }

func (c *Client) processCancel(msg *message) (err os.Error)	{print("recieved cancel msg"); return}

func NewClient(addr net.Addr) (c *Client){
	c = new(Client);
	c.addr = addr;
	return c
}

func (c *Client) run(){
	for {
		_, err := c.conn.Read(c.buffer[0:3]);	//read msg length
		if err != nil {
			return
		}
		length := binary.LittleEndian.Uint32(c.buffer[0:3]);
		switch {
		case length == 0:
			//keep alive
			continue

		case length >= MaxInMsgSize:
			//too big
			return
		}
		_, err = c.conn.Read(c.buffer[4:length-1]);
		if err != nil {
			return
		}
		msg := new(message);
		msg.length = length;
		msg.msgId = c.buffer[4];
		bytes.Add(msg.payLoad, c.buffer[5:len(c.buffer)]);
		c.processMsg(msg);

	}

}

func (c *Client) SetConn(conn net.Conn){
	c.conn = conn;
}

func (c *Client) Connect() (err os.Error){
	return nil;
}

func (c *Client) Run() (err os.Error) {
	if(c.conn == nil){
		err = c.Connect();
		if(err != nil){
			return;
		}
	}
	go c.run();
	return;
}
