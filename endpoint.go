package featherbyte

type Endpoint struct {
    connection Conn
    isConnected bool
}

func CreateEndpoint() (*Endpoint, error) {
    e := new(Endpoint)
    return e
}

func (ep *Endpoint) Listen(protocol string, address string) {
}

func (ep *Endpoint) Connect(protocol string, address string) error {
    conn, err := net.Dial(protocol, address)

    ep.connection = conn

    err = ep.hello()
    return err
}

func (ep *Endpoint) hello() error {

    conn := ep.connection

    data := [1]byte{0}
    _, err := conn.Write(data[:])

    if err != nil {
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
