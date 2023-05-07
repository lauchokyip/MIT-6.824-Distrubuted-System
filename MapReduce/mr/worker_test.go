package mr

import (
	"fmt"
	"testing"
)

func TestIHASH(t *testing.T) {
	num := ihash("foo")
	fmt.Println(num)
	num = ihash("bar")
	fmt.Println(num)
}
