// The MIT License (MIT)
//
// Copyright (c) 2020 cupnoodles
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

func unbindAll(x *xgbutil.XUtil) error {
	keybind.Detach(x, x.RootWin())
	return nil
}

func unbind(xu *xgbutil.XUtil, btn string, onRelease bool) error {
	mod, codes, err := keybind.ParseString(xu, btn)
	if err != nil {
		err = fmt.Errorf(": %w", err)
		log.Printf("%v", err)

		return err
	}

	var evtype int

	if !onRelease {
		evtype = xevent.KeyPress
	} else {
		evtype = xevent.KeyRelease
	}

	detach(xu, evtype, xu.RootWin(), mod, codes[0])

	return nil
}

// detach ubinds a single keybinding. It's based on detach() from xgbutil (which
// unbinds all keys from the window) and accesses things that shouldn't be
// accessed.
func detach(xu *xgbutil.XUtil, evtype int, win xproto.Window, mods uint16,
	keycode xproto.Keycode) {
	xu.KeybindsLck.RLock()
	defer xu.KeybindsLck.RUnlock()

	for key := range xu.Keybinds {
		if win != key.Win || keycode != key.Code ||
			evtype != key.Evtype || mods != key.Mod {
			continue
		}

		ungrab(xu, key)
	}
}

func ungrab(xu *xgbutil.XUtil, key xgbutil.KeyKey) {
	xu.Keygrabs[key] -= len(xu.Keybinds[key])

	if xu.Keygrabs[key] == 0 {
		delete(xu.Keybinds, key)

		// check if the other event type is used and ungrab key if it isn't
		var otherEvType int

		if key.Evtype == xevent.KeyPress {
			otherEvType = xevent.KeyRelease
		} else {
			otherEvType = xevent.KeyPress
		}

		otherKey := key
		otherKey.Evtype = otherEvType

		if xu.Keygrabs[otherKey] == 0 {
			keybind.Ungrab(xu, key.Win, key.Mod, key.Code)
		}
	}
}
