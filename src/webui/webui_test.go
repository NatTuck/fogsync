package webui

import (
	"testing"
	"time"
	"fmt"
)

func TestWebserver (tt *testing.T) {
	fmt.Println("Starting server...")
	Start()
	time.Sleep(1000 * time.Second)
}
