#
# * Produced: Sun July 13, 2025
# * Author: Alec M.
# * GitHub: https://amattu.com/links/github
# * Copyright: (C) 2025 Alec M.
# * License: License GNU Affero General Public License v3.0
# *
# * This program is free software: you can redistribute it and/or modify
# * it under the terms of the GNU Affero General Public License as published by
# * the Free Software Foundation, either version 3 of the License, or
# * (at your option) any later version.
# *
# * This program is distributed in the hope that it will be useful,
# * but WITHOUT ANY WARRANTY; without even the implied warranty of
# * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# * GNU Affero General Public License for more details.
# *
# * You should have received a copy of the GNU Affero General Public License
# * along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

#
# Variables
#
build_args = -a -o

#
# Targets
#

# Generate build and run tests
all: build tests

# Generate multi-platform builds
build: build_linux build_freebsd build_windows build_mac

# Build for Linux
build_linux:
	@echo "Building for Linux"
	GOOS=linux GOARCH=386 go build $(build_args) bin/linux

# Build for FreeBSD
build_freebsd:
	@echo "Building for FreeBSD"
	GOOS=freebsd GOARCH=386 go build $(build_args) bin/freebsd

# Build for Windows
build_windows:
	@echo "Building for windows"
	GOOS=windows GOARCH=386 go build $(build_args) bin/windows.exe

# Build for macOS platforms
build_mac:
	@echo "Building for macOS x86 chipset"
	GOOS=darwin GOARCH=amd64 go build $(build_args) bin/macos-amd64
	@echo "Building for macOS arm chipset"
	GOOS=linux GOARCH=arm go build $(build_args) bin/macos-arm

# Run tests
tests:
	go test -v -cover ./...

# Clean and reset workspace
clean:
	rm -rf ./bin
