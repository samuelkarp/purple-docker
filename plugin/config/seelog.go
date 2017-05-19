/* Copyright (c) 2017, Samuel Karp.  All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */
package config

import log "github.com/cihub/seelog"

func SetupLogger() {
	logger, err := log.LoggerFromConfigAsString(loggerConfig())
	if err == nil {
		log.ReplaceLogger(logger)
	} else {
		log.Error(err)
	}
}

func loggerConfig() string {
	config := `
	<seelog type="asyncloop" minlevel="debug">
		<outputs formatid="main">
			<filter levels="debug,info,warn,error,critical">
				<console />
			</filter>
		</outputs>
		<formats>
			<format id="main" format="%UTCDate(2006-01-02T15:04:05Z07:00) [%LEVEL] %Msg%n" />
		</formats>
	</seelog>
`
	return config
}
