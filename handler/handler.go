package handler

import (
	"fmt"
	"strings"

	"ramdb/aof"
	"ramdb/db"
	"ramdb/server"
)

// StringResponse implementa a interface server.Response do seu colega
type StringResponse struct {
	Data  string
	Error error
}

func (r StringResponse) Fail() error {
	return r.Error
}

type CommandHandler struct {
	db *db.Engine
	aof *aof.AOF
}

func New(db *db.Engine, a *aof.AOF) *CommandHandler {
	return &CommandHandler{
		db: db,
		aof: a,
	}
}

// método para o TCP
func (h *CommandHandler) Handle(r server.Request) server.Response {
	payload := strings.TrimSpace(string(r.RequestBody))
	args := strings.Split(payload, " ")

	if len(args) == 0 || args[0] == "" {
		return StringResponse{Error: fmt.Errorf("comando vazio")}
	}

	comando := strings.ToUpper(args[0])

	switch comando {
	case "SET":
		if len(args) < 3 {
			return StringResponse{Error: fmt.Errorf("uso correto: SET <chave> <valor>")}
		}
		chave := args[1]
		valor := strings.Join(args[2:], " ")
		err := h.db.Set(chave, valor)
		if err != nil {
			return StringResponse{Error: err}
		}
		h.db.Set(chave, valor) //Não duplicaram isso aqui sem querer não?
		if h.aof != nil {
			h.aof.Append(payload + "\n") 
		}
		return StringResponse{Data: "OK"}

	case "GET":
		if len(args) < 2 {
			return StringResponse{Error: fmt.Errorf("uso correto: GET <chave>")}
		}
		chave := args[1]
		valor, err := h.db.Get(chave)
		if err != nil {
			return StringResponse{Error: err}
		}
		return StringResponse{Data: valor}

	case "DEL":
		if len(args) < 2 {
			return StringResponse{Error: fmt.Errorf("uso correto: DEL <chave>")}
		}
		chave := args[1]
		h.db.Delete(chave)

		if h.aof != nil {
			h.aof.Append(payload + "\n") 
		}
		
		return StringResponse{Data: "OK"}

	default:
		return StringResponse{Error: fmt.Errorf("comando desconhecido: %s", comando)}
	}
}
