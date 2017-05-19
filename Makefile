# Copyright (c) 2017, Samuel Karp.  All rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

ROOT := $(shell pwd)

SOURCEDIR = ./plugin
GOSOURCES = $(shell find $(SOURCEDIR) -name '*.go')
CSOURCES = $(shell find $(SOURCEDIR) -name '*.c')
HEADERBASE = $(SOURCEDIR)/goplugin
HEADER = $(HEADERBASE).h
LIBS = glib-2.0 purple
LIBFLAGS = $(shell pkg-config --cflags $(LIBS) --libs $(LIBS))
TARGET = purpledockerplugin.so

all: build

.PHONY: build
build: $(TARGET)
$(TARGET): $(GOSOURCES) $(CSOURCES) $(HEADER)
	go build -buildmode=c-shared -o $(TARGET) $(SOURCEDIR)
	rm purpledockerplugin.h

.PHONY: header
header: $(HEADER)
$(HEADER): $(GOSOURCES)
	@# This is a hack to get a header file with exported symbols callable from C
	go build -buildmode=c-shared -o $(HEADERBASE).so $(SOURCEDIR)/purple/main/main.go
	rm $(HEADERBASE).so

.PHONY: localinstall
localinstall: build
	mkdir -p ~/.purple/plugins/
	cp $(TARGET) ~/.purple/plugins/

.PHONY: clean
clean:
	rm $(TARGET) ||:
	rm $(HEADERBASE).h ||:
