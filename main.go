package main

import "ramdb/server"

type CustomResponse struct{}

func (CustomResponse) Fail() error {
	return nil
}

func main() {
	serv := server.Server{
		ListenAddress: ":8000",
		RequestHandleCallback: func(r server.Request) server.Response {
			return CustomResponse{}
		},
	}
	serv.StartServing()
}
