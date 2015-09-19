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
	"testing"
)

func TestBinarySearch(t *testing.T) {
	set1 := []string{"a", "b", "c", "d", "e", "f"}
	knowns1 := []int{2, 0, 5, -1} // known good values

	results1 := []int{
		binarySearch("c", set1),
		binarySearch("a", set1),
		binarySearch("f", set1),
		binarySearch("k", set1),
	}

	if !testEq(results1, knowns1) {
		t.Error("Bad binary search")
	}

	set2 := []string{"b", "d", "e", "f", "g"}
	knowns2 := []int{0, -1, -1}
	results2 := []int{
		binarySearch("b", set2),
		binarySearch("c", set2),
		binarySearch("a", set2),
	}

	if !testEq(results2, knowns2) {
		t.Error("Bad binary search")
	}

	set3 := []string{"b"}
	knowns3 := []int{-1, 0, -1}
	results3 := []int{
		binarySearch("a", set3),
		binarySearch("b", set3),
		binarySearch("c", set3),
	}

	if !testEq(results3, knowns3) {
		t.Error("Bad binary search")
	}

}

func testEq(a, b []int) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Tests only positive numerics. Negative numerics are undefined behavior.
func TestNumericPad(t *testing.T) {
	if padNumeric(1) != "001" {
		t.Error("Bad numeric padding")
	}

	if padNumeric(42) != "042" {
		t.Error("Bad numeric padding")
	}

	if padNumeric(256) != "256" {
		t.Error("Bad numeric padding")
	}
}
