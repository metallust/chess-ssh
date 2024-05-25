package connector


type Msg struct {
    Name string
    Data interface{}
}

type Connector struct {
    Sender chan Msg
    Reciever chan Msg
}

func NewConnector() *Connector {
    return &Connector{
        Sender: make(chan Msg),
        Reciever: make(chan Msg),
    }
}

func CreateConnectorPair(c *Connector) *Connector {
    return &Connector{
        Sender: c.Reciever,
        Reciever: c.Sender,
    }
}

func (c *Connector) SendMsg(msg Msg) {
    c.Sender <- msg
}

func (c *Connector) GetMsg() (Msg, bool) {
    msg, more := <-c.Reciever
    return msg, more
}

func (c *Connector) Close() {
    close(c.Sender)
    close(c.Reciever)
}


