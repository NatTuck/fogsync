package fs

import (
	"os"
	"os/exec"
	"io"
	"fmt"
	"path"
	"encoding/hex"
)

func CopyFile(dst string, src string) error {
    in, err := os.Open(src)
    if err != nil { 
		return TraceError(err)
	}
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil { 
		return TraceError(err)
	}
    
	_, err = io.Copy(out, in)
    
	cerr := out.Close()

    if err != nil {
		return TraceError(err)
	}

	if cerr != nil {
		return TraceError(cerr)
	}

    return nil
}

func CopyAll(dst string, src string) error {
	cp := exec.Command("cp", "-r", src, dst)

	err := cp.Run()
	if err != nil {
		msg := fmt.Errorf("Error in 'cp' command: %v\n", err)
		return TraceError(msg)
	}

	return nil
}

func HashToPath(hash []byte) string {
	text := hex.EncodeToString(hash)

	d0 := text[0:3]
	d1 := text[3:6]
	//d2 := text[6:9]

	return path.Join(d0, d1, text)
}

func ReadN(file *os.File, nn int64) []byte {
	temp := make([]byte, nn)

	mm, err := file.Read(temp)
	CheckError(err)

	if int64(mm) != nn {
		PanicHere("Read came up short")
	}
	
	return temp
}

func FindFiles(base_path string, fn func(string)) {
	FindAll(base_path, func(pp string) {
		info, err := os.Lstat(pp)
		CheckError(err)

		if !info.Mode().IsDir() {
			fn(pp)
		}
	})
}

func FindAll(base_path string, fn func(string)) {
	info, err := os.Lstat(base_path)
	CheckError(err)
		
	fn(base_path)

	if !info.Mode().IsDir() {
		return
	}

	dir, err := os.Open(base_path)
	CheckError(err)
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	CheckError(err)

	for _, name := range(names) {
		FindAll(path.Join(base_path, name), fn)
	}
}

