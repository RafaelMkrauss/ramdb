package bufferedreader

type Reader[T any] interface {
	Read([]T) (int, error)
}

type BufferedReader[T any] struct {
	reader   Reader[T]
	buffer   []T
	pos      int
	data_end int
}

func New[T any](reader Reader[T], buffer_size int) BufferedReader[T] {
	return BufferedReader[T]{
		reader:   reader,
		buffer:   make([]T, buffer_size),
		pos:      0,
		data_end: 0,
	}
}

func (reader *BufferedReader[T]) numBufBytes() int {
	return reader.data_end - reader.pos
}

func (reader *BufferedReader[T]) Read(buf []T) (int, error) {
	remaining := reader.numBufBytes()
	if len(buf) <= remaining {
		copy(buf, reader.buffer[reader.pos:reader.data_end])
		reader.pos += len(buf)
		return len(buf), nil
	}

	missing := len(buf) - remaining
	copy(buf, reader.buffer[reader.pos:reader.data_end])
	tmp := make([]T, missing+len(reader.buffer))
	nread, err := reader.reader.Read(tmp)
	if nread < missing {
		copy(buf[remaining:], tmp[:nread])
		reader.pos += remaining
		return remaining + nread, err
	}
	copy(buf[remaining:], tmp[:missing])
	copy(reader.buffer, tmp[missing:nread])
	reader.pos = 0
	reader.data_end = nread - missing
	return len(buf), err
}
