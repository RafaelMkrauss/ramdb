package bufferedreader_test

import (
	"errors"
	"io"
	"testing"

	bufferedreader "ramdb/utils/buffered_reader"
)

// sliceReader entrega os itens de data, no maximo chunk por chamada.
// Quando data acaba, devolve err (io.EOF por padrao).
type sliceReader[T any] struct {
	data  []T
	chunk int
	err   error
	calls int
}

func (r *sliceReader[T]) Read(p []T) (int, error) {
	r.calls++
	if len(r.data) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		return 0, io.EOF
	}
	n := min(len(p), len(r.data))
	if r.chunk > 0 {
		n = min(n, r.chunk)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

// errReader falha em toda chamada de Read.
type errReader[T any] struct{ err error }

func (r *errReader[T]) Read(p []T) (int, error) { return 0, r.err }

func bytesOf(s string) []byte { return []byte(s) }

func TestReadMenorQueOBuffer(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("hello world")}
	reader := bufferedreader.New[byte](src, 8)
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("Read: erro inesperado: %v", err)
	}
	if n != 5 {
		t.Fatalf("Read: n = %d, esperado 5", n)
	}
	if got := string(buf); got != "hello" {
		t.Fatalf("Read: buf = %q, esperado %q", got, "hello")
	}
}

func TestReadSubsequenteVemDoBufferInterno(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("hello world")}
	reader := bufferedreader.New[byte](src, 16)

	buf := make([]byte, 5)
	if _, err := reader.Read(buf); err != nil {
		t.Fatalf("primeiro Read: %v", err)
	}
	callsDepoisDoPrimeiro := src.calls

	buf2 := make([]byte, 6)
	n, err := reader.Read(buf2)
	if err != nil {
		t.Fatalf("segundo Read: %v", err)
	}
	if n != 6 || string(buf2) != " world" {
		t.Fatalf("segundo Read: n = %d, buf = %q, esperado 6 e %q", n, string(buf2), " world")
	}
	if src.calls != callsDepoisDoPrimeiro {
		t.Errorf("segundo Read chamou o reader base %d vez(es); deveria ser servido do buffer interno",
			src.calls-callsDepoisDoPrimeiro)
	}
}

func TestReadSequencialConsomeTudoNaOrdem(t *testing.T) {
	const conteudo = "abcdefghijklmnopqrstuvwxyz0123456789"

	casos := []struct {
		nome       string
		bufferSize int
		tamanhos   []int
	}{
		{"leituras uniformes menores que o buffer", 8, []int{4, 4, 4, 4}},
		{"leitura parcial seguida de leitura maior", 8, []int{3, 5, 6, 6}},
		{"leitura maior que o buffer interno", 4, []int{10, 10}},
		{"buffer interno de tamanho 1", 1, []int{2, 3, 4}},
		{"alterna pequeno e grande", 8, []int{1, 9, 2, 7}},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			src := &sliceReader[byte]{data: bytesOf(conteudo)}
			reader := bufferedreader.New[byte](src, caso.bufferSize)

			offset := 0
			for i, tamanho := range caso.tamanhos {
				buf := make([]byte, tamanho)
				n, err := reader.Read(buf)
				if err != nil {
					t.Fatalf("Read #%d (tamanho %d): erro inesperado: %v", i, tamanho, err)
				}
				if n != tamanho {
					t.Fatalf("Read #%d: n = %d, esperado %d", i, n, tamanho)
				}
				esperado := conteudo[offset : offset+tamanho]
				if got := string(buf[:n]); got != esperado {
					t.Fatalf("Read #%d: buf = %q, esperado %q", i, got, esperado)
				}
				offset += n
			}
		})
	}
}

func TestReadMaiorQueOBufferInterno(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("0123456789abcdef")}
	reader := bufferedreader.New[byte](src, 4)

	buf := make([]byte, 12)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("Read: erro inesperado: %v", err)
	}
	if n != 12 || string(buf) != "0123456789ab" {
		t.Fatalf("Read: n = %d, buf = %q, esperado 12 e %q", n, string(buf), "0123456789ab")
	}
}

