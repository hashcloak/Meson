// main.go - Crypto currency transaction submition Kaetzchen service plugin program.
// Copyright (C) 2018  David Stainton.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"syscall"

	"github.com/hashcloak/Meson/plugin/pkg/config"
	"github.com/hashcloak/Meson/plugin/pkg/proxy"
	"github.com/katzenpost/server/cborplugin"
	"github.com/ugorji/go/codec"
	"gopkg.in/op/go-logging.v1"
)

var log = logging.MustGetLogger("Meson")
var logFormat = logging.MustStringFormatter(
	"%{level:.4s} %{id:03x} %{message}",
)

func stringToLogLevel(level string) (logging.Level, error) {
	switch level {
	case "DEBUG":
		return logging.DEBUG, nil
	case "INFO":
		return logging.INFO, nil
	case "NOTICE":
		return logging.NOTICE, nil
	case "WARNING":
		return logging.WARNING, nil
	case "ERROR":
		return logging.ERROR, nil
	case "CRITICAL":
		return logging.CRITICAL, nil
	}
	return -1, fmt.Errorf("invalid logging level %s", level)
}

func setupLoggerBackend(level logging.Level, writer io.Writer) logging.LeveledBackend {
	format := logFormat
	backend := logging.NewLogBackend(writer, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	leveler := logging.AddModuleLevel(formatter)
	leveler.SetLevel(level, "currency")
	return leveler
}

func parametersHandler(currency *proxy.Currency, response http.ResponseWriter, req *http.Request) {
	p := currency.GetParameters()
	params := cborplugin.Parameters(p)
	var serialized []byte
	enc := codec.NewEncoderBytes(&serialized, new(codec.CborHandle))
	if err := enc.Encode(params); err != nil {
		panic(err)
	}
	_, err := response.Write(serialized)
	if err != nil {
		panic(err)
	}
}

func requestHandler(currency *proxy.Currency, response http.ResponseWriter, req *http.Request) {
	log.Debug("request handler")
	cborHandle := new(codec.CborHandle)
	request := cborplugin.Request{
		Payload: make([]byte, 0),
	}
	err := codec.NewDecoder(req.Body, new(codec.CborHandle)).Decode(&request)
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
	currencyRequestLen := binary.BigEndian.Uint32(request.Payload[:4])
	log.Debug("decoded request")
	currencyResponse, err := currency.OnRequest(request.ID, request.Payload[4:4+currencyRequestLen], request.HasSURB)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// send length prefixed CBOR response
	reply := cborplugin.Response{
		Payload: currencyResponse,
	}
	var serialized []byte
	enc := codec.NewEncoderBytes(&serialized, cborHandle)
	if err := enc.Encode(reply); err != nil {
		log.Error(err.Error())
		panic(err)
	}
	log.Debugf("serialized response is len %d", len(serialized))
	_, err = response.Write(serialized)
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
	log.Debug("sent response")
}

func main() {
	var logLevel string
	var logDir string
	flag.StringVar(&logDir, "log_dir", "", "logging directory")
	flag.StringVar(&logLevel, "log_level", "DEBUG", "logging level could be set to: DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL")
	cfgFile := flag.String("f", "currency.toml", "Path to the currency config file.")
	flag.Parse()

	level, err := stringToLogLevel(logLevel)
	if err != nil {
		fmt.Println("Invalid logging-level specified.")
		os.Exit(1)
	}

	// Ensure that the log directory exists.
	s, err := os.Stat(logDir)
	if os.IsNotExist(err) {
		fmt.Printf("Log directory '%s' doesn't exist.", logDir)
		os.Exit(1)
	}
	if !s.IsDir() {
		fmt.Println("Log directory must actually be a directory.")
		os.Exit(1)
	}

	// Log to a file.
	logFile := path.Join(logDir, fmt.Sprintf("currency.%d.log", os.Getpid()))
	f, err := os.Create(logFile)
	logBackend := setupLoggerBackend(level, f)
	log.SetBackend(logBackend)
	log.Info("currency server started")

	// Set the umask to something "paranoid".
	syscall.Umask(0077)

	// Load config file.
	cfg, err := config.LoadFile(*cfgFile)
	if err != nil {
		log.Errorf("Failed to load config file '%v: %v\n'", *cfgFile, err)
		log.Error("Exiting")
		os.Exit(-1)
	}

	// Start service.
	currency, err := proxy.New(cfg)
	if err != nil {
		log.Errorf("Failed to load proxy config: %v\n", err)
		log.Error("Exiting")
		panic(err)
	}
	_requestHandler := func(response http.ResponseWriter, request *http.Request) {
		requestHandler(currency, response, request)
	}
	_parametersHandler := func(response http.ResponseWriter, request *http.Request) {
		parametersHandler(currency, response, request)
	}
	server := http.Server{}
	socketFile := fmt.Sprintf("/tmp/%d.currency.socket", os.Getpid())
	unixListener, err := net.Listen("unix", socketFile)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/request", _requestHandler)
	http.HandleFunc("/parameters", _parametersHandler)
	fmt.Printf("%s\n", socketFile)
	server.Serve(unixListener)
	os.Remove(socketFile)
}
