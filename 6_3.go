package h2spec

import (
	"bytes"
	"fmt"
	"github.com/bradfitz/http2"
	"github.com/bradfitz/http2/hpack"
	"time"
)

func TestPriority(ctx *Context) {
	PrintHeader("6.3. PRIORITY", 0)

	func(ctx *Context) {
		desc := "Sends a PRIORITY frame with 0x0 stream identifier"
		msg := "The endpoint MUST respond with a connection error of type PROTOCOL_ERROR."
		result := false

		http2Conn := CreateHttp2Conn(ctx, true)
		defer http2Conn.conn.Close()

		var buf bytes.Buffer
		hdrs := []hpack.HeaderField{
			pair(":method", "GET"),
			pair(":scheme", "http"),
			pair(":path", "/"),
			pair(":authority", ctx.Authority()),
		}
		enc := hpack.NewEncoder(&buf)
		for _, hf := range hdrs {
			_ = enc.WriteField(hf)
		}

		var hp http2.HeadersFrameParam
		hp.StreamID = 1
		hp.EndStream = false
		hp.EndHeaders = true
		hp.BlockFragment = buf.Bytes()
		http2Conn.fr.WriteHeaders(hp)

		fmt.Fprintf(http2Conn.conn, "\x00\x00\x05\x02\x00\x00\x00\x00\x00")
		http2Conn.conn.Write(buf.Bytes())
		fmt.Fprintf(http2Conn.conn, "\x80\x00\x00\x01\x0a")

		timeCh := time.After(3 * time.Second)

	loop:
		for {
			select {
			case f := <-http2Conn.dataCh:
				gf, ok := f.(*http2.GoAwayFrame)
				if ok {
					if gf.ErrCode == http2.ErrCodeProtocol {
						result = true
					}
				}
			case <-http2Conn.errCh:
				break loop
			case <-timeCh:
				break loop
			}
		}

		PrintResult(result, desc, msg, 0)
	}(ctx)

	PrintFooter()
}
