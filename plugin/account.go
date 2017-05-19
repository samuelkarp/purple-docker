/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
package main

import (
	"sync"
	"unsafe"

	"context"

	log "github.com/cihub/seelog"
	"github.com/fsouza/go-dockerclient"
)

// #cgo pkg-config: purple
// #include "account.h"
// #include "blist.h"
import "C"

// Account is a go struct encapsulating a PurpleAccount
type Account struct {
	ctx          context.Context
	cancel       context.CancelFunc
	online       bool
	cAccount     *C.PurpleAccount
	dockerClient *docker.Client
	buddies      map[string]*C.PurpleBuddy
	containers   map[string]*Container

	lock sync.RWMutex

	funcQueue     []func()
	funcQueueLock sync.Mutex
}

var accountMap map[*C.PurpleAccount]*Account
var accountMapLock sync.RWMutex

// NewAccount returns an Account from a PurpleAccount
func NewAccount(ctx context.Context, cAccount *C.PurpleAccount) *Account {
	accountMapLock.Lock()
	defer accountMapLock.Unlock()
	if accountMap == nil {
		accountMap = map[*C.PurpleAccount]*Account{}
	}
	if account, ok := accountMap[cAccount]; ok {
		return account
	}
	accountCtx, cancel := context.WithCancel(ctx)
	account := &Account{
		ctx:        accountCtx,
		cancel:     cancel,
		online:     true,
		cAccount:   cAccount,
		buddies:    map[string]*C.PurpleBuddy{},
		containers: map[string]*Container{},
		funcQueue:  []func(){},
	}
	accountMap[cAccount] = account
	return account
}

func GetAccount(cAccount *C.PurpleAccount) (*Account, bool) {
	accountMapLock.RLock()
	defer accountMapLock.RUnlock()
	account, ok := accountMap[cAccount]
	return account, ok
}

func (account *Account) Cancel() {
	account.cancel()
	accountMapLock.Lock()
	defer accountMapLock.Unlock()
	delete(accountMap, account.cAccount)
}

func (account *Account) EventLoop() {
	account.funcQueueLock.Lock()
	localQueue := account.funcQueue
	account.funcQueue = []func(){}
	account.funcQueueLock.Unlock()

	select {
	case <-account.ctx.Done():
	default:
		for _, function := range localQueue {
			function()
		}
	}
}

func (account *Account) enqueueFunction(function func()) {
	account.funcQueueLock.Lock()
	defer account.funcQueueLock.Unlock()

	account.funcQueue = append(account.funcQueue, function)
}

// SetConnected sets the account state to connected
func (account *Account) SetConnected() {
	purpleConnection := C.purple_account_get_connection(account.cAccount)
	C.purple_connection_set_state(purpleConnection, C.PURPLE_CONNECTED)
}

// IsConnected returns whether the account is connected
func (account *Account) IsConnected() bool {
	connected := C.purple_account_is_connected(account.cAccount)
	return connected != 0
}

func (account *Account) ReceiveIM(from, message string) {
	receiveTime := C.time(nil)
	account.enqueueFunction(func() {
		account.lock.RLock()
		defer account.lock.RUnlock()
		if !account.online {
			return
		}
		cFrom := C.CString(from)
		defer C.free(unsafe.Pointer(cFrom))
		cMessage := C.CString(message)
		defer C.free(unsafe.Pointer(cMessage))
		conversation := C.purple_find_conversation_with_account(C.PURPLE_CONV_TYPE_IM, cFrom, account.cAccount)
		if conversation == nil {
			log.Tracef("creating new conversation with from %s", from)
			conversation = C.purple_conversation_new(C.PURPLE_CONV_TYPE_IM, account.cAccount, cFrom)
			if conversation == nil {
				log.Error("conversation is nil")
				return
			}
		}
		log.Tracef("Got conversation %p", conversation)
		C.purple_conversation_write(conversation, cFrom, cMessage, C.PURPLE_MESSAGE_RECV, receiveTime)
		log.Tracef("Called purple_conversation_write")
	})
}

func (account *Account) SendIM(to, message string) error {
	account.lock.RLock()
	container := account.containers[to]
	account.lock.RUnlock()
	return container.ToStdinAttached(message)
}
