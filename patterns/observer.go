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

package patterns

type ObserverTopic int
type ObserverFn func(data interface{}) error

type ObservableInterface interface {
	NotifyObservers(topic ObserverTopic)
	AddObserver(topic ObserverTopic, fn ObserverFn)
}

type Observer struct {
	Fn    ObserverFn
	Topic ObserverTopic
}

type Observable struct {
	ObservableInterface
	observers []Observer
}

func (o *Observable) AddObserver(topic ObserverTopic, fn ObserverFn) {
	o.observers = append(o.observers, Observer{Fn: fn, Topic: topic})
}

func (o *Observable) NotifyObservers(topic ObserverTopic, data interface{}) {
	var observer Observer

	for i := 0; i < len(o.observers); i++ {
		observer = o.observers[i]
		if observer.Topic == topic {
			if err := observer.Fn(data); err != nil {
				panic(err)
			}
		}
	}
}
