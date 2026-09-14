package integration_test

import (
	"fmt"
	"log"
	"net"
	"ramdb/aof"
	"ramdb/db"
	"ramdb/handler"
	"ramdb/server"
	"testing"
	"time"
)

func setupServer(handler *handler.CommandHandler) {

	serv := server.Server{
		ListenAddress: ":8000",
		RequestHandleCallback: func(r server.Request) server.Response {

			return handler.Handle(r)

			//return CustomResponse{}
		},
	}
	serv.StartServing()
}

func TestMessageExchange(t *testing.T) {
	persistencia, err := aof.NewAOF("database.aof")
	if err != nil {
		log.Fatal("Erro ao iniciar AOF:", err)
	}
	defer persistencia.Close()
	db := db.NewEngine()
	cmdHandler := handler.New(db, persistencia)

	go setupServer(cmdHandler)

	comandosSimulados := []string{
		"SET linguagem go",
		"GET linguagem",
		"DEL linguagem",
		"GET linguagem", //deve retornar erro
	}
	time.Sleep(time.Second * 2)

	for _, payload := range comandosSimulados {
		request := []byte{1, 1, 1, 1, byte(len(payload)), 0, 0, 0, 0, 1, 0, 0}
		request = append(request, []byte(payload)...)

		conn, err := net.Dial("tcp", ":8000")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		_, err = conn.Write(request)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("\n%s", err.Error())
		}
		fmt.Printf("%s\n", buffer[:n])
	}
}
