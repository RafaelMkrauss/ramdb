package main

import (
	//"fmt"
	"log"
	"ramdb/aof"
	"ramdb/db"
	"ramdb/handler"
	"ramdb/server"
)

type CustomResponse struct{}

func (CustomResponse) Fail() error {
	return nil
}

func main() {
	persistencia, err := aof.NewAOF("database.aof")
	if err != nil {
		log.Fatal("Erro ao iniciar AOF:", err)
	}
	defer persistencia.Close()

	db := db.NewEngine()
	cmdHandler := handler.New(db,persistencia)

	serv := server.Server{
		ListenAddress: ":8000",
		RequestHandleCallback: func(r server.Request) server.Response {

			return cmdHandler.Handle(r)

			//return CustomResponse{}
		},
	}
	serv.StartServing()
}

/*
	fmt.Println("teste")

	// 2. Simulamos uma requisição que acabou de sair do TCP
	req1 := server.Request{RequestBody: []byte("SET chave secreta")}
	resp1 := cmdHandler.Handle(req1)
	fmt.Printf("Comando SET: %+v\n", resp1)

	// 3. Simulamos a busca dessa mesma chave
	req2 := server.Request{RequestBody: []byte("GET chave")}
	resp2 := cmdHandler.Handle(req2)
	fmt.Printf("Comando GET: %+v\n", resp2)

	// 4. Simulamos um erro
	req3 := server.Request{RequestBody: []byte("GET nao_existe")}
	resp3 := cmdHandler.Handle(req3)
	fmt.Printf("Comando GET (Erro): %+v\n", resp3)
*/
