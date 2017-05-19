/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
#define PURPLE_PLUGINS

#include <stdlib.h>
#include <glib.h>

#include "notify.h"
#include "plugin.h"
#include "prpl.h"
#include "accountopt.h"
#include "version.h"

#include "goplugin.h"


static gboolean plugin_load(PurplePlugin *plugin) {
    return TRUE;
}

/**
 * Called to get the icon name for the given buddy and account.
 *
 * If buddy is NULL and the account is non-NULL, it will return the
 * name to use for the account's icon. If both are NULL, it will
 * return the name to use for the protocol's icon.
 *
 * For now, everything just uses the 'docker' icon.
 *
 * This function is required for the plugin to be loadable.
 */
static const char* dockerprpl_list_icon(PurpleAccount *acct, PurpleBuddy *buddy)
{
    return "docker";
}

gboolean dockerprpl_poll(gpointer data) {
    PurpleAccount *acct = data;
    EventLoop(acct);
    // return TRUE so this continues to get called every second
    return TRUE;
}

/**
 * Start the connection to Docker
 */
void dockerprpl_login(PurpleAccount *acct) {
  Login(acct);
  // main event loop trigger, starting at 1-second granularity for now
  purple_timeout_add_seconds(1, dockerprpl_poll, acct);
}

/**
 * Called to handle closing the connection to Docker
 */
static void dockerprpl_close(PurpleConnection *pc) {
  Close(pc);
}

/**
 * Called to get a list of the PurpleStatusType which are valid for this account
 *
 * PURPLE_STATUS_UNAVAILABLE - Created container
 * PURPLE_STATUS_AVAILABLE - Running container
 * PURPLE_STATUS_AWAY - Paused container
 * PURPLE_STATUS_OFFLINE - Stopped container
 */
static GList* dockerprpl_status_types(PurpleAccount *acct)
{
    GList *types = NULL;
    PurpleStatusType *type;

    type = purple_status_type_new_full(PURPLE_STATUS_OFFLINE, NULL, NULL, TRUE, TRUE, FALSE);
    types = g_list_prepend(types, type);

    type = purple_status_type_new_full(PURPLE_STATUS_AVAILABLE, NULL, NULL, TRUE, TRUE, FALSE);
    types = g_list_prepend(types, type);

    return types;
}

/**
 * Called to send a message
 */
int dockerprpl_send_im(PurpleConnection *pc, const gchar *who, const gchar *message, PurpleMessageFlags flags){
  // Because Go doesn't handle const, we need to copy the strings.  There will be even more copying in the Go source
  // too to convert to a go string.
  char* whocopy = g_strdup(who);
  char* messagecopy = g_strdup(message);
  int retval = SendIM(pc, whocopy, messagecopy, flags);
  g_free(whocopy);
  g_free(messagecopy);
  return retval;
}

void dockerprpl_tooltip_text(PurpleBuddy *buddy, PurpleNotifyUserInfo *user_info, gboolean full)
{
    DebugLog("dockerprpl_tooltip_text");
}

void dockerprpl_get_info(PurpleConnection *pc, const gchar *uid) {
    DebugLog("dockerprpl_get_info");
}

gchar* dockerprpl_status_text(PurpleBuddy *buddy) {
    // trigger the event loop
    EventLoop(buddy->account);
    return NULL;
}

