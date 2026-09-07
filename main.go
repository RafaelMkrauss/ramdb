package main

import "ramdb/server"

type CustomResponse struct{}

func (CustomResponse) Ok() error {
	return nil
}

func main() {
	serv := server.Server{
		ListenAddress: ":8000",
	}
	serv.StartServing(func(r server.Request) server.Response {
		return CustomResponse{}
	})
}
