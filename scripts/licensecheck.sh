#!/usr/bin/env bash

#   Copyright IBM Corporation 2022
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

find_files_without_string() {
  find . \( -name '.git' -o -name 'node_modules' \) -prune -o -type f \( -name '*.go' -o -name '*.sh' \) -exec grep -Le "$1" {} +
}

filepaths=$(find_files_without_string 'Licensed under the Apache License, Version 2.0 (the "License")')

if [[ $filepaths ]]; then
  echo "The following files are missing the license:"
  echo "$filepaths"
  exit 1
fi
