package featherbyte_test

import fb "github.com/jarro2783/featherbyte"
import "fmt"
import "testing"
import "time"

// The protocol is simple, each packet consists of a single byte indicating
// the type of the packet, along with the data for that packet.
// There are some predefined packet types, and then users are free to define
// their own for other functionality.

// 0 | OK
// 1 | Hello

// The handshake is a Hello with an OK reply.

type handler struct {
}

func (h *handler) Connection(ep *fb.Endpoint) {
    fmt.Printf("Started handler\n")
}

type reader struct {
}

func (r *reader) Data(messageType byte, data []byte) {
    fmt.Printf("Got message: %d\n", messageType)
}

func TestHello(t *testing.T) {
    port := 35123
    address := fmt.Sprintf("%s:%d", "localhost", port)

    go func () {
        fmt.Printf("Starting server\n")
        fb.Listen("tcp", address, new(handler), new(reader))
    }()

    time.Sleep(time.Millisecond * 100)

    fmt.Printf("Connecting\n")
    ep, err := fb.Connect("tcp", address, new(reader))

    if err != nil {
        fmt.Printf("Unable to connect: %s\n", err.Error())
        t.FailNow()
    }

    if !ep.Connected() {
        fmt.Printf("Hello not set correctly\n")
        t.FailNow()
    }
}
