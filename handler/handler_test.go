package handler

import (
	"ramdb/db"
	"ramdb/server"
	"testing"
)

func TestCommandHandler(t *testing.T) {
	// 1. Setup: Instancia o DB e o Handler 
	// (Passamos nil para o AOF para focar apenas na lógica em memória neste teste)
	database := db.NewEngine()
	cmdHandler := New(database, nil)

	// 2. Definimos os cenários que queremos testar sem usar o Netcat
	cenarios := []struct {
		nome           string
		comando        string
		esperaErro     bool
		respostaAguard string
	}{
		{"Comando SET válido", "SET heroi batman", false, "OK"},
		{"Comando GET válido", "GET heroi", false, "batman"},
		{"GET em chave inexistente", "GET vilao", true, ""},
		{"Falta de argumentos no SET", "SET vazio", true, ""},
		{"Comando DEL válido", "DEL heroi", false, "OK"},
	}

	// 3. O loop processa as requisições simuladas
	for _, tc := range cenarios {
		t.Run(tc.nome, func(t *testing.T) {
			// Simula o que o TCP entregaria
			req := server.Request{
				RequestBody: []byte(tc.comando),
			}

			// Injeta no handler
			resp := cmdHandler.Handle(req)

			// Converte a interface de volta para StringResponse para ler os dados
			strResp, ok := resp.(StringResponse)
			if !ok {
				t.Fatalf("O handler não retornou uma StringResponse")
			}

			// Validações
			if tc.esperaErro && strResp.Error == nil {
				t.Errorf("Esperava falha, mas o comando funcionou")
			}
			if !tc.esperaErro && strResp.Error != nil {
				t.Errorf("Comando falhou inesperadamente: %v", strResp.Error)
			}
			if strResp.Data != tc.respostaAguard {
				t.Errorf("Esperava '%s', recebeu '%s'", tc.respostaAguard, strResp.Data)
			}
		})
	}
}