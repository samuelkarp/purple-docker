/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

// This file is a hack in order to generate a header file that we need in the C
// code.

package main

// #cgo pkg-config: glib-2.0 purple
// #include "stdlib.h"
// #include "account.h"
// #include "blist.h"
import "C"

// NOTE(samuelkarp) main() is required so buildmode=c-shared works
func main() {}

//export InitializeGoPlugin
func InitializeGoPlugin() {
}

//export TeardownGoPlugin
func TeardownGoPlugin() {}

//export DebugLog
func DebugLog(cmessage *C.char) {}

// Login starts the polling loop
//export Login
func Login(cAccount *C.PurpleAccount) {}

// Close closes an account and cancels the context
//export Close
func Close(cConnection *C.PurpleConnection) {}

// EventLoop is the main event loop of the plugin, suitable for invoking functions on libpurple
//export EventLoop
func EventLoop(cAccount *C.PurpleAccount) {}

//export SendIM
func SendIM(cConnection *C.PurpleConnection, who *C.char, message *C.char, flags C.PurpleMessageFlags) int32 {
	return 0
}