static PurplePluginProtocolInfo docker_protocol_info = {
    /* options */
    OPT_PROTO_NO_PASSWORD,             /*| OPT_PROTO_SLASH_COMMANDS_NATIVE, */
    NULL,                              /* user_splits */
    NULL,                              /* protocol_options */
    {                                  /* icon_spec, a PurpleBuddyIconSpec */
     "png,jpg,gif",                    /* format */
     0,                                /* min_width */
     0,                                /* min_height */
     128,                              /* max_width */
     128,                              /* max_height */
     10000,                            /* max_filesize */
     PURPLE_ICON_SCALE_DISPLAY,        /* scale_rules */
     },
    dockerprpl_list_icon,              /* list_icon */
    NULL,                              /* list_emblems */
    dockerprpl_status_text,            /* status_text */
    dockerprpl_tooltip_text,           /* tooltip_text */
    dockerprpl_status_types,           /* status_types */
    NULL,                              /* blist_node_menu */
    NULL,                              /* chat_info */
    NULL,                              /* chat_info_defaults */
    dockerprpl_login,                  /* login */
    dockerprpl_close,                  /* close */
    dockerprpl_send_im,                /* send_im */
    NULL,                              /* set_info */
    NULL,                              /* send_typing */
    dockerprpl_get_info,               /* get_info */
    NULL,                              /* set_status */
    NULL,                              /* set_idle */
    NULL,                              /* change_passwd */
    NULL,                              /* add_buddy */
    NULL,                              /* add_buddies */
    NULL,                              /* remove_buddy */
    NULL,                              /* remove_buddies */
    NULL,                              /* add_permit */
    NULL,                              /* add_deny */
    NULL,                              /* rem_permit */
    NULL,                              /* rem_deny */
    NULL,                              /* set_permit_deny */
    NULL,                              /* join_chat */
    NULL,                              /* reject chat invite */
    NULL,                              /* get_chat_name */
    NULL,                              /* chat_invite */
    NULL,                              /* chat_leave */
    NULL,                              /* chat_whisper */
    NULL,                              /* chat_send */
    NULL,                              /* keepalive */
    NULL,                              /* register_user */
    NULL,                              /* get_cb_info */
    NULL,                              /* get_cb_away */
    NULL,                              /* alias_buddy */
    NULL,                              /* group_buddy */
    NULL,                              /* rename_group */
    NULL,                              /* buddy_free */
    NULL,                              /* convo_closed */
    purple_normalize_nocase,           /* normalize */
    NULL,                              /* set_buddy_icon */
    NULL,                              /* remove_group */
    NULL,                              /* get_cb_real_name */
    NULL,                              /* set_chat_topic */
    NULL,                              /* find_blist_chat */
    NULL,                              /* roomlist_get_list */
    NULL,                              /* roomlist_cancel */
    NULL,                              /* roomlist_expand_category */
    NULL,                              /* can_receive_file */
    NULL,                              /* send_file */
    NULL,                              /* new_xfer */
    NULL,                              /* offline_message */
    NULL,                              /* whiteboard_prpl_ops */
    NULL,                              /* send_raw */
    NULL,                              /* roomlist_room_serialize */
    NULL,                              /* unregister_user */
    NULL,                              /* send_attention */
    NULL,                              /* attention_types */
    sizeof(PurplePluginProtocolInfo),  /* struct_size */
    NULL,                              /* get_account_text_table */
    NULL,                              /* initiate_media */
    NULL,                              /* get_media_caps */
#if PURPLE_MAJOR_VERSION > 1
#if PURPLE_MINOR_VERSION > 6
    NULL,                              /* get_moods */
    NULL,                              /* set_public_alias */
    NULL,                              /* get_public_alias */
#if PURPLE_MINOR_VERSION > 7
    NULL,                              /* add_buddy_with_invite */
    NULL,                              /* add_buddies_with_invite */
#endif /* PURPLE_MINOR_VERSION > 7 */
#endif /* PURPLE_MINOR_VERSION > 6 */
#endif /* PURPLE_MAJOR_VERSION > 1 */
};

static void init_plugin(PurplePlugin *plugin) {
    InitializeGoPlugin();
    GList *protocol_options = NULL;

    protocol_options = g_list_append(protocol_options,
        purple_account_option_string_new(
                "Home server URL", "test_option",
                "default_value"));

    docker_protocol_info.protocol_options = protocol_options;
}

static gboolean plugin_unload(PurplePlugin *plugin) {
    TeardownGoPlugin();
    return TRUE;
}

