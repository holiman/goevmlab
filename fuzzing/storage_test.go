package fuzzing

import (
	"fmt"
	"testing"
)

func TestStorageOps(t *testing.T){
	p := RandStorageOps()
	fmt.Printf("%x \n", p)
}
