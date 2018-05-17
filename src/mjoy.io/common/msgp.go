////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: msgp.go
// @Date: 2018/05/17 11:03:17
////////////////////////////////////////////////////////////////////////////////

package common

import (
	"github.com/tinylib/msgp/msgp"
	"bytes"
	"mjoy.io/utils/crypto/sha3"
	"mjoy.io/common/types"
	"fmt"
)

func msgpHash(x interface{}) (h types.Hash, err error) {
	defer func() {
		panic := recover()
		if panic != nil {
			err = fmt.Errorf("%s", panic)
		}
	}()

	var buf bytes.Buffer
	wr := msgp.NewWriter(&buf)
	err = wr.WriteIntf(x)
	if err != nil{
		return
	}

	err = wr.Flush()
	if err != nil{
		return
	}

	hw:=sha3.NewKeccak256()
	hw.Write(buf.Bytes())
	hw.Sum(h[:0])

	return h, nil
}