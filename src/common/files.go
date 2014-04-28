package common

import (
	"os"
	"io"
	"path"
	"encoding/hex"
)

func CopyFile(dst string, src string) error {
    in, err := os.Open(src)
    if err != nil { 
		return err 
	}
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil { 
		return err 
	}
    
	_, err = io.Copy(out, in)
    
	cerr := out.Close()

    if err != nil {
		return err 
	}

    return cerr
}


func HashToPath(hash []byte) string {
	text := hex.EncodeToString(hash)

	d0 := text[0:3]
	d1 := text[3:6]
	d2 := text[6:9]

	return path.Join(d0, d1, d2, text)
}

