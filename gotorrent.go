package main

import "net"
import "os"
import "bytes"
import "runtime"
import "strconv"
import "encoding/binary"

const pstrlen = 19
const pstr = "BitTorrent protocol"
const peer_id = "-AZ2060-2"


const keepalive = 0
const (
	choke = iota
	unchoke
	interested
	uninterested
	have
	bitfield
	request
	piece
	cancel
	port
)

const (
	MaxRequestSize   = (1 << 14)
	MaxInRequestSize = (1 << 17)
	MaxInMsgSize     = MaxInRequestSize + 9
)

type handshake struct {
	pstrlen  uint8
	reserved [8]byte
}

type message struct {
	length  uint32
	msgId   uint8
	payLoad []byte
}

type bttracker interface {
	Request()
}

type btpeer interface {
	KeepAlive()
	Choke()
	Unchoke()
	Interested()
	NotInterested()
	Have()
	BitField()
	Request()
	Piece()
	Cancel()
	Port()
}

type Client struct {
	amUnchoked     bool
	amInterested   bool
	peerUnchoked   bool
	peerInterested bool
	addr           net.TCPAddr
	conn           net.Conn
	torrent        *torrent
	buffer         [MaxInMsgSize]byte
	bitfield       *[]byte
	todo           *chan message
	peerId		string
}

type tracker struct {
	name string
}

type torrent struct {
	trackers  []tracker
	Pstr      string
	info_hash [20]byte
	peer_id   [20]byte
}

/*handshake: <pstrlen><pstr><reserved><info_hash><peer_id>*/
func (c *Client) HandShake() os.Error {
	var pstrlen uint8 = (uint8)(len(c.torrent.Pstr))
	blank := make([]byte, 9)
	msg := make([]byte, 49+pstrlen) //49 = pstrlen(1)+ reserved(20) + infohash(20) + peerid(8)
	buffer := bytes.NewBuffer(msg)
	buffer.WriteByte(pstrlen)
	buffer.WriteString(c.torrent.Pstr)
	buffer.Write(blank)
	buffer.Write(&(c.torrent.info_hash))
	buffer.Write(&(c.torrent.peer_id))
	_, err := c.conn.Write(buffer.Bytes())
	return err
}

/*piece: <len=0009+X><id=7><index><begin><block>*/
func (c *Client) Piece(index uint32, begin uint32, block []byte) os.Error {
	blocksize := (uint32)(len(block))
	msg := make([]byte, 9)
	binary.BigEndian.PutUint32(msg[0:3], blocksize+9)
	binary.BigEndian.PutUint32(msg[5:8], index)
	binary.BigEndian.PutUint32(msg[9:12], begin)
	msg[4] = piece
	_, err := c.conn.Write(msg)
	if err == nil {
		_, err2 := c.conn.Write(block)
		return err2
	}
	return err

}

/*cancel: <len=0013><id=8><index><begin><length>*/
func (c *Client) Cancel(index uint32, begin uint32, length uint32) os.Error {
	msg := make([]byte, 17)
	binary.BigEndian.PutUint32(msg[0:3], 13)
	msg[4] = cancel
	binary.BigEndian.PutUint32(msg[5:8], index)
	binary.BigEndian.PutUint32(msg[9:12], begin)
	binary.BigEndian.PutUint32(msg[13:16], length)
	_, err := c.conn.Write(msg)
	return err
}

/*port: <len=0003><id=9><listen-port>*/
func (c *Client) Port(portnum uint16) os.Error {
	msg := make([]byte, 7)
	binary.BigEndian.PutUint32(msg[0:3], 3)
	msg[4] = port
	binary.BigEndian.PutUint16(msg[5:6], portnum)
	_, err := c.conn.Write(msg)
	return err
}
func (c *Client) Have(index uint32) os.Error {
	msg := make([]byte, 9)
	binary.BigEndian.PutUint32(msg[0:3], 5)
	binary.BigEndian.PutUint32(msg[5:8], piece)
	msg[4] = have
	_, err := c.conn.Write(msg)
	return err

}
func (c *Client) Request(index uint32, begin uint32, length uint32) os.Error {
	msg := make([]byte, 17)
	binary.BigEndian.PutUint32(msg[0:3], 13)
	binary.BigEndian.PutUint32(msg[5:8], index)
	binary.BigEndian.PutUint32(msg[9:12], begin)
	binary.BigEndian.PutUint32(msg[13:16], length)
	msg[4] = request
	_, err := c.conn.Write(msg)
	return err
}

