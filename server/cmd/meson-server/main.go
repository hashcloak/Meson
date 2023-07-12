// main.go - Katzenpost server binary.
// Copyright (C) 2017  Yawning Angel.
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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	server "github.com/hashcloak/Meson/server"
	"github.com/hashcloak/Meson/server/config"
)

func main() {
	cfgFile := flag.String("f", "katzenpost.toml", "Path to the server config file.")
	genOnly := flag.Bool("g", false, "Generate the keys and exit immediately.")
	testConfig := flag.Bool("t", false, "Test meson server config.")
	cpuProfilePath := flag.String("cpuprofilepath", "", "Path to the pprof cpu profile")
	memProfilePath := flag.String("memprofilepath", "", "Path to the pprof memory profile")
	flag.Parse()

	if *cpuProfilePath != "" {
		f, err := os.Create(*cpuProfilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create CPU profile '%v': %v\n", *cpuProfilePath, err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start CPU profile '%v': %v\n", *cpuProfilePath, err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	if *memProfilePath != "" {
		f, err := os.Create(*memProfilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create memory profile '%v': %v\n", *memProfilePath, err)
			os.Exit(1)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write memory profile'%v': %v\n", *memProfilePath, err)
			os.Exit(1)
		}
		f.Close()
	}

	// Set the umask to something "paranoid".
	syscall.Umask(0077)

	// Ensure that a sane number of OS threads is allowed.
	if os.Getenv("GOMAXPROCS") == "" {
		// But only if the user isn't trying to override it.
		nProcs := runtime.GOMAXPROCS(0)
		nCPU := runtime.NumCPU()
		if nProcs < nCPU {
			runtime.GOMAXPROCS(nCPU)
		}
	}

	cfg, err := config.LoadFile(*cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config file '%v': %v\n", *cfgFile, err)
		os.Exit(-1)
	}
	if *genOnly && !cfg.Debug.GenerateOnly {
		cfg.Debug.GenerateOnly = true
	}
	if *testConfig {
		fmt.Printf("The Meson server configuration looks good.\n")
		os.Exit(0)
	}

	// Setup the signal handling.
	haltCh := make(chan os.Signal)
	signal.Notify(haltCh, os.Interrupt, syscall.SIGTERM) // nolint

	rotateCh := make(chan os.Signal)
	signal.Notify(rotateCh, syscall.SIGHUP) // nolint

	// Start up the server.
	svr, err := server.New(cfg)
	if err != nil {
		if err == server.ErrGenerateOnly {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Failed to spawn server instance: %v\n", err)
		os.Exit(-1)
	}
	defer svr.Shutdown()

	// Halt the server gracefully on SIGINT/SIGTERM.
	go func() {
		<-haltCh
		svr.Shutdown()
	}()

	// Rotate server logs upon SIGHUP.
	go func() {
		<-rotateCh
		svr.RotateLog()
	}()

	// Wait for the server to explode or be terminated.
	svr.Wait()
}
