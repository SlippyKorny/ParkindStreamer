package streaming

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/TheSlipper/ParkindStreamer/logging"
	"github.com/google/uuid"
)

// Handler of the client's http server
type parkindStreamerHandler struct {
	verbose bool           // verbosity flag
	token   uuid.UUID      // new connection token
	cs      *cameraSession // streaming session
}

// Implementation of the http.Handler interface. Used as a simple router that calls handlers
// that correspond to given urls
func (p parkindStreamerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := req.URL.String()
	logging.InfoLog(p.verbose, "Received a", req.Method, "request at", url)

	if url == "/check/" {
		connectionTestHandle(rw, req, &p)
	} else {
		invalidURLHandle(rw, req)
	}
}

// CreateHTTPServer sets up an http server of the parkind client that will listen only to requests
// from the passed address and then return it
func CreateHTTPServer(verbose bool, address string) (s *http.Server, ecs *cameraSession, err error) {
	// Create an http handler
	handler := parkindStreamerHandler{verbose: verbose}

	// Generate a connection token for this session
	handler.token, err = uuid.NewRandom()
	if err != nil {
		return
	}
	logging.InfoLog(verbose, "Successfully generated a new connection token:",
		handler.token.String())

	// Create a streaming session
	cs, err := NewCameraSession(1, 1)
	if err != nil {
		logging.ErrorLog(err.Error())
		os.Exit(2)
	}
	ecs = &cs
	handler.cs = &cs

	if address != "" {
		cs.AddDestination(address+":8000", "api/frame")
		go func() {
			err := cs.Stream()
			if err != nil {
				logging.ErrorLog(err.Error())
				os.Exit(3)
			}
		}()

		handler.cs = &cs
	}

	// Set up the server
	s = &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	logging.InfoLog(verbose, "Server listening at 127.0.0.1:8080")

	return
}

// invalidURLHandle handles all of the invalid incoming requests
func invalidURLHandle(rw http.ResponseWriter, req *http.Request) {
	logging.ErrorLog("invalid", req.Method, "request for URL:", req.URL.String())
	rw.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(rw, "<h1>Error 400: Bad request</h1>")
}

// Url handle that checks if the connection can be established with the given data
func connectionTestHandle(rw http.ResponseWriter, req *http.Request, p *parkindStreamerHandler) {
	// as of now it'll be the default until token generation is implemented
	fail := func() {
		rw.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(rw, "<h1>Error 403: Forbidden</h1>")
		logging.InfoLog(p.verbose, "A connection test request failed")
	}

	// If invalid method then fail
	if req.Method != "POST" {
		fail()
		return
	}

	// if unable to get the form arguments then fail
	if err := req.ParseForm(); err != nil {
		fail()
		return
	}

	// get the token and compare it with the local one
	// infoLog("Local uuid:", p.token.String(), "\n\tReceived form:", req.Pos)
	token := req.FormValue("token")
	if token == "" || token != p.token.String() {
		fail()
		return
	}

	rw.WriteHeader(http.StatusOK)
}
