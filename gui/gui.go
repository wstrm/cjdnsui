// cjdnsui - Graphical user interface for Cjdns
// Copyright (C) 2017  William Wennerström
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
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

func newStatusWidget() *statusWidget {
	widget := NewStatusWidget(nil, 0)

	mainVBox := widgets.NewQVBoxLayout2(widget)

	localInfoGroup := widgets.NewQGroupBox2("Peering information", nil)

	mainVBox.AddWidget(localInfoGroup, 0, core.Qt__AlignLeft)

	localInfoGrid := widgets.NewQGridLayout2()
	localInfoGroup.SetLayout(localInfoGrid)

	localInfoGrid.AddWidget(widgets.NewQLabel2("Cjdns IP:", nil, 0), 0, 0, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(widgets.NewQLabel2("Public Key:", nil, 0), 1, 0, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(widgets.NewQLabel2("Port:", nil, 0), 2, 0, core.Qt__AlignLeft)

	cjdnsIpLabel := widgets.NewQLabel2("Unknown", nil, 0)
	publicKeyLabel := widgets.NewQLabel2("Unknown", nil, 0)
	portLabel := widgets.NewQLabel2("Unknown", nil, 0)

	localInfoGrid.AddWidget(cjdnsIpLabel, 0, 1, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(publicKeyLabel, 1, 1, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(portLabel, 2, 1, core.Qt__AlignLeft)

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
	authPasswordGroup := widgets.NewQGroupBox2("Authorized passwords", nil)

	mainVBox.AddWidget(adminLoginGroup, 0, core.Qt__AlignLeft)
	mainVBox.AddWidget(authPasswordGroup, 0, core.Qt__AlignLeft)

	adminLoginVBox := widgets.NewQVBoxLayout()
	adminLoginGroup.SetLayout(adminLoginVBox)

	adminAddressInput := widgets.NewQInputDialog(nil, 0)
	adminAddressInput.SetLabelText("Address:")
	adminPasswordInput := widgets.NewQInputDialog(nil, 0)
	adminPasswordInput.SetLabelText("Password:")

	adminLoginVBox.AddWidget(adminAddressInput, 1, core.Qt__AlignLeft)
	adminLoginVBox.AddWidget(adminPasswordInput, 1, core.Qt__AlignLeft)

	authPasswordVBox := widgets.NewQVBoxLayout()
	authPasswordGroup.SetLayout(authPasswordVBox)

	authPasswordTextEdit := widgets.NewQPlainTextEdit(nil)
	authPasswordTextEdit.SetPlaceholderText("Authorized passwords")
	authPasswordTextEdit.SetLineWrapMode(widgets.QPlainTextEdit__NoWrap)
	authPasswordVBox.AddWidget(authPasswordTextEdit, 1, core.Qt__AlignLeft)

	saveButton := widgets.NewQPushButton2("Save", nil)
	mainVBox.AddWidget(saveButton, 1, core.Qt__AlignLeft)

	saveButton.ConnectClick(func() {
		widget.Save()
	})

	widget.ConnectSet(func(settings Settings) error {
		adminAddressInput.SetTextValue(settings.AdminAddress)
		adminPasswordInput.SetTextValue(settings.AdminPassword)

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
			AdminAddress:        adminAddressInput.TextValue(),
			AdminPassword:       adminPasswordInput.TextValue(),
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
