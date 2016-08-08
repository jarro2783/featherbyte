package featherbyte

import "fmt"
import "net"
import "time"

type Endpoint struct {
    connection net.Conn
    isConnected bool
}

type ConnectionHandler interface {
    Connection(ep *Endpoint)
}

type ReadData interface {
    Data(messageType byte, data []byte)
}

func Listen(
    protocol string,
    address string,
    handler ConnectionHandler,
    read ReadData) {
    listener, err := net.Listen(protocol, address)

    if err != nil {
        fmt.Printf("Unable to listen: %s\n", err.Error())
        return
    }

    for true {
        fmt.Printf("Accepting new connection\n")
        conn, err := listener.Accept()

        if err != nil {
            fmt.Printf("Error accepting: %s\n", err.Error())
            listener.Close()
            return
        }

        fmt.Printf("Accepted connection\n")

        ep := new(Endpoint)
        ep.connection = conn

        ep.startReader(read)

        go handler.Connection(ep)
    }
}

func (ep *Endpoint) startReader(read ReadData) {
    go ep.readRoutine(read)
}

func (ep *Endpoint) readRoutine(read ReadData) {
    fmt.Printf("Starting reader\n")
    conn := ep.connection

    data := make([]byte, 1024)

    var err error

    for {
        n, err := conn.Read(data);
        if err != nil {
            fmt.Printf("Error reading: %s\n", err.Error())
            break
        }
        fmt.Printf("Read %d bytes\n", n)
        switch data[0] {
            case 42:
                data[0] = 1
                conn.Write(data[0:1])
            default:
                read.Data(data[0], data[1:n])
        }
    }

    if err != nil {
        fmt.Printf("Error reading: %s\n", err.Error())
    }
}

func Connect(
    protocol string,
    address string,
    read ReadData) (*Endpoint, error) {

    fmt.Printf("Dialing\n")
    conn, err := net.Dial(protocol, address)

    if err != nil {
        fmt.Printf("Error dialing: %s\n", err.Error())
        return nil, err
    }

    fmt.Printf("Accepted\n")

    ep := new(Endpoint)

    ep.connection = conn

    err = ep.hello()

    ep.startReader(read)

    return ep, err
}

func (ep *Endpoint) hello() error {

    conn := ep.connection

    fmt.Printf("Writing hello\n")
    data := [1]byte{42}

    conn.SetDeadline(time.Now().Add(time.Second * 5))
    _, err := conn.Write(data[:])

    if err != nil {
        fmt.Printf("Error writing hello: %s\n", err.Error())
        return err
    }

    _, err = conn.Read(data[:])

    if err != nil {
        return err
    }

    if data[0] == 1 {
        ep.isConnected = true
    }

    return nil
}

func (ep* Endpoint) Connected() bool {
    return ep.isConnected
}
