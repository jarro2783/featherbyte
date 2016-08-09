package featherbyte_test

import fb "github.com/jarro2783/featherbyte"
import "fmt"
import "testing"
import "time"

// The protocol is simple, each packet consists of a single byte indicating
// the type of the packet, along with the data for that packet.
// There are some predefined packet types, and then users are free to define
// their own for other functionality.

// 0 | Hello
// 1 | OK

// The handshake is a Hello with an OK reply.

const (
    json = fb.UserMessageStart
    str = iota
)

type handler struct {
}

func slicesEqual(a []byte, b []byte) bool {
    if len(a) != len(b) {
        return false
    }

    for i := 0; i != len(a); i++ {
        if a[i] != b[i] {
            return false
        }
    }

    return true
}

func (h *handler) Connection(ep *fb.Endpoint) {
    fmt.Printf("Started handler\n")
}

type reader struct {
    expected [][]byte
    got int
    *testing.T

    expectedMessages [][]byte
    messages int
}

func makeReader(t *testing.T) *reader {
    r := new(reader)
    r.got = 0
    r.expected = make([][]byte, 0, 5)
    r.T = t

    r.expectedMessages = make([][]byte, 0, 5)
    r.messages = 0

    return r
}

func (r *reader) addExpected(expected []byte) {
    r.expected = append(r.expected, expected)
}

func (r *reader) addMessage(message []byte) {
    r.expectedMessages = append(r.expectedMessages, message)
}

func (r *reader) Data(messageType byte, data []byte) {
    fmt.Printf("Got message: %d\n", messageType)

    if slicesEqual(data, r.expected[r.got]) {
        r.got++
    } else {
        r.T.Error(data, " != ", r.expected[r.got])
    }
}

func (r* reader) Message(messageType byte, data []byte) {
    if slicesEqual(data, r.expectedMessages[r.messages]) {
        r.messages++
    } else {
        r.T.Error(data, " != ", r.expectedMessages[r.messages])
    }
}

func makeAddress() string {
    return "localhost:34123"
}

func sleep() {
    time.Sleep(time.Millisecond * 100)
}

func makeLong() []byte {
    b := make([]byte, 1024)

    for i := 0; i != len(b); i++ {
        b[i] = byte(i)
    }

    return b
}

func TestHello(t *testing.T) {
    address := makeAddress()

    short := []byte{0, 1, 2, 3}
    long := makeLong()
    message := []byte("{\"a\" : 5}")

    serverRead := makeReader(t)
    serverRead.addExpected(short)
    serverRead.addExpected(long)
    serverRead.addMessage(message)

    go func () {
        fmt.Printf("Starting server\n")
        fb.Listen("tcp", address, new(handler), serverRead)
    }()

    sleep()

    fmt.Printf("Connecting\n")
    ep, err := fb.Connect("tcp", address, new(reader))

    defer ep.Close()

    if err != nil {
        fmt.Printf("Unable to connect: %s\n", err.Error())
        t.FailNow()
    }

    if !ep.Connected() {
        fmt.Printf("Hello not set correctly\n")
        t.FailNow()
    }

    ep.WriteBytes(short)
    sleep()
    ep.WriteBytes(long)
    sleep()

    ep.WriteMessage(json, message)

    sleep()

    if serverRead.got != len(serverRead.expected) {
        t.Errorf("Expected %d byte messages, got %d\n",
            len(serverRead.expected),
            serverRead.got)
    }

    if serverRead.messages != len(serverRead.expectedMessages) {
        t.Errorf("Expected %d text messages, got %d\n",
            len(serverRead.expectedMessages),
            serverRead.messages)
    }
}
