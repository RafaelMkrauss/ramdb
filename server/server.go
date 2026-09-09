package server

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type Server struct {
	ListenAddress string
}

type Request struct {
	RawRequestContent []byte
}

type Response interface {
	Ok() error // Indicates if  the request could be resolved correctly
}

func encodeResponse(Response) []byte {
	return []byte{}
}

func getRequestsFullContent(conn net.Conn) ([]byte, error) {
	mgs_content := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	tot_read_size, err := conn.Read(buf)
	if err != nil {
		return []byte{}, err
	}
	if tot_read_size < 4 {
		return mgs_content, fmt.Errorf("Request size less than 4 bytes")
	}
	message_size := int(binary.LittleEndian.Uint32(buf[:4]))
	mgs_content = append(mgs_content, buf[4:tot_read_size]...)

	for tot_read_size < message_size {
		nbytes, err := conn.Read(buf)
		if err != nil {
			return mgs_content, err
		}
		mgs_content = append(mgs_content, buf[:nbytes]...)
		tot_read_size += nbytes
	}
	return mgs_content, nil
}

func parseRequest(conn net.Conn) (Request, error) {
	msg, err := getRequestsFullContent(conn)
	if err != nil {
		return Request{RawRequestContent: msg}, err
	}
	return Request{RawRequestContent: msg}, nil
}

func (serv *Server) StartServing(request_handle_callback func(Request) Response) error {
	listener, err := net.Listen("tcp", serv.ListenAddress)
	if err != nil {
		return fmt.Errorf("Error listening to address %s: %v.", serv.ListenAddress, err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting to connection: %v.", err)
		}
		conn.SetReadDeadline(time.Now().Add(time.Second * 3)) // deadline de segurança
		go func() {
			defer conn.Close()
			decoded_request, err := parseRequest(conn)
			if err != nil {
				log.Printf("Error in message decription.")
			}
			response := encodeResponse(request_handle_callback(decoded_request))
			n, err := conn.Write(response)
			if err != nil || n != len(response) {
				log.Printf("Error in sending response: %v.", err)
			}
		}()
	}
}
