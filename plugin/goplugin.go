/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
package main

import (
	"context"

	log "github.com/cihub/seelog"

	"github.com/samuelkarp/purple-docker/plugin/config"
)

// #cgo pkg-config: glib-2.0 purple
// #include "stdlib.h"
// #include "account.h"
// #include "blist.h"
// #include "conversation.h"
// void test_call(char *test);
import "C"

// InitializeGoPlugin lets the plugin be initialized
//export InitializeGoPlugin
func InitializeGoPlugin() {
	config.SetupLogger()
}

// TeardownGoPlugin lets the plugin exit gracefully
//export TeardownGoPlugin
func TeardownGoPlugin() {
	log.Flush()
}

// DebugLog logs messages from C
//export DebugLog
func DebugLog(cmessage *C.char) {
	message := C.GoString(cmessage)
	log.Debug(message)
}

// Login starts the main event loop of the plugin.  The plugin's event loop is
// distinct from that of libpurple and calls should not be made synchronously
// from the plugin's event loop to libpurple.
//export Login
func Login(cAccount *C.PurpleAccount) {
	account := NewAccount(context.Background(), cAccount)
	account.SetConnected()
	account.ListenForContainerEvents()
	account.ScanContainers()
	EventLoop(cAccount)
}

// Close closes an account and cancels the context
//export Close
func Close(cConnection *C.PurpleConnection) {
	account, ok := GetAccount(cConnection.account)
	if !ok {
		log.Error("purple-docker: cannot close unknown account")
		return
	}

	account.Cancel()
}

// EventLoop is the hook into libpurple's event loop and is suitable for
// invoking libpurple's exported functions.  goroutines should *not* be started
// from within EventLoop.
//export EventLoop
func EventLoop(cAccount *C.PurpleAccount) {
	account, ok := GetAccount(cAccount)
	if !ok {
		log.Error("purple-docker: cannot run callbacks for unknown account")
		return
	}
	account.EventLoop()
}

// SendIM sends a message to a named identity in the plugin.
//export SendIM
func SendIM(cConnection *C.PurpleConnection, cWho *C.char, cMessage *C.char, flags C.PurpleMessageFlags) int32 {
	who := C.GoString(cWho)
	message := C.GoString(cMessage)
	log.Debugf("Send message to %s: %s", who, message)
	account, ok := GetAccount(cConnection.account)
	if !ok {
		return int32(0)
	}
	err := account.SendIM(who, message)
	if err != nil {
		log.Errorf("Cannot send message: %v", err)
		return int32(0)
	}
	return int32(len(message))
}

// NOTE(samuelkarp) main() is required so buildmode=c-shared works
func main() {}

//TODO RemoveBuddy and RemoveBuddies - these need to be handled because we're keeping pointers
//TODO Logout
