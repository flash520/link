/**
 * @Author: koulei
 * @Description:
 * @File: hj212_test
 * @Version: 1.0.0
 * @Date: 2023/3/5 23:24
 */

package link

import (
	"bytes"
	"fmt"
	"testing"
)

func TestHJ212(t *testing.T) {
	const prefixID = "##"
	const suffixID = "\r\n"
	var rn = "##1234\r\n"

	buff := make([]byte, 20)
	reader := bytes.NewReader([]byte(rn))
	n, err := reader.Read(buff)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buff[:2]) == prefixID)
	fmt.Println(string(buff[n-2:n]) == suffixID)
}
