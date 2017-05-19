/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
package main

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"
	"unicode"

	log "github.com/cihub/seelog"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

// #cgo pkg-config: purple
// #include "account.h"
// #include "blist.h"
import "C"

type Container struct {
	ctx         context.Context
	cancel      context.CancelFunc
	id          string
	name        string
	interactive bool // interactive containers have "attach", noninteractive have "exec"
	account     *Account

	purpleBuddy *C.PurpleBuddy
	lock        sync.RWMutex

	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader
}

// TODO figure out some decent behavior here; note that a PurpleBuddy should
// only be allocated on the event loop thread.

func (account *Account) NewContainer(dockerContainer *docker.Container) *Container {
	name := dockerContainer.Name[1:] // strip the / that Docker puts as the first character
	interactive := dockerContainer.Config.OpenStdin
	containerCtx, cancel := context.WithCancel(account.ctx)
	container := &Container{
		ctx:         containerCtx,
		cancel:      cancel,
		id:          dockerContainer.ID,
		name:        name,
		interactive: interactive,
		account:     account,
	}

	account.lock.Lock()
	account.containers[name] = container
	account.lock.Unlock()
	account.enqueueFunction(func() {
		container.lock.Lock()
		defer container.lock.Unlock()
		container.purpleBuddy = account.addBuddyUnsafe(name, "containers", false)
	})
	return container
}

func (container *Container) Attach() {
	container.lock.Lock()
	defer container.lock.Unlock()
	if container.interactive {
		log.Debugf("attaching to interactive container %s", container.name)
		// hook 'em up
		stdinReader, stdinWriter := io.Pipe()
		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()
		container.stdin = stdinWriter
		container.stdout = stdoutReader
		container.stderr = stderrReader
		go container.monitor(stdinReader, stdoutWriter, stderrWriter)
		go container.receive()
	}

}

// monitor attaches to a container, hooking up stdin, stdout, and stderr
func (container *Container) monitor(stdin io.Reader, stdout io.Writer, stderr io.Writer) {
	client, ok := container.account.getDockerClient()
	if !ok {
		return
	}
	err := client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    container.id,
		InputStream:  stdin,
		OutputStream: stdout,
		ErrorStream:  stderr,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
	})
	if err != nil {
		log.Error(err)
	}
	log.Debugf("detached from container %s", container.name)
}

// receive reads stdout and stderr to send them to the account
func (container *Container) receive() {
	stdout := bufio.NewReader(container.stdout)
	stderr := bufio.NewReader(container.stderr)
	stdoutChan := readStringChan(stdout)
	stderrChan := readStringChan(stderr)
	log.Debugf("receiving from container %s", container.name)
	for {
		select {
		case <-container.ctx.Done():
			return
		case stdoutLine := <-stdoutChan:
			container.account.ReceiveIM(container.name, stdoutLine)
		case stderrLine := <-stderrChan:
			container.account.ReceiveIM(container.name, stderrLine)
		}
	}
}

func readStringChan(reader *bufio.Reader) <-chan string {
	channel := make(chan string)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Trace("detaching from stream")
				if err != io.EOF {
					log.Error("stream ended unexpectedly")
				}
				close(channel)
			}
			channel <- strings.TrimRightFunc(line, unicode.IsSpace)
		}
	}()
	return channel
}

// ToStdinAttached writes to the stdin of the attached container
func (container *Container) ToStdinAttached(content string) error {
	if !container.interactive {
		return errors.Errorf("stdin: container not interactive: %s", container.name)
	}
	toWrite := content + "\n"
	written, err := container.stdin.Write([]byte(toWrite))
	if written != len(toWrite) {
		return errors.Errorf("stdin: could not write bytes: %s (%d written)", container.name, written)
	}
	return err
}