func (c *Client) msgNoPayLoad(id uint8) os.Error {
	msg := make([]byte, 5)
	binary.BigEndian.PutUint32(msg[0:4], 1)
	msg[4] = id
	_, err := c.conn.Write(msg)
	return err
}

func (c *Client) Unchoke() os.Error { return c.msgNoPayLoad(unchoke) }

func (c *Client) Interested() os.Error { return c.msgNoPayLoad(interested) }

func (c *Client) Uninterested() os.Error { return c.msgNoPayLoad(uninterested) }

func (c *Client) Choke() os.Error { return c.msgNoPayLoad(choke) }

func (c *Client) KeepAlive() os.Error {
	msg := make([]byte, 4)
	binary.BigEndian.PutUint32(msg[0:4], 0)
	_, err := c.conn.Write(msg)
	return err
}
func (c *Client) processMsg(msg *message) (err os.Error) {
	err = nil
	print("processing msg\n")
	switch msg.msgId {
	case choke:
		c.peerUnchoked = false
		print("choke msg\n")
	case unchoke:
		c.peerUnchoked = true
		print("unchoke msg\n")
	case interested:
		c.peerInterested = true
		print("interested msg\n")
	case uninterested:
		c.peerInterested = false
		print("uninterested msg\n")
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
	case port:
		fallthrough //ignore for now
	default:
		err = os.NewError("Recieved unknown message")

	}
	return
}

func (c *Client) processBitField(msg *message) (err os.Error) {
	size := (binary.LittleEndian.Uint32(c.buffer[0:3]) - 1)
	bits := make([]byte, size)
	buffer := bytes.NewBuffer(bits)
	buffer.Write(c.buffer[2:(size + 1)])
	tmp := buffer.Bytes()
	c.bitfield = &tmp
	return
}

func (c *Client) processHave(msg *message) (err os.Error) {
	print("recieved have msg")
	return
}
func (c *Client) processPiece(msg *message) (err os.Error) {
	print("recieved piece msg")
	return
}

func (c *Client) processRequest(msg *message) (err os.Error) {
	print("recieved request msg")
	return
}

func (c *Client) processCancel(msg *message) (err os.Error) {
	print("recieved cancel msg")
	return
}
func (c *Client) processHandShake(msg *message) (err os.Error) {
        print("recieved handshake msg")
	return
}

/*handshake: <pstrlen><pstr><reserved><info_hash><peer_id>*/
func (c *Client) WaitHandShake() (msg *message, err os.Error) {
	_, err = c.conn.Read(c.buffer[0:1]) //read pstr length
	if err != nil {
		return err
	}
	var num int;
	if num, err := strconv.Atoi(string(c.buffer[0:1]));err != nil || num != 19 {
		return err
	}
	_,err = c.conn.Read(c.buffer[1:ptrstrlen+1])
	if err != nil {
		return err
	}
	if ptrstr != string(c.buffer[1:ptrstrlen+1]) {
		return os.NewError("bad ptrstr")
	}
	msg,err = c.conn.Read(c.buffer[1:49+uint8(c.buffer[0:1][0])])
	msg 	
	return msg,err;
}

func NewClient(addr *net.TCPAddr) (c *Client) {
	c = new(Client)
	if addr != nil {
		c.addr = *addr
	}
	return c
}

