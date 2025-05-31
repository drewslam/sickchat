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
// This software uses the tview and tcell libraries.
// tview is licensed under the MIT License.
// tcell is licensed under the Apache License, Version 2.0.

package ui

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/drewslam/sickchat/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Application struct {
	App        *tview.Application
	ChatWindow *tview.TextView
	UserList   *tview.List
	UserInput  *tview.InputField
	Conn       net.Conn
	ClientID   string
}

func NewApp() *Application {
	conn, err := net.Dial("tcp", common.HOST+":"+common.PORT)
	if err != nil {
		panic(err)
	}

	app := &Application{
		App:        tview.NewApplication(),
		ChatWindow: tview.NewTextView(),
		UserList:   tview.NewList(),
		UserInput:  tview.NewInputField(),
		Conn:       conn,
	}

	// yourID := conn.LocalAddr().String()

	// Configure input field
	app.UserInput.
		SetBorder(true).
		SetTitle("Input")
	app.UserInput.
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault)

	// Configure chat window
	app.ChatWindow.
		SetBorder(true).
		SetTitle("Chat")
	app.ChatWindow.
		SetScrollable(true).
		SetChangedFunc(func() {
			app.App.Draw()
		})

	// Configure user list
	app.UserList.
		SetBorder(true).
		SetTitle("Users")

	// Create layout
	mainColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.ChatWindow, 0, 1, false).
		AddItem(app.UserInput, 3, 0, true)

	rootFlex := tview.NewFlex().
		AddItem(mainColumn, 0, 4, true).
		AddItem(app.UserList, 0, 1, false)

	// Input handling
	app.UserInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			message := app.UserInput.GetText()
			if message == "/quit" {
				app.Conn.Close()
				app.App.Stop()
				return
			}

			// send to server
			if message != "" {
				fmt.Fprintf(app.Conn, "%s\n", message)
				app.UserInput.SetText("")
			}
		}
	})

	// Message reader goroutine
	go func() {
		reader := bufio.NewReader(app.Conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				app.App.QueueUpdateDraw(func() {
					fmt.Fprintln(app.ChatWindow, "Disconnected.")
				})
				return
			}
			app.App.QueueUpdateDraw(func() {
				msg = strings.TrimSpace(msg)

				switch {
				case strings.HasPrefix(msg, "ID:"):
					id := strings.TrimPrefix(msg, "ID:")
					app.ClientID = id
					// app.updateUserList([]string{id}) // show local ID as "You"
				case strings.HasPrefix(msg, "USERS:"):
					raw := strings.TrimPrefix(msg, "USERS:")
					if raw != "" {
						ids := strings.Split(raw, ",")
						var filtered []string
						for _, id := range ids {
							if id != app.ClientID {
								filtered = append(filtered, id)
							}
						}
						app.updateUserList(filtered)
					} else {
						app.updateUserList([]string{})
					}

				default:
					// Display message
					fmt.Fprintln(app.ChatWindow, msg)
				}

				// Try to parse sender from msg
				if i := strings.Index(msg, ":"); i != -1 {
					sender := msg[:i]
					if sender != "USERS" && sender != "ID" {
						// found := false
						for i := 0; i < app.UserList.GetItemCount(); i++ {
							name, _ := app.UserList.GetItemText(i)
							if name == sender {
								// found = true
								break
							}
						}
						// if !found {
						//	app.UserList.AddItem(sender, "", 0, nil)
						// }
					}
				}
			})
		}
	}()

	app.App.SetRoot(rootFlex, true)
	app.App.SetFocus(app.UserInput)

	return app
}

func (a *Application) Run() error {
	return a.App.Run()
}

func (a *Application) updateUserList(ids []string) {
	a.UserList.Clear()

	if a.ClientID != "" {
		a.UserList.AddItem(fmt.Sprintf("User %s", a.ClientID), "", 0, nil)
	}

	for _, id := range ids {
		if id == "" || id == a.ClientID {
			continue
		}
		a.UserList.AddItem(fmt.Sprintf("User %s", id), "", 0, nil)
	}
}
