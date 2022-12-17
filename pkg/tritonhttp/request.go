package tritonhttp

import (
	"bufio"
	"fmt"
	"strings"
)

/**
 * This class is a wrapper for DS
 */

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	// It's a map of key:Val pair

	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
// It rquests into req
// Return bool if request aby buffer

func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {

	bytesReceived = false
	var count int = 0
	req = &Request{}
	// Read start line
	oneLine, err := ReadLine(br)

	if err != nil {
		return nil, bytesReceived, err
	}
	bytesReceived = true
	fmt.Println("The start line is:", oneLine)
	array := strings.Split(oneLine, " ")

	if len(array) != 3 {

		fmt.Println("the spilit has problem")
		return nil, bytesReceived, strResult("Array Problem", oneLine)
	}

	fmt.Println("Spliting cuased the result:", array)

	req.Method = strings.TrimSpace(array[0])

	if req.Method != "GET" {
		fmt.Println("req.Method is not GET")
		return nil, bytesReceived, strResult("req.Method PRoblem", req.Method)
	}

	fmt.Println("The Method is:", req.Method)

	req.URL = strings.TrimSpace(array[1])

	fmt.Println("The URL is:", req.URL)

	if req.URL == "" {
		fmt.Println("req.URL problem")
		return nil, bytesReceived, strResult("req.URL PRoblem", req.URL)
	}

	if req.URL[0] != '/' {

		fmt.Println("req.URL problem")
		return nil, bytesReceived, strResult("req.URL PRoblem", req.URL)
	}

	req.Proto = strings.TrimSpace(array[2])

	if req.Proto == "" {
		fmt.Println("req.Proto problem")
		return nil, bytesReceived, strResult("req.Proto PRoblem", req.Proto)
	}

	if req.Proto != "HTTP/1.1" {
		return nil, bytesReceived, strResult("req.Proto PRoblem", req.Proto)
	}

	fmt.Println("The Proto is:", req.Proto)

	req.Header = make(map[string]string)

	for {

		// Read headers
		oneLine, err := ReadLine(br)

		if err != nil {
			return nil, bytesReceived, strResult("line PRoblem", oneLine)
		}

		fmt.Println("The header Line is:", oneLine)
		if oneLine == "" {
			fmt.Println("END OF FILE")
			break
		}

		if !strings.Contains(oneLine, ":") {
			fmt.Println("The is no : for sepration")
			return nil, bytesReceived, strResult("line PRoblem", oneLine)
		}

		group := strings.Split(oneLine, ":")
		fmt.Println("The Group is:", group)
		tempKey := group[0]
		fmt.Println("The key before parsing is:", tempKey)
		if tempKey == "" {
			fmt.Println("Key is empty")
			return nil, bytesReceived, strResult("req.Proto PRoblem", tempKey)
		}

		key := CanonicalHeaderKey(tempKey)
		fmt.Println("The key after parsing is:", key)
		if key == "" {
			fmt.Println("Key is empty")
			return nil, bytesReceived, strResult("req.Proto PRoblem", key)
		}
		tempValue := group[1]
		fmt.Println("The value before triming is:", tempValue)
		fmt.Println("The size of value before triming is:", len(tempValue))
		value := strings.TrimSpace(tempValue)
		fmt.Println("The value after triming is:", value)
		fmt.Println("The size of value after triming is:", len(value))
		// Check required headers

		if key == "Host" {

			count++
			fmt.Println("The key is Host")
			req.Host = value

		} else if key == "Connection" {

			fmt.Println("The key is Connection")
			req.Close = true

		} else {

			fmt.Println("Other headers")
			req.Header[key] = value
		}

	}

	if req.Host == "" {
		fmt.Println("No Host or host value")
		return nil, bytesReceived, strResult("req.Proto PRoblem", req.Host)
	}

	if count != 1 {
		return nil, bytesReceived, strResult("req.Proto PRoblem", req.Host)
	}
	fmt.Println("The Host value is:", req.Host)
	return req, bytesReceived, nil
}

func strResult(what, val string) error {

	return fmt.Errorf("%s %q", what, val)
}
