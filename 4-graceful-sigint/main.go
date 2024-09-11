// ////////////////////////////////////////////////////////////////////
//
// Given is a mock process which runs indefinitely and blocks the
// program. Right now the only way to stop the program is to send a
// SIGINT (Ctrl-C). Killing a process like that is not graceful, so we
// want to try to gracefully stop the process first.
//
// Change the program to do the following:
//  1. On SIGINT try to gracefully stop the process using
//     `proc.Stop()`
//  2. If SIGINT is called again, just kill the program (last resort)
package main

import (
	"os"
	"os/signal"
)

func main() {
	// Create a process
	proc := &MockProcess{}
	// Run the process (blocking)
	go proc.Run()

	// Channel to listen for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Channel to indicate forceful termination
	forceQuitChan := make(chan bool)

	go func() {
		<-sigChan // Listen for this channel
		go func() {
			<-sigChan
			forceQuitChan <- true
		}()
		proc.Stop()
	}()

	<-forceQuitChan // Listen for another signal
	os.Exit(0)

}
