package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	"io"
)

// é absurdamente necessário  que tenha 32 bitts.
var devEncryptionKey = []byte("ramdb-super-secret-key-32-bytes!")

// implementar depois
/*
func LoadEncryptionKey() ([]byte, error) {
	keyStr := os.Getenv("RAMDB_ENCRYPTION_KEY")

	if keyStr == "" {
		return nil, fmt.Errorf("variável de ambiente RAMDB_ENCRYPTION_KEY não encontrada")
	}

	keyBytes := []byte(keyStr)

	// O AES-256 exige estritamente 32 bytes. Nem mais, nem menos.
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("a chave de criptografia deve ter exatamente 32 bytes (recebeu %d bytes)", len(keyBytes))
	}

	return keyBytes, nil
}
*/
// Encrypt sela o payload usando AES-GCM-256. A key precisa ter exatamente 32 bytes.
func Encrypt(payload []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Cria um nonce aleatório do tamanho padrão do GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal anexa o nonce e os dados criptografados juntos no array de destino
	ciphertext := aesGCM.Seal(nonce, nonce, payload, nil)
	return ciphertext, nil
}

// Decrypt abre o payload encriptado usando a mesma chave AES.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext muito curto")
	}

	// Separa o nonce do payload criptografado
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Descriptografa os dados
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err // Pode ser chave errada ou arquivo corrompido
	}

	return plaintext, nil
}
