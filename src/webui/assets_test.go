package webui

import (
	"testing"
	"fmt"
)

func TestRead (tt *testing.T) {
	derp := read()
	fmt.Println(derp)
}
