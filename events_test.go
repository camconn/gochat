/*
gochat -- A light and speedy IRC server.
Copyright (C) 2015 Cameron Conn <cam_at_camconn_dot_cc>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"regexp"
	"testing"
)

func TestNickRegex(t *testing.T) {
	// NICKREGEX is found in events.go
	nickRE, _ := regexp.Compile(NICKREGEX)
	failed := false

	testSet := []string{
		"loooooooooongnick",
		"cameron",
		"2kool",
		"camconn",
		"lt",
		"okay...",
		"#channel",
		"&channelheretoo",
		"Bad chars",
		"()pls",
		"Wut()",
		"OK[]",
		"[]Wow",
	}

	knowns := []bool{
		false,
		true,
		true,
		true,
		true,
		true,
		false,
		false,
		false,
		false,
		true,
		true,
		false,
	}

	for i, nick := range testSet {
		matched := nickRE.MatchString(nick)
		if matched != knowns[i] {
			t.Logf("Matched %s when we should not have\n", nick)
			failed = true
		}

		if matched {
			t.Logf("Matched %s\n", nick)
		}
	}

	if failed {
		t.Fail()
	}
}