func TestReadComReaderBaseEntregandoPoucoPorVez(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("hello world"), chunk: 2}
	reader := bufferedreader.New[byte](src, 8)

	buf := make([]byte, 6)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("Read: erro inesperado: %v", err)
	}
	if n <= 0 || n > 6 {
		t.Fatalf("Read: n = %d, fora do intervalo (0, 6]", n)
	}
	if got, esperado := string(buf[:n]), "hello "[:n]; got != esperado {
		t.Fatalf("Read: buf[:%d] = %q, esperado %q", n, got, esperado)
	}
}

func TestReadRetornaDadosPendentesQuandoAFonteAcaba(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("hello")}
	reader := bufferedreader.New[byte](src, 8)

	buf := make([]byte, 3)
	if _, err := reader.Read(buf); err != nil {
		t.Fatalf("primeiro Read: %v", err)
	}

	// Restam 2 itens no buffer interno e a fonte esta vazia.
	buf2 := make([]byte, 4)
	n, err := reader.Read(buf2)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("segundo Read: erro inesperado: %v", err)
	}
	if n != 2 {
		t.Fatalf("segundo Read: n = %d, esperado 2 (os itens ainda no buffer interno)", n)
	}
	if got := string(buf2[:n]); got != "lo" {
		t.Fatalf("segundo Read: buf = %q, esperado %q", got, "lo")
	}
}

func TestReadPropagaErroDoReaderBase(t *testing.T) {
	falha := errors.New("falha na fonte")
	reader := bufferedreader.New[byte](&errReader[byte]{err: falha}, 8)

	buf := make([]byte, 4)
	n, err := reader.Read(buf)
	if !errors.Is(err, falha) {
		t.Fatalf("Read: err = %v, esperado %v", err, falha)
	}
	if n != 0 {
		t.Fatalf("Read: n = %d, esperado 0", n)
	}
}

func TestReadNaoChamaAFonteQuandoOBufferJaAtende(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("hello world")}
	reader := bufferedreader.New[byte](src, 16)

	if _, err := reader.Read(make([]byte, 1)); err != nil {
		t.Fatalf("primeiro Read: %v", err)
	}
	if src.calls != 1 {
		t.Fatalf("primeiro Read fez %d chamadas a fonte, esperado 1", src.calls)
	}
	for i := range 5 {
		if _, err := reader.Read(make([]byte, 2)); err != nil {
			t.Fatalf("Read #%d: %v", i+1, err)
		}
	}
	if src.calls != 1 {
		t.Errorf("fonte chamada %d vezes; os dados ja estavam no buffer interno", src.calls)
	}
}

func TestReadBufferVazio(t *testing.T) {
	src := &sliceReader[byte]{data: bytesOf("hello")}
	reader := bufferedreader.New[byte](src, 8)

	n, err := reader.Read([]byte{})
	if err != nil {
		t.Fatalf("Read com buf vazio: erro inesperado: %v", err)
	}
	if n != 0 {
		t.Fatalf("Read com buf vazio: n = %d, esperado 0", n)
	}
	if src.calls != 0 {
		t.Errorf("Read com buf vazio chamou a fonte %d vez(es), esperado 0", src.calls)
	}
}

func TestReadGenericoComTipoNaoByte(t *testing.T) {
	type ponto struct{ X, Y int }
	origem := []ponto{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}}

	src := &sliceReader[ponto]{data: append([]ponto(nil), origem...)}
	reader := bufferedreader.New[ponto](src, 2)

	lidos := make([]ponto, 0, len(origem))
	for len(lidos) < len(origem) {
		buf := make([]ponto, 2)
		n, err := reader.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("Read: erro inesperado: %v", err)
		}
		if n == 0 {
			break
		}
		lidos = append(lidos, buf[:n]...)
	}

	if len(lidos) != len(origem) {
		t.Fatalf("lidos %d pontos, esperado %d (%v)", len(lidos), len(origem), lidos)
	}
	for i := range origem {
		if lidos[i] != origem[i] {
			t.Fatalf("ponto %d = %v, esperado %v", i, lidos[i], origem[i])
		}
	}
}
