package rpc

import (
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	pb "code.aliyun.com/chain33/chain33/types"
	"github.com/rs/cors"

	"google.golang.org/grpc"
)

// adapt HTTP connection to ReadWriteCloser
type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

func (j *JsonRpcServer) Listen(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Crit("listen:", "err", err)
		panic(err)
	}
	j.Listener = listener
	server := rpc.NewServer()

	server.Register(&j.jrpc)
	co := cors.New(cors.Options{})

	// Insert the middleware
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			serverCodec := jsonrpc.NewServerCodec(&HttpConn{in: r.Body, out: w})
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(200)
			err := server.ServeRequest(serverCodec)
			if err != nil {
				log.Debug("Error while serving JSON request: %v", err)
				return
			}
		}
	})

	handler = co.Handler(handler)
	http.Serve(j.Listener, handler)
}

func (g *Grpcserver) Listen(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Crit("failed to listen:", "err", err)
		panic(err)
	}
	g.Listener = listener
	s := grpc.NewServer()
	pb.RegisterGrpcserviceServer(s, &g.grpc)
	s.Serve(g.Listener)
}
