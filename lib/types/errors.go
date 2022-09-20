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

package types

import (
	"errors"
	"fmt"
)

type RequestError struct {
	StatusCode int
	Err        error
}

var (
	// ErrPluginNotInstalled is returned if we try to get data for a plugin that is not installed.
	ErrPluginNotInstalled = errors.New("the plugin is not installed")
	// ErrPluginAlreadyInstalled is returned if we try to install an already installed plugin.
	ErrPluginAlreadyInstalled = errors.New("the plugin is already installed")
)

// Error returns the string version of the error.
func (e *RequestError) Error() string {
	return fmt.Sprintf("Status %d: Error: %q", e.StatusCode, e.Err)
}

func (e *RequestError) Unwrap() error { return e.Err }

// IsRequestError checks if the given error is a request error.
func IsRequestError(err error) bool {
	var e *RequestError
	return errors.As(err, &e)
}

// IsNotFoundError checks if the given error is a 404 Not Found.
func IsNotFoundError(err error) bool {
	var e *RequestError
	return errors.As(err, &e) && e.StatusCode == 404
}
