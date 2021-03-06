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
    Message(messageType byte, data []byte)
    Exiting()
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
    handler ConnectionHandler) error {
    listener, err := net.Listen(protocol, address)

    if err != nil {
        return err
    }

    for true {
        conn, err := listener.Accept()

        if err != nil {
            listener.Close()
            return err
        }

        ep := new(Endpoint)
        ep.connection = conn

        go handler.Connection(ep)
    }

    return nil
}

func (ep *Endpoint) StartReader(read ReadData) {
    go ep.readPacket(read)
}

func (ep *Endpoint) readPacket(read ReadData) {

    Loop:
    for true {
        //ep.connection.SetReadDeadline(time.Now().Add(time.Second * 5))
        var length uint16
        message, err := ep.readBytes(1)

        if err != nil {
            switch e := err.(type) {
            case net.Error:
                if e.Timeout() {
                    fmt.Printf("Read timedout\n")
                    continue
                }
            default:
                break Loop
            }
        }

        if len(message) > 0 {
        } else {
            break
        }

        switch message[0] {
            case hello:
                data := []byte{ok}
                ep.connection.Write(data[0:1])
                continue
            case ok:
                continue
            case shortBytes:
                lengthBytes, err := ep.readBytes(1)
                length = uint16(lengthBytes[0])

                if err != nil {
                    break
                }
            default:
                lengthBytes, err := ep.readBytes(2)
                buffer := bytes.NewBuffer(lengthBytes)
                binary.Read(buffer, binary.BigEndian, &length)

                if err != nil {
                    break Loop
                }
        }

        data, err := ep.readBytes(int(length))

        if err != nil {
            break
        }

        switch message[0] {
            case shortBytes:
            fallthrough
            case longBytes:
            read.Data(message[0], data)

            default:
            read.Message(message[0], data)
        }

    }

    read.Exiting()
}

func (ep *Endpoint) readBytes(length int) ([]byte, error) {
    data := make([]byte, length)

    n, err := ep.connection.Read(data)

    return data[0:n], err
}

func (ep *Endpoint) readRoutine(read ReadData) {
    conn := ep.connection

    data := make([]byte, 2056)

    for {
        _, err := conn.Read(data);
        if err != nil {
            break
        }
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
                if data[0] == longBytes {
                    read.Data(data[0], data[3:3+length])
                } else {
                    read.Message(data[0], data[3:3+length])
                }
        }
    }
}

func Connect(
    protocol string,
    address string,
    read ReadData) (*Endpoint, error) {

    conn, err := net.Dial(protocol, address)

    if err != nil {
        return nil, err
    }

    ep := new(Endpoint)

    ep.connection = conn

    err = ep.hello()

    ep.StartReader(read)

    return ep, err
}

func (ep *Endpoint) hello() error {

    conn := ep.connection

    data := [1]byte{hello}

    conn.SetDeadline(time.Now().Add(time.Second * 5))

    defer conn.SetDeadline(time.Time{})
    _, err := conn.Write(data[:])

    if err != nil {
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

func (ep* Endpoint) writeLongMessage(
    messageType byte,
    data []byte,
    length int) error {
    towrite := make([]byte, 3, length + 3)
    towrite[0] = messageType

    buffer := new(bytes.Buffer)
    binary.Write(buffer, binary.BigEndian, uint16(length))
    copy(towrite[1:3], buffer.Bytes())
    towrite = append(towrite, data...)
    _, err := ep.connection.Write(towrite)

    return err
}

func (ep* Endpoint) WriteBytes(data []byte) error {

    length := len(data)

    var err error

    if length <= 255 {
        towrite := make([]byte, 2, length + 2)
        towrite[0] = shortBytes
        towrite[1] = byte(len(data))
        towrite = append(towrite, data...)
        _, err = ep.connection.Write(towrite)
    } else {
        err = ep.writeLongMessage(longBytes, data, length)
    }

    return err
}

func (ep* Endpoint) WriteMessage(messageType byte, data []byte) error {
    return ep.writeLongMessage(messageType, data, len(data))
}
