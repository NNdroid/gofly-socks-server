package reality

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/xtls/reality"
	"go.uber.org/zap"
	"gofly/pkg/cipher"
	"gofly/pkg/config"
	"gofly/pkg/logger"
	"gofly/pkg/protocol/basic"
	"gofly/pkg/utils"
	"gofly/pkg/x/xproto"
	"net"
	"time"
)

type ServerListener struct {
	net.Listener
}

type ServerConfig config.RealityConfig

func (s *ServerConfig) ShortIDMap() (map[[8]byte]bool, error) {
	maps := make(map[[8]byte]bool, len(s.ShortID))

	for _, v := range s.ShortID {
		var id [8]byte
		length, err := hex.Decode(id[:], []byte(v))
		if err != nil {
			return nil, fmt.Errorf("decode hex failed: %w", err)
		}

		if length > 8 {
			return nil, fmt.Errorf("short id length is large than 8")
		}

		maps[id] = true
	}

	return maps, nil
}

func (s *ServerConfig) ServerNameMap() map[string]bool {
	maps := make(map[string]bool, len(s.ServerNames))

	for _, v := range s.ServerNames {
		maps[v] = true
	}

	return maps
}

func NewServer(lis net.Listener, config *ServerConfig) (*ServerListener, error) {
	ids, err := config.ShortIDMap()
	if err != nil {
		return nil, err
	}
	privateKey, err := base64.RawURLEncoding.DecodeString(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("decode private_key failed: %w", err)
	}
	return &ServerListener{
		reality.NewListener(lis, &reality.Config{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial(network, address)
			},
			Show:                   config.Debug,
			Type:                   "tcp",
			ShortIds:               ids,
			ServerNames:            config.ServerNameMap(),
			Dest:                   config.Dest,
			PrivateKey:             privateKey,
			SessionTicketsDisabled: true,
		}),
	}, nil
}

type Server struct {
	basic.Server
}

