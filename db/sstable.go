package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// WriteSSTable recebe os dados em ordem, serializa em binário e salva encriptado.
func WriteSSTable(data []KVPair, filepath string) error {
	var buf bytes.Buffer

	for _, kv := range data {
		// Grava tamanho e bytes da chave
		err := binary.Write(&buf, binary.LittleEndian, uint32(len(kv.Key)))
		if err != nil {
			return err
		}
		buf.Write(kv.Key)

		// Grava tamanho e bytes do valor
		err = binary.Write(&buf, binary.LittleEndian, uint32(len(kv.Value)))
		if err != nil {
			return err
		}
		buf.Write(kv.Value)
	}

	// Chama o Encrypt APENAS com os bytes, sem pedir a chave (já tá no crypto.go)
	encryptedData, err := Encrypt(buf.Bytes())
	if err != nil {
		return fmt.Errorf("falha ao encriptar sstable: %v", err)
	}

	err = os.WriteFile(filepath, encryptedData, 0600)
	if err != nil {
		return fmt.Errorf("falha ao gravar arquivo sstable no disco: %v", err)
	}

	return nil
}

// Adicione os imports "bytes" e "encoding/binary" se não estiverem lá,
// e o ErrKeyNotFound (pode referenciar do pacote db diretamente).

// FindInSSTable abre o arquivo, decripta a carga AES-256 e busca a chave sequencialmente.
func FindInSSTable(searchKey []byte, filepath string) ([]byte, error) {
	// 1. Lê os bytes criptografados do disco
	encryptedData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// 2. Descriptografa tudo para a memória
	decryptedData, err := Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("falha ao decriptar %s: %v", filepath, err)
	}

	// 3. Buffer de leitura para o nosso formato binário denso
	buf := bytes.NewReader(decryptedData)

	for buf.Len() > 0 {
		// Lê o tamanho da chave (4 bytes)
		var keyLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &keyLen); err != nil {
			return nil, err
		}

		// Extrai a chave
		k := make([]byte, keyLen)
		if _, err := buf.Read(k); err != nil {
			return nil, err
		}

		// Lê o tamanho do valor (4 bytes)
		var valLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &valLen); err != nil {
			return nil, err
		}

		// Extrai o valor
		v := make([]byte, valLen)
		if _, err := buf.Read(v); err != nil {
			return nil, err
		}

		// Se achamos a chave desejada, retornamos o valor imediatamente!
		if bytes.Equal(k, searchKey) {
			return v, nil
		}
	}

	// Varreu o arquivo todo e não achou
	return nil, ErrKeyNotFound
}
