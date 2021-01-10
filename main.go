package main

import (
	"errors"
	"flag"
	"os"

	"github.com/TheSlipper/ParkindStreamer/logging"
	"github.com/TheSlipper/ParkindStreamer/streaming"
)

// runtimeArgs contains runtime arguments of the parkind client
type runtimeArgs struct {
	verbosity bool
	addr      string
	login     string
	password  string
}

// setUpRuntimeArgs loads the command line and environment arguments into the args singleton
func setUpRuntimeArgs() (args runtimeArgs, err error) {
	flag.BoolVar(&args.verbosity, "verbose", false, "defines how much information should be printed out")
	flag.StringVar(&args.addr, "address", "", "ip address to the Parkind server (e.g.: 127.0.0.1)")
	flag.Parse()

	args.login = os.Getenv("LOGIN")
	args.password = os.Getenv("PASSWORD")
	if args.login == "" || args.password == "" {
		return args, errors.New("LOGIN or PASSWORD were not provided")
	}

	return args, nil
}

func main() {
	// Set up runtime arguments or stop execution if not satisfied
	args, err := setUpRuntimeArgs()
	if err != nil {
		logging.ErrorLog(err.Error())
		os.Exit(1)
	}

	// Set up a local http server with a streaming session
	var addr string
	if args.addr == "" {
		addr = ""
	} else {
		addr = args.addr
	}
	server, cs, err := streaming.CreateHTTPServer(args.verbosity, addr)
	if err != nil {
		logging.ErrorLog(err.Error())
		os.Exit(4)
	}
	defer cs.Close()
	defer server.Close()

	// Start the server
	err = server.ListenAndServe()
	if err != nil {
		logging.ErrorLog(err.Error())
		os.Exit(5)
	}
}
