package db

// KVPair representa uma linha de dado para ser trafegada entre a Memória e o Disco
type KVPair struct {
	Key   []byte
	Value []byte
}
