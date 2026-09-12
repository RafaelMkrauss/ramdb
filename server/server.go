package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"ramdb/constants"
	bfreader "ramdb/utils/buffered_reader"
	bufferedreader "ramdb/utils/buffered_reader"
	"time"
)

type Server struct {
	ListenAddress         string
	RequestHandleCallback func(Request) Response
}

type Request struct {
	RequestId   uint32
	RequestBody []byte
}

type Response interface {
	Fail() error // Indicates if  the request could be resolved correctly
}

func encodeResponse(Response) []byte {
	return []byte{}
}

func gotoMessageStart(reader *bfreader.BufferedReader[byte]) error {
	padding_count := 0
	char_buffer := []byte{0}
	for {
		_, err := reader.Read(char_buffer[:1])
		if err != nil {
			return err
		}
		if char_buffer[0] == constants.SOH {
			if padding_count == 3 {
				return nil
			}
			padding_count += 1
		}
	}
}

func parseRequest(reader *bfreader.BufferedReader[byte]) (Request, error) {
	err := gotoMessageStart(reader)
	if err != nil {
		return Request{}, err
	}
	char_buffer := make([]byte, 8)
	n, err := reader.Read(char_buffer)
	if err != nil {
		return Request{}, err
	} else if n != 8 {
		return Request{}, fmt.Errorf("Não foi possível ler o tamanho da mensagem.")
	}
	msg_length := binary.LittleEndian.Uint32(char_buffer[:4])
	request_id := binary.LittleEndian.Uint32(char_buffer[4:])
	msg_body_buffer := make([]byte, msg_length)
	n, err = reader.Read(char_buffer) // isso aq vai explodir,n?
	if err != nil {
		return Request{}, err
	} else if n != 8 {
		return Request{}, fmt.Errorf("Não foi possível ler o tamanho da mensagem.")
	} else {
		return Request{
			RequestId:   request_id,
			RequestBody: msg_body_buffer[:n],
		}, nil
	}
}

func (serv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	reader := bufferedreader.New[byte](conn, 1024)
	for {
		decoded_request, err := parseRequest(&reader)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF) {
				break
			}
			log.Printf("Error in message decription: %s", err.Error())
		}
		response := encodeResponse(serv.RequestHandleCallback(decoded_request))
		n, err := conn.Write(response)
		if err != nil || n != len(response) {
			log.Printf("Error in sending response: %v.", err)
		}
	}
}

func (serv *Server) StartServing() error {
	listener, err := net.Listen("tcp", serv.ListenAddress)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting to connection: %v.", err)
		}
		go serv.handleConnection(conn)
	}
}
