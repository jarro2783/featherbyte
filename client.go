package featherbyte

import (
  "net"
)

type Client struct {
    protocol string
    address string
    connection net.Conn
    isConnected bool
}

func NewClient(protocol string, address string) *Client {
    client := new(Client)

    client.protocol = protocol
    client.address = address
    client.isConnected = false

    return client
}

func (client *Client) Connected() bool {
    return client.isConnected
}

