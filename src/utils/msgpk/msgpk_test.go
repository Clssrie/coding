package msgpk

import (
	"fmt"
	"testing"
)

func testUnpack(t *testing.T) {
	fmt.Println("111111111111111111111111")
	pack := Pack([]byte("zhao"))
	fmt.Println(pack, string(pack), len(pack))

	unpack, err := Unpack(pack)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(unpack)

}