// StartServerForApi starts the tcp server
func (x *Server) StartServerForApi() {
	//serverConfig := &ServerConfig{
	//	ShortID:     []string{"abcd"},
	//	ServerNames: []string{"gkreg.rk.gov.ru"},
	//	Dest:        "193.47.166.43:443",
	//	PrivateKey:  "eLW3EAsrdEyrVj0hru6QpkzZjerKDVROiXHdZsmEKnw",
	//	Debug:       x.Config.VTun.Verbose,
	//}
	x.ConnectionCache = cache.New(15*time.Minute, 24*time.Hour)
	cipher.SetKey(x.Config.VTunSettings.Key)
	listener, err := net.Listen("tcp", x.Config.VTunSettings.LocalAddr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	serverConfig := ServerConfig(x.Config.RealitySettings)
	server, err := NewServer(listener, &serverConfig)
	if err != nil {
		panic(err)
	}
	defer server.Listener.Close()
	logger.Logger.Sugar().Infof("gofly %s server started on %v", x.Config.VTunSettings.Protocol, x.Config.VTunSettings.LocalAddr)
	// server -> client
	go x.ToClient()
	// client -> server
	for basic.ContextOpened(x.CTX) {
		conn, err := server.Listener.Accept()
		if err != nil {
			logger.Logger.Sugar().Errorf("accept error, %v\n", err)
			continue
		}
		x.Statistics.Push(conn.RemoteAddr())
		logger.Logger.Sugar().Debugf("accept connect: %s", conn.RemoteAddr().String())
		err = x.HandshakeFromClient(conn, x.AuthKey())
		if err != nil {
			x.closeTheClient(conn, errors.New("active shutdown"))
			logger.Logger.Sugar().Errorf("error, %v\n", err)
			continue
		}
		go x.ToServer(conn)
	}
}

// ToClient sends packets from iFace to conn
func (x *Server) ToClient() {
	buffer := make([]byte, x.Config.VTunSettings.BufferSize)
	var n int
	var err error
	var ns int
	for basic.ContextOpened(x.CTX) {
		n, err = x.ReadFunc(buffer)
		if err != nil {
			logger.Logger.Sugar().Errorf("error, %v\n", err)
			continue
		}
		b := buffer[:n]
		x.ConvertDstAddr(b)
		if key := utils.GetDstKey(b); key != "" {
			if v, ok := x.ConnectionCache.Get(key); ok {
				x.ConnectionCache.Set(key, v, 15*time.Minute)
				conn := v.(net.Conn)
				b, err = x.ExtendEncode(b)
				if err != nil {
					logger.Logger.Sugar().Errorf("encode error, %v\n", err)
					x.closeTheClient(conn, err)
					continue
				}
				ph := &xproto.ServerSendPacketHeader{
					ProtocolVersion: xproto.ProtocolVersion,
					Length:          len(b),
				}
				ns, err = conn.Write(xproto.Merge(ph.Bytes(), b))
				if err != nil {
					logger.Logger.Sugar().Errorf("error, %v\n", err)
					x.ConnectionCache.Delete(key)
					x.closeTheClient(conn, err)
					continue
				}
				x.Statistics.IncrTransportBytes(ns)
				x.Statistics.IncrClientReceivedBytes(conn.RemoteAddr(), ns)
			} else if v, _, ok := x.ConnectionCache.GetWithExpiration(key); ok {
				x.closeTheClient(v.(net.Conn), errors.New("active shutdown, cache was expired"))
				x.ConnectionCache.Delete(key)
			}
		}
	}
}

func (x *Server) HandshakeFromClient(conn net.Conn, authKey *xproto.AuthKey) error {
	handshake := make([]byte, xproto.ClientHandshakePacketLength)
	n, err := conn.Read(handshake)
	if err != nil {
		return fmt.Errorf("error, %v\n", err)
	}
	if n != xproto.ClientHandshakePacketLength {
		return fmt.Errorf("received handshake length <%d> not equals <%d>!\n", n, xproto.ClientHandshakePacketLength)
	}
	hs := xproto.ParseClientHandshakePacket(handshake[:n])
	if hs == nil {
		return fmt.Errorf("hs == nil")
	}
	if !hs.Key.Equals(authKey) {
		return fmt.Errorf("authentication failed")
	}
	x.ConnectionCache.Set(hs.CIDRv4.String(), conn, 15*time.Minute)
	x.ConnectionCache.Set(hs.CIDRv6.String(), conn, 15*time.Minute)
	return nil
}

// ToServer sends packets from conn to iFace
func (x *Server) ToServer(conn net.Conn) {
	defer x.closeTheClient(conn, errors.New("active shutdown"))
	header := make([]byte, xproto.ClientSendPacketHeaderLength)
	packet := make([]byte, x.Config.VTunSettings.BufferSize)
	var err error
	var total int
	var length int
	var n int
	for basic.ContextOpened(x.CTX) {
		total = 0
		n, err = splitRead(conn, xproto.ClientSendPacketHeaderLength, header)
		if err != nil {
			logger.Logger.Sugar().Errorf("error, %v\n", err)
			break
		}
		if n != xproto.ClientSendPacketHeaderLength {
			logger.Logger.Sugar().Errorf("received length <%d> not equals <%d>!", n, xproto.ClientSendPacketHeaderLength)
			break
		}
		total += n
		ph := xproto.ParseClientSendPacketHeader(header[:n])
		if ph == nil {
			logger.Logger.Sugar().Errorln("ph == nil")
			break
		}
		if !ph.Key.Equals(x.AuthKey()) {
			logger.Logger.Sugar().Errorln("authentication failed")
			break
		}
		length, err = splitRead(conn, ph.Length, packet[:ph.Length])
		if err != nil {
			logger.Logger.Sugar().Errorf("error, %v\n", err)
			break
		}
		if length != ph.Length {
			logger.Logger.Sugar().Errorf("received length <%d> not equals <%d>!", n, ph.Length)
			break
		}
		total += length
		b := packet[:length]
		b, err = x.ExtendDecode(b)
		if err != nil {
			logger.Logger.Sugar().Errorf("decode error, %v\n", err)
			break
		}
		if dstKey := utils.GetDstKey(b); dstKey != "" {
			if v, ok := x.ConnectionCache.Get(dstKey); ok && !x.Config.VTunSettings.ClientIsolation {
				dstConn := v.(net.Conn)
				_, err = dstConn.Write(b)
				if err != nil {
					logger.Logger.Sugar().Errorf("error, %v\n", err)
					x.closeTheClient(dstConn, err)
					continue
				}
			} else {
				x.ConvertSrcAddr(b)
				x.WriteFunc(b)
				x.Statistics.IncrReceivedBytes(total)
				x.Statistics.IncrClientTransportBytes(conn.RemoteAddr(), total)
			}
		}
	}
}

func (x *Server) closeTheClient(conn net.Conn, err error) {
	x.Statistics.Remove(conn.RemoteAddr())
	defer conn.Close()
	logger.Logger.Sugar().Debugf("closed: %s -> %v", conn.RemoteAddr().String(), zap.Error(err))
}
