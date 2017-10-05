// cjdnsui - Graphical user interface for Cjdns
// Copyright (C) 2017  William Wennerström
//
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"github.com/willeponken/cjdnsui/patterns"
)

const (
	settingsSaveTopic = iota
)

type Status struct {
	CjdnsIp   string
	PublicKey string
	Port      int
}

type Settings struct {
	AuthorizedPasswords []string
	AdminAddress        string
	AdminPassword       string
}

//go:generate qtmoc
type statusWidget struct {
	widgets.QWidget

	_ func(status Status) error `slot:"set"`
	_ func() Status             `slot:"get"`
}

type settingsWidget struct {
	widgets.QWidget

	_ func()                        `signal:"save"`
	_ func(settings Settings) error `slot:"set"`
	_ func() Settings               `slot:"get"`
}

func newSelectableQLabel(label string) *widgets.QLabel {
	widget := widgets.NewQLabel2(label, nil, 0)
	widget.SetTextInteractionFlags(core.Qt__TextSelectableByMouse)
	widget.SetCursor(gui.NewQCursor2(core.Qt__IBeamCursor))
	return widget
}

func newStatusWidget() *statusWidget {
	widget := NewStatusWidget(nil, 0)

	mainVBox := widgets.NewQVBoxLayout2(widget)

	localInfoGroup := widgets.NewQGroupBox2("Local node", nil)
	localInfoGroup.SetSizePolicy2(widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Minimum)
	localInfoGroup.SetMinimumWidth(500)

	mainVBox.AddWidget(localInfoGroup, 0, core.Qt__AlignLeft)

	localInfoForm := widgets.NewQFormLayout(nil)
	localInfoGroup.SetLayout(localInfoForm)

	cjdnsIpLabel := newSelectableQLabel("Unknown")
	publicKeyLabel := newSelectableQLabel("Unknown")
	portLabel := newSelectableQLabel("Unknown")

	localInfoForm.AddRow3("Cjdns IP:", cjdnsIpLabel)
	localInfoForm.AddRow3("Public Key:", publicKeyLabel)
	localInfoForm.AddRow3("Port:", portLabel)

	currentPeersGroup := widgets.NewQGroupBox2("Current peers", nil)
	currentPeersGroup.SetSizePolicy2(widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)
	currentPeersGroup.SetMinimumWidth(500)

	mainVBox.AddWidget(currentPeersGroup, 0, core.Qt__AlignLeft)

	widget.ConnectSet(func(status Status) error {
		cjdnsIpLabel.SetText(status.CjdnsIp)
		publicKeyLabel.SetText(status.PublicKey)
		portLabel.SetText(string(status.Port))

		return nil
	})

	widget.ConnectGet(func() Status {
		port, err := strconv.Atoi(portLabel.Text())
		if err != nil {
			panic(err)
		}

		return Status{
			CjdnsIp:   cjdnsIpLabel.Text(),
			PublicKey: publicKeyLabel.Text(),
			Port:      port,
		}
	})

	return widget
}

func newSettingsWidget() *settingsWidget {
	widget := NewSettingsWidget(nil, 0)

	mainVBox := widgets.NewQVBoxLayout2(widget)

	adminLoginGroup := widgets.NewQGroupBox2("Administration login", nil)
	adminLoginGroup.SetSizePolicy2(widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Minimum)
	adminLoginGroup.SetMinimumWidth(500)
	authPasswordGroup := widgets.NewQGroupBox2("Authorized passwords", nil)
	authPasswordGroup.SetSizePolicy2(widgets.QSizePolicy__Maximum, widgets.QSizePolicy__Expanding)
	authPasswordGroup.SetMinimumWidth(500)

	mainVBox.AddWidget(adminLoginGroup, 0, core.Qt__AlignLeft)
	mainVBox.AddWidget(authPasswordGroup, 1, core.Qt__AlignLeft)

	adminLoginForm := widgets.NewQFormLayout(nil)
	adminLoginGroup.SetLayout(adminLoginForm)

	adminAddressInput := widgets.NewQLineEdit(nil)
	adminAddressInput.SetMinimumWidth(200)
	adminPasswordInput := widgets.NewQLineEdit(nil)
	adminPasswordInput.SetEchoMode(widgets.QLineEdit__Password)
	adminPasswordInput.SetMinimumWidth(200)

	adminLoginForm.AddRow3("Address:", adminAddressInput)
	adminLoginForm.AddRow3("Password:", adminPasswordInput)

	authPasswordForm := widgets.NewQFormLayout(nil)
	authPasswordGroup.SetLayout(authPasswordForm)

	authPasswordTextEdit := widgets.NewQPlainTextEdit(nil)
	authPasswordTextEdit.SetSizePolicy2(widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Expanding)
	authPasswordTextEdit.SetPlaceholderText("Authorized passwords")
	authPasswordTextEdit.SetLineWrapMode(widgets.QPlainTextEdit__NoWrap)
	authPasswordTextEdit.SetMinimumSize2(300, 200)
	authPasswordForm.AddWidget(authPasswordTextEdit)

	saveButton := widgets.NewQPushButton2("Save", nil)
	mainVBox.AddWidget(saveButton, 0, core.Qt__AlignRight)

	saveButton.ConnectReleased(func() {
		widget.Save()
	})

	widget.ConnectSet(func(settings Settings) error {
		adminAddressInput.SetText(settings.AdminAddress)
		adminPasswordInput.SetText(settings.AdminPassword)

		authPasswordTextEdit.Clear()
		for _, password := range settings.AuthorizedPasswords {
			authPasswordTextEdit.AppendPlainText(fmt.Sprintf("%s\n", password))
		}

		return nil
	})

	widget.ConnectGet(func() Settings {
		buff := bufio.NewReader(strings.NewReader(authPasswordTextEdit.ToPlainText()))

		var authPasswords []string
		for {
			line, _, err := buff.ReadLine()
			if err == nil {
				password := strings.Map(func(r rune) rune {
					if unicode.IsSpace(r) {
						return -1
					}
					return r
				}, string(line))

				if len(password) > 0 {
					authPasswords = append(authPasswords, password)
				}

			} else if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		return Settings{
			AdminAddress:        adminAddressInput.Text(),
			AdminPassword:       adminPasswordInput.Text(),
			AuthorizedPasswords: authPasswords,
		}
	})

	return widget
}

type View struct {
	patterns.Observable

	status   *statusWidget
	settings *settingsWidget
}

func (view *View) SetStatus(status Status) {
	view.status.Set(status)
}

func (view *View) GetStatus() Status {
	return view.status.Get()
}

func (view *View) SetSettings(settings Settings) {
	view.settings.Set(settings)
}

func (view *View) GetSettings() Settings {
	return view.settings.Get()
}

func (view *View) Run() {
	widgets.NewQApplication(len(os.Args), os.Args)

	tabWidget := widgets.NewQTabWidget(nil)

	statusWidget := newStatusWidget()
	settingsWidget := newSettingsWidget()

	tabWidget.AddTab(statusWidget, "Status")
	tabWidget.AddTab(settingsWidget, "Settings")

	settingsWidget.ConnectSave(func() {
		view.NotifyObservers(settingsSaveTopic, nil)
	})

	tabWidget.Show()

	widgets.QApplication_Exec()
}

func NewView() *View {
	view := &View{}

	return view
}
