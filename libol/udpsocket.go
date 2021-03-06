package libol

import (
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

type UdpConfig struct {
	Block   kcp.BlockCrypt
	Timeout time.Duration // ns
}

var defaultUdpConfig = UdpConfig{
	Timeout: 120 * time.Second,
}

type UdpServer struct {
	socketServer
	udpCfg   *UdpConfig
	listener net.Listener
}

func NewUdpServer(listen string, cfg *UdpConfig) *UdpServer {
	if cfg == nil {
		cfg = &defaultUdpConfig
	}
	k := &UdpServer{
		udpCfg: cfg,
		socketServer: socketServer{
			address:    listen,
			sts:        ServerSts{},
			maxClient:  1024,
			clients:    NewSafeStrMap(1024),
			onClients:  make(chan SocketClient, 4),
			offClients: make(chan SocketClient, 8),
		},
	}
	k.close = k.Close
	if err := k.Listen(); err != nil {
		Debug("NewUdpServer: %s", err)
	}
	return k
}

func (k *UdpServer) Listen() (err error) {
	k.listener, err = XDPListen(k.address)
	if err != nil {
		k.listener = nil
		return err
	}
	Info("UdpServer.Listen: udp://%s", k.address)
	return nil
}

func (k *UdpServer) Close() {
	if k.listener != nil {
		_ = k.listener.Close()
		Info("UdpServer.Close: %s", k.address)
		k.listener = nil
	}
}

func (k *UdpServer) Accept() {
	for {
		if k.listener != nil {
			break
		}
		if err := k.Listen(); err != nil {
			Warn("UdpServer.Accept: %s", err)
		}
		time.Sleep(time.Second * 5)
	}
	defer k.Close()
	for {
		conn, err := k.listener.Accept()
		if err != nil {
			Error("TcpServer.Accept: %s", err)
			return
		}
		k.sts.AcceptCount++
		k.onClients <- NewUdpClientFromConn(conn, k.udpCfg)
	}
}

// Client Implement

type UdpClient struct {
	socketClient
	udpCfg *UdpConfig
}

func NewUdpClient(addr string, cfg *UdpConfig) *UdpClient {
	if cfg == nil {
		cfg = &defaultUdpConfig
	}
	c := &UdpClient{
		udpCfg: cfg,
		socketClient: socketClient{
			address: addr,
			newTime: time.Now().Unix(),
			dataStream: dataStream{
				maxSize: 1514,
				minSize: 15,
				message: &DataGramMessage{
					timeout: cfg.Timeout,
					block:   cfg.Block,
				},
			},
			status: ClInit,
		},
	}
	c.connecter = c.Connect
	return c
}

func NewUdpClientFromConn(conn net.Conn, cfg *UdpConfig) *UdpClient {
	if cfg == nil {
		cfg = &defaultUdpConfig
	}
	c := &UdpClient{
		socketClient: socketClient{
			address: conn.RemoteAddr().String(),
			dataStream: dataStream{
				connection: conn,
				maxSize:    1514,
				minSize:    15,
				message: &DataGramMessage{
					timeout: cfg.Timeout,
					block:   cfg.Block,
				},
			},
			newTime: time.Now().Unix(),
		},
	}
	c.connecter = c.Connect
	return c
}

func (c *UdpClient) Connect() error {
	if !c.retry() {
		return nil
	}
	Info("UdpClient.Connect: udp://%s", c.address)
	conn, err := net.Dial("udp", c.address)
	if err != nil {
		return err
	}
	c.lock.Lock()
	c.connection = conn
	c.status = ClConnected
	c.lock.Unlock()
	if c.listener.OnConnected != nil {
		_ = c.listener.OnConnected(c)
	}
	return nil
}

func (c *UdpClient) Close() {
	c.lock.Lock()
	if c.connection != nil {
		if c.status != ClTerminal {
			c.status = ClClosed
		}
		Info("UdpClient.Close: %s", c.address)
		_ = c.connection.Close()
		c.connection = nil
		c.private = nil
		c.lock.Unlock()
		if c.listener.OnClose != nil {
			_ = c.listener.OnClose(c)
		}
	} else {
		c.lock.Unlock()
	}
}

func (c *UdpClient) Terminal() {
	c.SetStatus(ClTerminal)
	c.Close()
}

func (c *UdpClient) SetStatus(v uint8) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != v {
		if c.listener.OnStatus != nil {
			c.listener.OnStatus(c, c.status, v)
		}
		c.status = v
	}
}
