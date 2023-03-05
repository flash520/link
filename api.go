/**
 * @Author: koulei
 * @Description:
 * @File: api
 * @Version: 1.0.0
 * @Date: 2023/3/4 11:46
 */

package link

import "io"

type Codec interface {
	Send(msg interface{}) error
	Receive() (interface{}, error)
	Close() error
}

type Protocol interface {
	NewCodec(writer io.ReadWriter) (Codec, error)
}
