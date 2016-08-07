package featherbyte_test

import fb "github.com/jarro2783/featherbyte"
import "testing"
import "fmt"

// The protocol is simple, each packet consists of a single byte indicating
// the type of the packet, along with the data for that packet.
// There are some predefined packet types, and then users are free to define
// their own for other functionality.

// 0 | OK
// 1 | Hello

// The handshake is a Hello with an OK reply.

func TestHello(t *testing.T) {
    port := 35123
    address := fmt.Sprintf("%s:%d", "localhost", port)
    server := fb.NewServer("tcp", address)
    client := fb.NewClient("tcp", address)

    var err error

    err = server.Listen()

    if err != nil {
        fmt.Printf("Unable to listen\n")
        t.FailNow()
    }

    go func (s *fb.Server) {
        s.Accept()
    }(server)

    err = client.Connect()

    if err != nil {
        fmt.Printf("Unable to connect: %s\n", err.Error())
        t.FailNow()
    }

    if !client.Connected() {
        fmt.Printf("Hello not set correctly\n")
        t.FailNow()
    }
}
