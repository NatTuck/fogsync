package pio

// IO functions that panic on error.

import (
	"os"
	"io"
	"fmt"
)

// First, file I/O

type File struct {
	file *os.File
}

func Open(name string) File {
	file, err := os.Open(name)
	checkError(err)
	return File{file}
}

func Create(name string) File {
	file, err := os.Create(name)
	checkError(err)
	return File{file}
}

func (ff *File) Close() {
	if ff.file == nil {
		return
	}

	err := ff.file.Close()
	checkError(err)

	ff.file = nil
}

func ReadFile(name string) []byte {
	file := Open(name)
	defer file.Close()

	data := make([]byte, 0)
	temp := make([]byte, 64 * 1024)

	for {
		nn, eof := file.Read(temp)
		if eof {
			break
		}

		data = append(data, temp[0:nn]...)
	}

	return data
}

func WriteFile(name string, data []byte) {
	file := Create(name)
	defer file.Close()
	file.Write(data)
}

// Returns (count, eof)
func (ff *File) Read(data []byte) (int, bool) {
	nn, err := ff.file.Read(data)

	if err == io.EOF {
		return 0, true
	}

	checkError(err)

	return nn, false
}

func (ff *File) MustReadN(nn int) []byte {
	data := make([]byte, nn)

	mm, err := ff.file.Read(data)
	checkError(err)

	if mm < nn {
		checkError(fmt.Errorf("Read less than %d bytes", nn))
	}

	return data
}

func (ff *File) Write(data []byte) {
	_, err := ff.file.Write(data)
	checkError(err)
}

func (ff *File) Seek(pos int64, whence int) {
	_, err := ff.file.Seek(pos, whence)
	checkError(err)
}

