// Implementation of a KeyValueServer. Students should write their code in this file.

package p0partA

import (
	"bufio"
	"fmt"
	"github.com/cmu440/p0partA/kvstore"
	"github.com/k0kubun/pp/v3"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var logger = log.New(os.Stdout, "[server impl]", log.Lshortfile)

func info(pattern string, args ...interface{}) {
	logger.Output(2, fmt.Sprintf(pattern, args...))
}

type keyValueServer struct {
	// TODO: implement this!
	store          *kvstore.KVStore
	commandChannel chan *command
	aliveRequest   chan bool
	aliveResult    chan int
	dropRequest    chan bool
	dropResult     chan int
	close          chan bool
	aliveCount     int
	droppedCount   int
	clients        map[string]*client
	listener       net.Listener
	listenerClosed chan bool
	closed         bool
}
type client struct {
	ip        string
	kvs       *keyValueServer
	conn      net.Conn
	writeChan chan string
	writeNoti chan string
	readNoti  chan string
	readChan  chan string
	closeChan chan bool
}
type command struct {
	ip          string
	commandType string
	payload     interface{}
}

// New creates and returns (but does not start) a new KeyValueServer.
func New(store kvstore.KVStore) KeyValueServer {
	return &keyValueServer{
		store:          &store,
		aliveCount:     0,
		droppedCount:   0,
		commandChannel: make(chan *command),
		aliveRequest:   make(chan bool),
		aliveResult:    make(chan int),
		dropRequest:    make(chan bool),
		dropResult:     make(chan int),
		close:          make(chan bool),
		listenerClosed: make(chan bool),
		clients:        make(map[string]*client),
		closed:         false,
	}
}

func (kvs *keyValueServer) Start(port int) error {
	// TODO: implement this!
	//need to listen to tcp port, accept tcp connection
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	kvs.listener = listener
	if err != nil {
		info("create listener fail! exiting...\n")
		return err
	}
	info("waiting for connection...\n")
	go kvs.listenHandler(listener)
	go kvs.mainRoutine()
	return nil
}
func (kvs *keyValueServer) mainRoutine() {
	for {
		if !kvs.closed {
			select {
			case <-kvs.aliveRequest:
				kvs.aliveResult <- kvs.aliveCount
			case <-kvs.dropRequest:
				kvs.dropResult <- kvs.droppedCount
			case cmd := <-kvs.commandChannel:
				kvs.commandHandler(cmd)
			case <-kvs.close:
				kvs.closed = true
				kvs.listener.Close()
				info("listener closed , graceful shutdown")
				return
			case <-kvs.listenerClosed:
				return
			}
		} else {
			select {
			case <-kvs.aliveRequest:
				kvs.aliveResult <- kvs.aliveCount
			case <-kvs.dropRequest:
				kvs.dropResult <- kvs.aliveCount
			case cmd := <-kvs.commandChannel:
				kvs.commandHandler(cmd)
			case <-kvs.close:
				kvs.listener.Close()
				kvs.closed = true
				info("listener closed , graceful shutdown")
				return
			}
		}

	}
}
func (kvs *keyValueServer) registerCommand(cmd *command) {
	kvs.commandChannel <- cmd
}
func (kvs *keyValueServer) commandHandler(cmd *command) {
	switch cmd.commandType {
	case "ADD":
		c, ok := cmd.payload.(*client)
		if !ok {
			info("%v", cmd.payload)
			return
		}
		kvs.addReal(c)
	case "DROP":
		kvs.dropReal(cmd.ip)
	case "Put":
		content := cmd.payload.([]string)
		store := *kvs.store
		store.Put(content[0], []byte(content[1]))
	case "Get":
		content := cmd.payload.(string)
		store := *kvs.store
		val := store.Get(content)
		kvs.handleWrite(cmd, val)
	case "Delete":
		store := *kvs.store
		content := cmd.payload.(string)
		store.Delete(content)
	case "Update":
		store := *kvs.store
		content := cmd.payload.([]string)
		store.Update(content[0], []byte(content[1]), []byte(content[2]))
	default:
		info("command unrecognized")
	}
}
func (kvs *keyValueServer) handleWrite(cmd *command, val [][]byte) {
	//register write to client
	c, ok := kvs.clients[cmd.ip]
	if !ok {
		info("client: %v does not exist!\n", cmd.ip)
		return
	}
	content := cmd.payload.(string)
	for _, v := range val {
		go c.write(content + ":" + string(v))
	}
}

func (kvs *keyValueServer) addReal(c *client) {
	kvs.clients[c.ip] = c
	kvs.aliveCount++
}
func (kvs *keyValueServer) dropReal(ip string) {
	delete(kvs.clients, ip)
	kvs.droppedCount++
	kvs.aliveCount--
}

func (kvs *keyValueServer) Close() {
	// TODO: implement this!
	kvs.close <- true
	<-kvs.listenerClosed
}

func (kvs *keyValueServer) CountActive() int {
	// TODO: implement this!
	kvs.aliveRequest <- true
	res := <-kvs.aliveResult
	return res
}

func (kvs *keyValueServer) CountDropped() int {
	// TODO: implement this!
	kvs.dropRequest <- true
	res := <-kvs.dropResult
	return res
}
func (kvs *keyValueServer) listenHandler(listener net.Listener) error {
	for {
		//block listening...
		conn, err := listener.Accept()
		if err != nil {
			info("Error accepting %v\n", err.Error())
			kvs.listenerClosed <- true
			return err
		}
		info("connection accepted...\n")
		go kvs.HandleConnection(conn)
	}

}

// TODO: add additional methods/functions below!
func (kvs *keyValueServer) HandleConnection(conn net.Conn) {
	//client routine two branch read routine and write routine
	//should create a client, client have its main routine, it has to be able to interact with kvs?
	fmt.Println(conn.RemoteAddr().String())
	c := &client{
		conn:      conn,
		ip:        conn.RemoteAddr().String(),
		kvs:       kvs,
		writeChan: make(chan string, 500),
		readChan:  make(chan string, 500),
		writeNoti: make(chan string),
		readNoti:  make(chan string),
		closeChan: make(chan bool),
	}
	com := &command{
		ip:          c.ip,
		commandType: "ADD",
		payload:     c,
	}
	kvs.registerCommand(com)
	go c.mainRoutine()
	go c.readRoutine()
	go c.writeRoutine()

}
func (c *client) mainRoutine() {
	for {
		select {
		case str := <-c.writeNoti:
			c.handleWrite(str)
		case str := <-c.readNoti:
			c.handleRead(str)
		case <-c.closeChan:
			return
		}
	}
}
func (c *client) handleWrite(str string) {
	if len(c.writeChan) == cap(c.writeChan) {
		//do nothing
		return
	}
	c.writeChan <- str

}
func (c *client) handleRead(str string) {
	//first format str to command, then send to server
	com := &command{
		ip: c.ip,
	}
	strArr := strings.Split(str, ":")
	com.commandType = strArr[0]
	switch strArr[0] {
	case "Put":
		if len(strArr) != 3 {
			info("argument number incorrect, got: '%v'\n", str)
			pp.Println(strArr)
			return
		}
		com.payload = strArr[1:]
	case "Get":
		if len(strArr) != 2 {
			info("argument number incorrect, got: '%v'\n", str)
			return
		}
		com.payload = strArr[1]
	case "Delete":
		if len(strArr) != 2 {
			info("argument number incorrect, got: '%v'\n", str)
			return
		}
		com.payload = strArr[1]
	case "Update":
		if len(strArr) != 4 {
			info("argument number incorrect, got: '%v'\n", str)
			return
		}
		com.payload = strArr[1:]
	default:
		info("unrecognized command: %v, %s\n", strArr[0], str)
		return
	}
	c.kvs.registerCommand(com)

}
func (c *client) readRoutine() {
	count := 0

	reader := bufio.NewReader(c.conn)
	for {
		count++

		str, err := reader.ReadString('\n')
		trimmedStr := strings.Trim(str, "\n")
		if err != nil {
			//info("Error reading... %v \n", err.Error())
			//delete
			c.handleClose()
			return
		}
		//info("%d: Received data: '%v'", count, trimmedStr)
		c.readNoti <- trimmedStr
		//info("waiting?")
		//parse info into command send to kv
	}

}
func (c *client) handleClose() {
	cmd := &command{
		ip:          c.ip,
		commandType: "DROP",
	}
	c.kvs.registerCommand(cmd)
	c.closeChan <- true
}
func (c *client) writeRoutine() {
	for {
		select {
		case str := <-c.writeChan:
			c.conn.Write(append([]byte(str), '\n'))
		case <-c.closeChan:
			return
		}

	}
}
func (c *client) write(str string) {
	c.writeNoti <- str
}
