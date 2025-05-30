// sickchat - A simple terminal chat client and server
// Copyright (C) 2025 Andrew Souza
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"github.com/drewslam/sickchat/common"
	"github.com/drewslam/sickchat/terminalClient/ui"
)

const (
	HOST = common.HOST
	PORT = common.PORT
	TYPE = common.TYPE
)

func main() {
	app := ui.NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
