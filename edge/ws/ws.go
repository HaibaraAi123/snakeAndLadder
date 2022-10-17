package ws

import (
	"context"
	"fmt"
	"go.uber.org/atomic"
	"golang.org/x/net/websocket"
	"log"
	"snakeAndLadder/utl"
	"strings"
	"sync"
	"time"
)

var (
	WebsocketServer *websocketServer
)

func NewWebsocketServer() (*websocketServer, error) {
	svr := &websocketServer{}
	svr.Init()
	return svr, nil
}

func WebsocketHandler(conn *websocket.Conn) {
	fn := "WsHandleFunc"
	remoteIp := utl.GetRemoteIp(conn.Request())
	log.Printf("%v, remoteIp:%v", fn, remoteIp)
	defer func(conn *websocket.Conn) {
		log.Println("close conn of ws.", time.Now())
		err := conn.Close()
		if err != nil {
			log.Fatalf("err in close conn:%v", err)
		}
	}(conn)
	worker := &websocketWorker{}
	worker.Init(conn)
	WebsocketServer.addWorker(worker)
	defer WebsocketServer.rmWorker(worker.connId)
	worker.Handler()
}

type websocketServer struct {
	// map [connId]*worker	map for reverse
	workMap map[string]*websocketWorker
	mu      *sync.RWMutex
}

func (w *websocketServer) Init() {
	w.workMap = make(map[string]*websocketWorker)
	w.mu = &sync.RWMutex{}
}

func (w *websocketServer) Handler(*websocket.Conn) {

}

func (w *websocketServer) addWorker(worker *websocketWorker) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.workMap[worker.connId] = worker
}

func (w *websocketServer) rmWorker(connId string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.workMap, connId)
}

func (w *websocketServer) getWorker(connId string) *websocketWorker {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.workMap[connId]
}

type websocketWorker struct {
	connId   string
	conn     *websocket.Conn
	sender   chan *Msg
	receiver chan *Msg
	done     chan struct{}
	closeMu  sync.Mutex
	wg       sync.WaitGroup
	exit     chan struct{}
	caller   map[uint32]chan *Msg // reverse call, wait resp
	callerMu sync.RWMutex
	seq      atomic.Uint32
}

func (w *websocketWorker) Init(conn *websocket.Conn) {
	w.sender, w.receiver = make(chan *Msg, 10), make(chan *Msg, 10)
	w.done, w.exit = make(chan struct{}), make(chan struct{})
	w.connId = fmt.Sprintf("%v-%v", time.Now().Format(time.RFC3339), utl.GetRemoteIp(conn.Request()))
	w.caller = make(map[uint32]chan *Msg)
}

func (w *websocketWorker) Handler() {
	w.wg.Add(2)
	go w.send()
	w.receive()
	<-w.exit
	w.wg.Wait()
}

func (w *websocketWorker) send() {
	fn := "wsWorker"
	for {
		select {
		case <-w.done:
			log.Printf("wsWorker leave, connId:%v", w.connId)
			w.wg.Done()
			return
		case msg := <-w.receiver:
			log.Printf("%v, send, msg:%v", fn, msg)
			err := websocket.JSON.Send(w.conn, msg)
			if err != nil {
				log.Printf("%v, sendMsg fail, err:%v", fn, err)
				w.wg.Done()
				w.close()
				return
			}
		}
	}

}

func (w *websocketWorker) receive() {
	fn := "wsWorker"
	for {
		select {
		case <-w.done:
			log.Printf("wsWorker leave, connId:%v", w.connId)
			w.wg.Done()
			return
		default:
			msg := &Msg{}
			err := websocket.JSON.Receive(w.conn, msg)
			if err != nil {
				log.Printf("%v, receive msg fail, err:%v", fn, err)
				w.wg.Done()
				w.close()
				return
			} else {
				log.Printf("%v, recieve msg:%v", fn, msg)
			}
			w.wg.Add(1)
			go w.dispatch(msg)
		}
	}
}

func (w *websocketWorker) close() {
	w.closeMu.Lock()
	defer w.closeMu.Unlock()
	close(w.done)
	w.wg.Wait()
	err := w.conn.Close()
	if err != nil {
		log.Printf("close websocket.conn fail, err:%v", err)
	}
	close(w.exit) // 是否区分exit 与 done
}

func (w *websocketWorker) dispatch(msg *Msg) {
	fn := "dispatch"
	defer w.wg.Done()
	switch msg.Type {
	case TypeHeatBreak:
		resp := &Msg{
			Type: TypeHeatBreak,
			Seq:  msg.Seq,
		}
		w.sender <- resp
	case TypeMsgRequest:
		// dispatch to grpc
		method := string(msg.Method)
		l := strings.Split(method, "/")
		if len(l) != 3 {
			log.Printf("%v, msg invalid:%v", fn, msg)
			return
		}
		service := l[1]
		dialer := utl.Dialer{}
		conn, err := dialer.Dial(service)
		if err != nil {
			log.Printf("%v, dial fail, err:%v, service:%v", fn, err, service)
			return
		}
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		var resp []byte
		err = conn.Invoke(ctx, method, nil, &resp)
		if err != nil {
			log.Printf("%v, invoke %v fail, err:%v", fn, method, err)
			return
		}
		w.sender <- &Msg{
			Type:       TypeMsgResponse,
			Seq:        msg.Seq,
			IsNeedResp: false,
			State:      0,
			Method:     msg.Method,
			Data:       resp,
		}

	case TypeMsgResponse:
		w.callerMu.RLock()
		defer w.callerMu.RUnlock()
		if respC, ok := w.caller[msg.Seq]; ok {
			respC <- msg
		} else {
			log.Printf("dispatch, resp cannot find caller, seq:%v", msg.Seq)
		}
		// set record
	}
}

// Call add `wg.add(1)` and `defer wg.done()`, if you use `go w.call()`
/*
w.wg.add(1)
msg := w.Call(msg)
w.wg.done()
*/
func (w *websocketWorker) Call(msg *Msg) *Msg {
	msg.Seq = w.increaseSeq()
	w.sender <- msg
	if !msg.IsNeedResp {
		return nil
	}
	waitChan := make(chan *Msg, 1)
	w.callerMu.Lock()
	w.caller[msg.Seq] = waitChan
	w.callerMu.Unlock()
	resp := <-waitChan
	w.callerMu.Lock()
	delete(w.caller, msg.Seq)
	w.callerMu.Unlock()
	close(waitChan)
	return resp
}

func (w *websocketWorker) increaseSeq() uint32 {
	return w.seq.Add(1)
}

func (w *websocketServer) BroadcastMsg(userIds []string, msg *Msg) error {
	/*
		get connIds by userIds
	*/
	var worker []*websocketWorker
	for _, r := range worker {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		select {
		case r.receiver <- msg:
		case <-ctx.Done():
		}
		cancel()
	}
	return nil
}