static PurplePluginInfo info = {
    PURPLE_PLUGIN_MAGIC,    /* Plugin magic, this must always be
                               PURPLE_PLUGIN_MAGIC.*/
    PURPLE_MAJOR_VERSION,   /* This is also defined in libpurple.  It helps
                               libpurple's plugin system determine which version
                               of libpurple this plugin was compiled for, and
                               whether loading it will cause problems. */
    PURPLE_MINOR_VERSION,   /* See previous */
    PURPLE_PLUGIN_PROTOCOL, /* PurplePluginType */
    NULL,                   /* This field is the UI requirement.  If you're writing
                               a core plugin, this must be NULL and the plugin must
                               not contain any UI code. */
    0,                      /* This field is for plugin flags.  Currently, the only
                               flag available to plugins is invisible
                               (PURPLE_PLUGIN_FLAG_INVISIBLE). It causes the plugin
                               NOT to appear in the list of plugins. */
    NULL,                   /* This is a GList of plugin dependencies.  In other words,
                               it's a GList of plugin id's that your plugin depends on.
                               Set this value to NULL no matter what.  If your plugin
                               has dependencies, set them at run-time in the
                               plugin_init function. */
    PURPLE_PRIORITY_DEFAULT,/* This is the priority libpurple will give your plugin.
                               There are three possible values for this field,
                               PURPLE_PRIORITY_DEFAULT, PURPLE_PRIORITY_HIGHEST, and
                               PURPLE_PRIORITY_LOWEST. */
    "prpl-samuelkarp-purple-docker", /* This is your plugin's id.  There is a whole page dedicated
                               to this in the Related Pages section of the API docs. */
    "Docker (purple-docker)",/* This is your plugin's name.  This is what will be
                               displayed for your plugin in the UI. */
    "0.0.1",                /* This is the version of your plugin. */
    "Docker protocol",      /* This is the summary of your plugin.  It should be a short
                               blurb.  The UI determines where, if at all, to display
                               this. */
    "Interact with Docker containers",/* This is the summary of your plugin.  It should be a short
                               blurb.  The UI determines where, if at all, to display
                               this. */
    "Samuel Karp <purple-docker@samuelkarp.com>", /* This is where you can put your name and e-mail
                               address. */
    "http://github.com/samuelkarp/purple-docker", /* This is the website for the plugin. */
    plugin_load,            /* This is a pointer to a function for libpurple to call when
                               it is loading the plugin.  It should be of the type:

                                  gboolean plugin_load(PurplePlugin *plugin)

                               Returning FALSE will stop the loading of the plugin.
                               Anything else would evaluate as TRUE and the plugin will
                               continue to load. */
    plugin_unload,          /* Same as above except it is called when libpurple tries to
                               unload your plugin.  It should be of the type:

                                  gboolean plugin_unload(PurplePlugin *plugin)

                               Returning TRUE will tell libpurple to continue unloading
                               while FALSE will stop the unloading of your plugin. */
    NULL,                   /* Similar to the two above members, except this is called
                               when libpurple tries to destory the plugin.  This is
                               generally only called when for some reason or another the
                               plugin fails to probe correctly.  It should be of the type:

                                  void plugin_destroy(PurplePlugin *plugin) */

    NULL,                   /* This is a pointer to a UI-specific struct.  For a Pidgin
                               plugin it will be a pointer to a PidginPluginUiInfo
                               struct, for example. */
    &docker_protocol_info,  /* This is a pointer to either a PurplePluginLoaderInfo
                               struct or a PurplePluginProtocolInfo struct, and is
                               beyond the scope of this document. */
    NULL,                  /* This is a pointer to a PurplePluginUiInfo struct.  It is
                               a core/ui split way for core plugins to have a UI
                               configuration frame.  You can find an example of this
                               code in libpurple/plugins/pluginpref_example.c */
    NULL,                  /* This is a function pointer where you can define "plugin
                               actions".  The UI controls how they're displayed.  It
                               should be of the type:

                                  GList *function_name(PurplePlugin *plugin,
                                                       gpointer context)

                               It must return a GList of PurplePluginActions. */
    NULL,                  /* This is a pointer reserved for future use. */
    NULL,                  /* This is a pointer reserved for future use. */
    NULL,                  /* This is a pointer reserved for future use. */
    NULL                   /* This is a pointer reserved for future use. */
};

PURPLE_INIT_PLUGIN(docker, init_plugin, info)
