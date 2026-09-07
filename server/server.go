package server

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	ListenAddress string
}

type Request interface{}

type Response interface {
	Ok() error // Indicates if  the request could be resolved correctly
}

func encodeResponse(Response) []byte {
	return []byte{}
}

func parseRequest(conn net.Conn) Request {
	buffer := make([]byte, 1024)
	conn.Read(buffer)
	fmt.Printf("%s", string(buffer))
	return ""
}

func (serv *Server) StartServing(request_handle_callback func(Request) Response) error {
	listener, err := net.Listen("tcp", serv.ListenAddress)
	if err != nil {
		return fmt.Errorf("Error listening to address %s: %v", serv.ListenAddress, err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error listening to connection: %v", err)
		}
		go func() {
			defer conn.Close()
			response := encodeResponse(request_handle_callback(parseRequest(conn)))
			n, err := conn.Write(response)
			if err != nil || n != len(response) {
				log.Printf("Error in sending response: %v", err)
			}
		}()
	}
}
