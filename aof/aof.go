package aof

import (
	"log"
	"os"
	"sync"
	"time"
)

type AOF struct {
	file *os.File
	ch   chan string
	wg   sync.WaitGroup
}

// NewAOF cria ou abre o arquivo de persistência e inicia a rotina de gravação
func NewAOF(path string) (*AOF, error) {
	// Abre o arquivo para Leitura/Escrita, cria se não existir e anexa no final (Append)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	aof := &AOF{
		file: f,
		ch:   make(chan string, 1024), // Buffer de 1024 comandos para evitar bloqueio do handler
	}

	aof.wg.Add(1)
	go aof.worker() // Inicia a goroutine de gravação em background

	return aof, nil
}

// worker processa os comandos do canal e salva no disco periodicamente
func (a *AOF) worker() {
	defer a.wg.Done()
	
	// Sincroniza o arquivo com o disco físico a cada 1 segundo (estilo Redis 'appendfsync everysec')
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case cmd, ok := <-a.ch:
			if !ok {
				// Canal fechado (servidor desligando), faz o flush final
				a.file.Sync()
				return
			}
			_, err := a.file.WriteString(cmd)
			if err != nil {
				log.Println("Erro ao gravar no AOF:", err)
			}
		case <-ticker.C:
			// Força a sincronização dos buffers do SO para o disco físico
			a.file.Sync()
		}
	}
}

// Append envia um comando (já formatado em RESP) para ser salvo de forma assíncrona
func (a *AOF) Append(cmd string) {
	a.ch <- cmd
}

// Close finaliza o AOF graciosamente (chame isso no encerramento do main.go)
func (a *AOF) Close() {
	close(a.ch) // Fecha o canal, avisando o worker para parar
	a.wg.Wait() // Aguarda o worker terminar a última gravação
	a.file.Close()
}

// ReadFile retorna o arquivo para que o boot do sistema possa reler os comandos
func (a *AOF) ReadFile() *os.File {
	// Retorna o ponteiro do arquivo para a posição 0
	a.file.Seek(0, 0)
	return a.file
}