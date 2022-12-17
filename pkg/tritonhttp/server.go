package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"net"
	"os"
)

/**
 * The proccess of request to response will be in server
 * Since it opens the port to listen and send the request
 * data to get response
 */

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	// The address is like the address we had in PA2
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	/**
	 * IT stores all the data that client can access on server. It means
	 * these data are visible to client
	 * If client requests any file under DocRoot, it's fine but not the
	 * outside of the path (client requests any file that is valid based on the path)
	 * It's array (based on disscussion)
	 * The client can only request the files that are in the path of DocRoot
	 */
	DocRoot string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
// Listen on TCP Address
func (s *Server) ListenAndServe() error {

	//Let's validate the file first
	if err := s.validateServer(); err != nil {

		return err
	}

	//Since server is valid, Let's open the port to listen
	listenTo, err := net.Listen("tcp", s.Addr)

	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		return err
	}

	fmt.Println("Listen on ", listenTo.Addr())
	defer listenTo.Close()

	//Let's accept the connection from the client
	for {

		connect, err := listenTo.Accept()

		if err != nil {
			continue
		}
		fmt.Println("Accepted", connect.RemoteAddr())
		go s.HandleConnection(connect)
	}
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {

	buffRead := bufio.NewReader(conn)

	for {

		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {

			fmt.Println("time has been failed for connection", conn)
			_ = conn.Close()
			return
		}

		// Read next request
		req, byteAmount, err := ReadRequest(buffRead)

		// Handle EOF
		if errors.Is(err, io.EOF) {

			_ = conn.Close()
			return
		}

		// Handle time Out (we go based on discussion and then check for byte)
		if err, ok := err.(net.Error); ok && err.Timeout() {

			fmt.Println("Time is out")

			if byteAmount {

				fmt.Println("only part of requset has arrived")
				resp := &Response{}
				resp.HandleBadRequest()
				resp.Write(conn)
			}

			_ = conn.Close()
			return
		}

		// Handeling Bad Request
		if err != nil {

			fmt.Println("Bad request occurs during reading file")
			resp2 := &Response{}
			resp2.HandleBadRequest()
			_ = resp2.Write(conn)
			_ = conn.Close()
			return
		}

		//Handle Good Request

		respGood := s.HandleGoodRequest(req)
		err = respGood.Write(conn)
		if err != nil {

			fmt.Println(err)
		}

		//Close conn if requested
		if req.Close {
			conn.Close()
			return
		}
	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
// If request is valid, How to handle it?
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {

	returnRes := &Response{}

	size := len(req.URL) - 1

	if req.URL[size] != '/' {

		filepath.Clean(req.URL)
		firstPath := filepath.Join(s.DocRoot, req.URL)
		filepath.Clean(firstPath)
		returnRes.FilePath = firstPath
	} else {

		firstPath := filepath.Join(req.URL, "index.html")
		filepath.Clean(firstPath)
		finaPath := filepath.Join(s.DocRoot, firstPath)
		filepath.Clean(finaPath)
		returnRes.FilePath = finaPath
	}

	// Now we need to make sure the file path exists

	checkPath, err := os.Stat(returnRes.FilePath)
	fmt.Println("This is what we got for Path", checkPath)
	// Check it if fail tests
	if os.IsNotExist(err) {

		fmt.Println("The checkPath failed ", checkPath)
		// Using handle not found
		returnRes.HandleNotFound(req)
		return returnRes
	}

	// Handle Okay

	returnRes.HandleOK(req, returnRes.FilePath)
	return returnRes
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {

	res.Header = make(map[string]string)
	res.Request = req
	res.StatusCode = 200
	res.Proto = "HTTP/1.1"
	checkFile, err := os.Stat(path)
	if err != nil {
		fmt.Println("File might have problem")
	}
	str1 := "Date"
	str1A := CanonicalHeaderKey(str1)
	res.Header[str1A] = FormatTime(time.Now())
	str2 := "Last-Modified"
	str2A := CanonicalHeaderKey(str2)
	res.Header[str2A] = FormatTime(checkFile.ModTime())
	str3 := "Content-Type"
	str3A := CanonicalHeaderKey(str3)
	res.Header[str3A] = MIMETypeByExtension(filepath.Ext(res.FilePath))
	str4 := "Content-Length"
	str4A := CanonicalHeaderKey(str4)
	res.Header[str4A] = strconv.FormatInt(checkFile.Size(), 10)

	if req.Close {
		str5 := "Connection"
		str5A := CanonicalHeaderKey(str5)
		res.Header[str5A] = "close"
	}

}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	res.StatusCode = 400
	res.Proto = "HTTP/1.1"
	res.FilePath = ""
	res.Header = make(map[string]string)
	str1 := "Date"
	str1A := CanonicalHeaderKey(str1)
	res.Header[str1A] = FormatTime(time.Now())
	str2 := "Connection"
	str2A := CanonicalHeaderKey(str2)
	res.Header[str2A] = "close"
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {

	res.StatusCode = 404
	res.Proto = "HTTP/1.1"
	res.FilePath = ""
	res.Header = make(map[string]string)
	str1 := "Date"
	str1A := CanonicalHeaderKey(str1)
	res.Header[str1A] = FormatTime(time.Now())
	if req.Close {
		str2 := "Connection"
		str2A := CanonicalHeaderKey(str2)
		res.Header[str2A] = "close"
	}

}

func (s *Server) validateServer() error {

	validDoc, err := os.Stat(s.DocRoot)

	if os.IsNotExist(err) {
		return err
	}

	if !validDoc.IsDir() {
		return badhandle("The intended file which is %q, is not Dir", s.DocRoot)
	}

	return nil
}

func badhandle(what, val string) error {

	return fmt.Errorf("%s %q", what, val)
}
