/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
package main

import (
	log "github.com/cihub/seelog"
	"github.com/fsouza/go-dockerclient"
)

func (account *Account) ListenForContainerEvents() {
	log.Debug("Listener starting")
	ctx := account.ctx
	client, ok := account.getDockerClient()
	if !ok {
		return
	}
	events := make(chan *docker.APIEvents)
	err := client.AddEventListener(events)
	if err != nil {
		log.Errorf("Unable to add a docker event listener: %v", err)
		return
	}
	go func() {
		<-ctx.Done()
		client.RemoveEventListener(events)
	}()
	go func() {
		for event := range events {
			// currently only container events type needs to be handled
			if event.Type != "container" || event.ID == "" {
				continue
			}

			containerID := event.ID
			log.Debugf("Got event from docker daemon: %s", event)

			dockerContainer, err := client.InspectContainer(containerID)
			if err != nil {
				log.Warnf("failed to inspect container %s", containerID)
				continue
			}

			name := dockerContainer.Name[1:] // strip the / that Docker puts as the first character

			switch event.Status {
			case "create":
				account.NewContainer(dockerContainer)
			case "start":
				account.handleContainerStart(containerID, name)
			case "die":
				account.handleContainerDie(containerID, name)
			}
		}
	}()
}

func (account *Account) handleContainerStart(id, name string) {
	account.lock.RLock()
	defer account.lock.RUnlock()
	container, ok := account.containers[name]
	if !ok {
		log.Warnf("ignoring start event for unknown container %s %s", id, name)
		return
	}
	account.SetBuddyAvailable(name, true)
	container.Attach()
}

func (account *Account) handleContainerDie(id, name string) {
	account.lock.Lock()
	defer account.lock.Unlock()
	container, ok := account.containers[name]
	if !ok {
		log.Warnf("ignoring die event for unknown container %s %s", id, name)
		return
	}
	container.cancel()
	delete(account.containers, name)
	account.SetBuddyAvailable(name, false)
}

func (account *Account) ScanContainers() {
	client, ok := account.getDockerClient()
	if !ok {
		return
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.Error("purple-docker: cannot list containers")
	}
	for _, i := range containers {
		dockerContainer, err := client.InspectContainer(i.ID)
		if err != nil {
			log.Warnf("failed to inspect container %s", i.ID)
			continue
		}

		name := dockerContainer.Name[1:] // strip the / that Docker puts as the first character
		container := account.NewContainer(dockerContainer)
		account.SetBuddyAvailable(name, true)
		container.Attach()
	}
}

func (account *Account) getDockerClient() (*docker.Client, bool) {
	account.lock.Lock()
	defer account.lock.Unlock()
	if account.dockerClient == nil {
		client, err := docker.NewVersionedClientFromEnv("1.24")
		if err != nil {
			log.Error("purple-docker: cannot create Docker client")
			return nil, false
		}
		account.dockerClient = client
	}
	return account.dockerClient, true
}
