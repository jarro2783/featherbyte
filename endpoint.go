package featherbyte

import "encoding/binary"
import "bytes"
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

const (
    hello byte = 0
    ok byte = iota
    closing byte = iota
    shortBytes byte = iota
    longBytes byte = iota
    UserMessageStart byte = iota
)

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

    data := make([]byte, 2056)

    var err error

    for {
        n, err := conn.Read(data);
        if err != nil {
            fmt.Printf("Error reading: %s\n", err.Error())
            break
        }
        fmt.Printf("Read %d bytes\n", n)
        switch data[0] {
            case hello:
                data[0] = ok
                conn.Write(data[0:1])
            case shortBytes:
                length := data[1]
                read.Data(data[0], data[2:2+length])
            //longBytes and other messages are the same
            default:
                var length uint16
                buffer := bytes.NewBuffer(data[1:3])
                binary.Read(buffer, binary.BigEndian, &length)
                fmt.Printf("Packet length: %d\n", length)
                if data[0] == longBytes {
                    read.Data(data[0], data[3:3+length])
                }
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
    data := [1]byte{hello}

    conn.SetDeadline(time.Now().Add(time.Second * 5))
    _, err := conn.Write(data[:])

    if err != nil {
        fmt.Printf("Error writing hello: %s\n", err.Error())
        return err
    }

    _, err = conn.Read(data[0:1])

    if err != nil {
        return err
    }

    if data[0] == ok {
        ep.isConnected = true
    }

    return nil
}

func (ep* Endpoint) Connected() bool {
    return ep.isConnected
}

func (ep* Endpoint) Close() {
    ep.connection.Close()
}

func (ep* Endpoint) WriteBytes(data []byte) {

    var towrite []byte

    length := len(data)

    if length <= 255 {
        towrite = make([]byte, 2, length + 2)
        towrite[0] = shortBytes
        towrite[1] = byte(len(data))
        towrite = append(towrite, data...)
    } else {
        towrite = make([]byte, 3, length + 3)
        towrite[0] = longBytes

        buffer := new(bytes.Buffer)
        binary.Write(buffer, binary.BigEndian, uint16(length))
        copy(towrite[1:3], buffer.Bytes())
        towrite = append(towrite, data...)
    }

    ep.connection.Write(towrite)
}

func (ep* Endpoint) WriteMessage(messageType byte, data []byte) {
}
