package aof

import (
	"os"
	"strings"
	"testing"
)

func TestAOF_AppendAndClose(t *testing.T) {
	filename := "test_database.aof"
	
	// Limpa arquivo de teste antes de rodar
	os.Remove(filename)
	defer os.Remove(filename) // Limpa no final

	a, err := NewAOF(filename)
	if err != nil {
		t.Fatalf("Falha ao criar AOF: %v", err)
	}

	// Simula comandos do Redis (protocolo RESP)
	cmds := []string{
		"*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$4\r\nval1\r\n",
		"*3\r\n$3\r\nSET\r\n$4\r\nkey2\r\n$4\r\nval2\r\n",
	}

	for _, cmd := range cmds {
		a.Append(cmd)
	}

	// Fecha para forçar a sincronização de tudo
	a.Close()

	// Lê o arquivo do disco para conferir
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Falha ao ler o arquivo AOF: %v", err)
	}

	strContent := string(content)
	for _, cmd := range cmds {
		if !strings.Contains(strContent, cmd) {
			t.Errorf("Esperava encontrar o comando %q no arquivo, mas não encontrou", cmd)
		}
	}
}