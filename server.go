package featherbyte

import "net"

type Server struct {
    protocol string
    address string
    listener net.Listener
}

func NewServer(protocol string, address string) *Server {
    server := new(Server)

    server.protocol = protocol
    server.address = address

    return server
}

func (server *Server) Accept() error {
    conn, err := server.listener.Accept()

    if err != nil {
        return err
    }

    data := make([]byte, 1, 1)

    _, err = conn.Read(data)

    if err != nil {
        return err
    }

    if data[0] == 0 {
        data[0] = 1
        conn.Write(data)
    }

    return nil
}