func (c *Client) run(complete chan bool) {
	print("client running\n")
	print("sending handshake\n")
	c.HandShake()
	print("waiting for handshake\n")
	if !c.WaitHandShake(){
		complete <- false
		return
	}

	c.processHandShake(msg)
	blank := make([]byte, 9)
	msg := make([]byte, 49+pstrlen)
	buffer := bytes.NewBuffer(msg)
	buffer.WriteByte(pstrlen)
	buffer.WriteString(c.torrent.Pstr)
	buffer.Write(blank)
	buffer.Write(&(c.torrent.info_hash))
	buffer.Write(&(c.torrent.peer_id))
	_, err = c.conn.Write(buffer.Bytes())
	return

	for {
		print("waiting for msg\n")
		_, err := c.conn.Read(c.buffer[0:4]) //read msg length
		if err != nil {
			complete <- false
			return
		}
		
		length := binary.BigEndian.Uint32(c.buffer[0:4])
		print("recieved msg length"+strconv.Itoa(int(length))+"\n")
		switch {
		case length == 0:
			print("keepalive msg")
			continue

		case length >= MaxInMsgSize:
			complete <- true
			return
		}
		_, err = c.conn.Read(c.buffer[4 : 4+length])
		if err != nil {
			complete <- true
			return
		}
		print("msg recieved payload\n")
		msg := new(message)
		msg.length = length
		msg.msgId = c.buffer[4]
		bytes.Add(msg.payLoad, c.buffer[5:len(c.buffer)])
		err = c.processMsg(msg)
		if err != nil {
			print(err.String())
			complete <- true
			return
		}
		runtime.Gosched()

	}

}

func (c *Client) SetConn(conn net.Conn) { c.conn = conn }

func (c *Client) Connect() (err os.Error) {
	c.conn, err = net.DialTCP("tcp", nil, &c.addr)
	return err
}

func (c *Client) Run() (complete chan bool, err os.Error) {
	err = nil
	if c.conn == nil {
		err = c.Connect()
		if err != nil {
			return
		}
	}
	complete = make(chan bool)
	go c.run(complete)
	return
}

func runServer(laddr *net.TCPAddr, kill chan bool, complete chan bool, clients chan chan bool) {

		for {
			/*select {
				case <-kill:
					print("server recieved kill");
					complete<-true;
			}*/
			listen, err := net.ListenTCP("tcp", laddr)
			if err != nil {
				os.Exit(1)
			}
			conn, err := listen.AcceptTCP()
			if err != nil {
				complete <- false
			}
			client := NewClient(nil)
			client.SetConn(conn)
			print("client connected\n")


			if err = client.Unchoke();err != nil {
				print("unchoke error\n")
				print(err.String())
				continue
			}
			print("unchoke sent\n")


	if err = client.Choke();err != nil {
				print("choke error\n")
				print(err.String())
				continue
			}
			print("choke sent\n")


	if err = client.Interested();err != nil {
				print("interested error\n")
				print(err.String())
				continue
			}
			print("interested sent\n")


	if err = client.KeepAlive();err != nil {
				print("keepalive error\n")
				print(err.String())
				continue
			}
			print("keepalive sent\n")


	if err = client.Uninterested();err != nil {
				print("uninterested error\n")
				print(err.String())
				continue
			}
			print("uninterested sent\n")

			clientComp, err := client.Run()

			if err != nil {
				print("client error\n")
				print(err.String())
				continue
			}
			clients <- clientComp
		}
		complete <- true
}
func RunServer(laddr *net.TCPAddr) (kill chan bool, complete chan bool, clients chan chan bool, err os.Error) {
	err = nil
	complete = make(chan bool)
	kill = make(chan bool)
	clients = make(chan chan bool)
	go runServer(laddr, kill,complete,clients)

	return
}

func main() {
	addr, _ := net.ResolveTCPAddr(os.Args[2])
	_, serverComp, clients, err := RunServer(addr)
	if err != nil {
		print("server error\n")
		print(err.String())
		os.Exit(1)
	}
	address, err := net.ResolveTCPAddr(os.Args[1])
	client := NewClient(address)
	clientComp, err := client.Run()
	if err != nil {
		print("client error\n")
		print(err.String())
		os.Exit(1);
	}

	<-clientComp
	for i := range (clients) {
		<-i
	}
	//killServer<-true
	<-serverComp

}
