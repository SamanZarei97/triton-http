package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	// Recieve from client
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if err := res.WriteBody(w); err != nil {
		return err
	}
	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {

	if res.Proto != "HTTP/1.1" {
		return badRespones("Value is not intended", res.Proto)
	}
	bw := bufio.NewWriter(w)
	var statText string
	if res.StatusCode == 200 {

		statText = "OK"

	} else if res.StatusCode == 400 {

		statText = "Bad Request"
	} else {

		statText = "Not Found"
	}

	startLine := fmt.Sprintf("%v %v %v\r\n", res.Proto, res.StatusCode, statText)

	if _, err := bw.WriteString(startLine); err != nil {

		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	return nil
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func (res *Response) WriteSortedHeaders(w io.Writer) error {

	bw := bufio.NewWriter(w)
	// Make sure that Map is in sorted order
	keyCode := make([]string, 0, len(res.Header))

	for i := range res.Header {
		keyCode = append(keyCode, i)
	}

	sort.Strings(keyCode)

	for _, i := range keyCode {

		headLine := fmt.Sprintf("%v: %v\r\n", i, res.Header[i])

		if _, err := bw.WriteString(headLine); err != nil {
			return err
		}

		if err := bw.Flush(); err != nil {
			return err
		}
	}

	if _, err := bw.WriteString("\r\n"); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	return nil
}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {

	if res.FilePath == "" {

		return badRespones("No file", res.FilePath)
	}

	br, err := os.ReadFile(res.FilePath)

	if err != nil {
		return err
	}

	bw := bufio.NewWriter(w)

	passage := fmt.Sprintf("%v", string(br))

	if _, err := bw.WriteString(passage); err != nil {

		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	return nil
}

func badRespones(what, val string) error {

	return fmt.Errorf("%s %q", what, val)
}
