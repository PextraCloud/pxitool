/*
Copyright 2025 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// ReadPassword reads a password from the terminal without echo or from stdin.
// https://github.com/containers/common/pull/1312/files#diff-9c54be49c38899332e913eaa6abfc98bd8bf21ff6a819614e2ef6cd093da9f92R96
func ReadPassword(fd int) ([]byte, error) {
	// Not a terminal, read from stdin with echo
	if !term.IsTerminal(fd) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Fprint(os.Stderr, "Password: ")
		pw, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		// remove trailing newline
		return []byte(pw[:len(pw)-1]), nil
	}

	// Store and restore the terminal status on interruptions to
	// avoid that the terminal remains in the password state
	// This is necessary as for https://github.com/golang/go/issues/31180

	oldState, err := term.GetState(fd)
	if err != nil {
		return make([]byte, 0), err
	}

	type Buffer struct {
		Buffer []byte
		Error  error
	}
	errorChannel := make(chan Buffer, 1)

	// SIGINT and SIGTERM restore the terminal, otherwise the no-echo mode would remain intact
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(interruptChannel)
		close(interruptChannel)
	}()
	go func() {
		for range interruptChannel {
			if oldState != nil {
				_ = term.Restore(fd, oldState)
			}
			errorChannel <- Buffer{Buffer: make([]byte, 0), Error: fmt.Errorf("interrupted")}
		}
	}()

	go func() {
		buf, err := term.ReadPassword(fd)
		errorChannel <- Buffer{Buffer: buf, Error: err}
	}()

	buf := <-errorChannel
	return buf.Buffer, buf.Error
}
