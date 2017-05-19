/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
package main

import (
	"unsafe"

	log "github.com/cihub/seelog"
)

// #cgo pkg-config: purple
/*
#include "account.h"
#include "blist.h"
#include "status.h"

// prpl_set_user_status is a shim because cgo can't call varadic functions
void prpl_set_user_status(PurpleAccount *account, const char* name, PurpleStatusPrimitive type) {
    const char* id = purple_primitive_get_id_from_type(type);
    purple_prpl_got_user_status(account, name, id, NULL);
}

int purple_buddy_is_online(PurpleBuddy* buddy) {
    return PURPLE_BUDDY_IS_ONLINE(buddy);
}
*/
import "C"

// AddTempBuddy adds a buddy to the buddy list of the account, but does not save
// the buddy.
func (account *Account) AddTempBuddy(name, group string, online bool) {
	account.enqueueFunction(func() {
		account.lock.Lock()
		defer account.lock.Unlock()
		if _, ok := account.buddies[name]; ok {
			return
		}
		purpleBuddy := account.addBuddyUnsafe(name, group, online)
		account.buddies[name] = purpleBuddy
	})
}

// addBuddyUnsafe adds the buddy to the buddy list and returns the PurpleBuddy.
// This method is unsafe as it must be run from the event loop thread.
func (account *Account) addBuddyUnsafe(name, group string, online bool) *C.PurpleBuddy {
	buddyName := C.CString(name)
	groupName := C.CString(group)
	defer C.free(unsafe.Pointer(buddyName))
	defer C.free(unsafe.Pointer(groupName))
	purpleBuddy := C.purple_buddy_new(account.cAccount, buddyName, nil)
	purpleGroup := C.purple_find_group(groupName)
	if purpleGroup == nil {
		log.Debugf("Creating new group %s", group)
		purpleGroup = C.purple_group_new(groupName)
	}
	C.purple_blist_add_buddy(purpleBuddy, nil, purpleGroup, nil)
	C.purple_blist_node_set_flags(&(purpleBuddy.node), C.PURPLE_BLIST_NODE_FLAG_NO_SAVE)
	if online {
		C.prpl_set_user_status(account.cAccount, buddyName, C.PURPLE_STATUS_AVAILABLE)
	}
	return purpleBuddy
}

func (account *Account) SetBuddyAvailable(name string, online bool) {
	account.enqueueFunction(func() {
		buddyName := C.CString(name)
		defer C.free(unsafe.Pointer(buddyName))
		if online {
			C.prpl_set_user_status(account.cAccount, buddyName, C.PURPLE_STATUS_AVAILABLE)
		} else {
			C.prpl_set_user_status(account.cAccount, buddyName, C.PURPLE_STATUS_OFFLINE)
		}
	})
}
