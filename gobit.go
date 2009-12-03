package main
import "./gotorrent"
import "./bencode"
import "bufio"
import "os"
import "fmt"
import "net"
import "container/list"
import "strconv"
import "runtime"

const lowerPort = 6881
const upperPort = 6999
/*
client manager:
 there is a list of clients that we have connected too or have connected to us.
 the manager loops through all the clients and gives them instructions the clients reply when their task is complete

Clients are given tasks, eg. get piece, etc
*/
func main(){
	runtime.GOMAXPROCS(3);
	clientList := list.New();
	servererr := make(chan os.Error);
	reportclient := make(chan *gotorrent.Client);
	trackererr := make(chan os.Error);
	_=trackererr;

	buff := new(bencode.BeString);
	reader := bufio.NewReader(os.Stdin);
	be, err := buff.Decode(reader);
	if err != nil {
		fmt.Printf("%s\n", err.String());
		os.Exit(1);
	}
	//numPieces = be['pieces']/pieceHashSize;
	//build piece map

	go server(servererr,reportclient);
	// we listen for connections
	// and we make connections
	/* find clients to connect to from tracker or DHT*/
	/*tracker := new(gotracker);
	go gotracker.Run(trackererr,reportclient);*/
	//main management loop
	print("entering select\n");
	var newclient *gotorrent.Client;
	var error os.Error;
	for {
		//print("selecting...\n");
		select {
			case newclient = <-reportclient:
				clientList.PushBack(newclient);
				print("new client spawned\n");
			case error= <-servererr:
				print("server had error:",error.String(),"\n");
				os.Exit(1);
			}
		}

	_=be;

}

func server(error chan os.Error, reportclient chan *gotorrent.Client){
	var conn net.Listener;
	var err os.Error;
	for port := lowerPort; port <= upperPort; port++ {
		netstr,err := net.ResolveTCPAddr("0.0.0.0:"+strconv.Itoa(port));
		if err != nil {error<-err;}
		conn,err =net.ListenTCP("tcp4", netstr);
		if err != nil{
			if(err == os.EADDRINUSE){
				continue;
			} else {
				error<-err;
			}
		} else {
			print("spawned server on port: "+strconv.Itoa(port)+"\n");
			break;
		}
	}
	if err != nil {error<-err;}

	for{
		print("waiting for connections\n");
		if clientConn, err := conn.Accept(); err==nil {
			print("client accepted\n");
			client := gotorrent.NewClient(conn.Addr());
			client.SetConn(clientConn);

			go client.Run();
			reportclient <- client;
		}else {
			error <- err;
			return;
		}
	}


}
