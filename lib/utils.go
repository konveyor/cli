/*
 *  Copyright IBM Corporation 2022
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package lib

func apply[T1 interface{}, T2 interface{}](f func(T1) T2, t1s []T1) []T2 {
	t2s := []T2{}
	for _, t1 := range t1s {
		t2s = append(t2s, f(t1))
	}
	return t2s
}

func filter[T interface{}](f func(T) bool, ts []T) []T {
	t1s := []T{}
	for _, t := range ts {
		if f(t) {
			t1s = append(t1s, t)
		}
	}
	return t1s
}

func findIndex[T interface{}](f func(T) bool, ts []T) int {
	for i, t := range ts {
		if f(t) {
			return i
		}
	}
	return -1
}

func contains[T comparable](t T, ts []T) bool {
	return findIndex(func(t1 T) bool { return t1 == t }, ts) != -1
}
