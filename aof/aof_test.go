package aof

import (
	"os"
	"strings"
	"testing"
)

func TestAOF_AppendAndClose(t *testing.T) {
	filename := "test_database.aof"
	
	os.Remove(filename) //limpa antes de rodar
	defer os.Remove(filename) //limpa depois de rodar

	a, err := NewAOF(filename)
	if err != nil {
		t.Fatalf("Falha ao criar AOF: %v", err)
	}

	//simulando sets
	cmds := []string{
		"*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$4\r\nval1\r\n",
		"*3\r\n$3\r\nSET\r\n$4\r\nkey2\r\n$4\r\nval2\r\n",
	}

	for _, cmd := range cmds {
		a.Append(cmd)
	}
	a.Close()

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