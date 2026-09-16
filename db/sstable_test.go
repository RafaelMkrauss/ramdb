package db

import (
	"bytes"
	"os"
	"testing"
)

func TestSSTable_WriteAndRead(t *testing.T) {
	filename := "test_sstable_9999.data"

	// Limpa o arquivo de teste antes e depois de rodar
	os.Remove(filename)
	defer os.Remove(filename)

	// 1. Prepara dados falsos (simulando a saída do inOrderTraversal)
	dados := []KVPair{
		{Key: []byte("a_chave"), Value: []byte("valor_a")},
		{Key: []byte("b_chave"), Value: []byte("valor_b")},
		{Key: []byte("c_chave"), Value: []byte("valor_c")},
	}

	// 2. Escreve no disco (encriptado e empacotado em binário)
	err := WriteSSTable(dados, filename)
	if err != nil {
		t.Fatalf("Falha ao escrever SSTable: %v", err)
	}

	// 3. Lê do disco buscando uma chave existente
	val, err := FindInSSTable([]byte("b_chave"), filename)
	if err != nil {
		t.Fatalf("Erro ao buscar chave existente: %v", err)
	}

	if !bytes.Equal(val, []byte("valor_b")) {
		t.Errorf("Esperava 'valor_b', recebeu '%s'", string(val))
	}

	// 4. Busca uma chave inexistente (deve retornar ErrKeyNotFound)
	_, err = FindInSSTable([]byte("z_chave_fantasma"), filename)
	if err != ErrKeyNotFound {
		t.Errorf("Esperava ErrKeyNotFound para chave inexistente, recebeu %v", err)
	}

	// 5. ReadSSTableKeys devolve todas as chaves, na ordem gravada
	chaves, err := ReadSSTableKeys(filename)
	if err != nil || len(chaves) != len(dados) {
		t.Fatalf("Esperava %d chaves, recebeu %d (erro: %v)", len(dados), len(chaves), err)
	}
	for i, chave := range chaves {
		if !bytes.Equal(chave, dados[i].Key) {
			t.Errorf("Chave %d: esperava '%s', recebeu '%s'", i, dados[i].Key, chave)
		}
	}
}

func TestSSTable_ValorVazioNoFimDoArquivo(t *testing.T) {
	filename := "test_sstable_vazio.data"
	os.Remove(filename)
	defer os.Remove(filename)

	dados := []KVPair{
		{Key: []byte("a_chave"), Value: []byte("valor_a")},
		{Key: []byte("b_chave"), Value: []byte{}},
	}
	if err := WriteSSTable(dados, filename); err != nil {
		t.Fatalf("Falha ao escrever SSTable: %v", err)
	}

	val, err := FindInSSTable([]byte("b_chave"), filename)
	if err != nil || len(val) != 0 {
		t.Errorf("Esperava valor vazio sem erro, recebeu '%s' (erro: %v)", val, err)
	}
}
