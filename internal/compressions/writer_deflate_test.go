// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved.

package compressions_test

import (
	"bytes"
	"github.com/dashenmiren/EdgeNode/internal/compressions"
	"strings"
	"testing"
)

func BenchmarkDeflateWriter_Write(b *testing.B) {
	var data = []byte(strings.Repeat("A", 1024))

	for i := 0; i < b.N; i++ {
		var buf = &bytes.Buffer{}
		writer, err := compressions.NewDeflateWriter(buf, 5)
		if err != nil {
			b.Fatal(err)
		}

		for j := 0; j < 100; j++ {
			_, err = writer.Write(data)
			if err != nil {
				b.Fatal(err)
			}

			/**err = writer.Flush()
			if err != nil {
				b.Fatal(err)
			}**/
		}

		_ = writer.Close()
	}
}
