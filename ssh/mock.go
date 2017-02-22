//
// Copyright (c) 2017 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package to mock ssh interface for unit tests
// Create a mock object and override the functions to test
// functions which return expected values
package ssh

import (
	"io"
	"os"
)

type MockSshExec struct {
	MockCopy     func(size int64, mode os.FileMode, fileName string, contents io.Reader, host, destinationPath string) error
	MockCopyPath func(sourcePath, host, destinationPath string) error
	MockExec     func(host string, commands []string, timeoutMinutes int, useSudo bool) ([]string, error)
}

func NewMockSshExecutor() *MockSshExec {
	m := &MockSshExec{}
	m.MockCopy = func(size int64,
		mode os.FileMode,
		fileName string,
		contents io.Reader,
		host, destinationPath string) error {
		return nil
	}

	m.MockCopyPath = func(sourcePath, host, destinationPath string) error {
		return nil
	}

	m.MockExec = func(host string,
		commands []string,
		timeoutMinutes int,
		useSudo bool) ([]string, error) {
		return []string{""}, nil
	}

	return m
}

func (m *MockSshExec) Copy(size int64,
	mode os.FileMode,
	fileName string,
	contents io.Reader,
	host, destinationPath string) error {
	return m.MockCopy(size, mode, fileName, contents, host, destinationPath)
}

func (m *MockSshExec) CopyPath(sourcePath, host, destinationPath string) error {
	return m.MockCopyPath(sourcePath, host, destinationPath)
}

func (m *MockSshExec) Exec(host string,
	commands []string,
	timeoutMinutes int,
	useSudo bool) ([]string, error) {
	return m.MockExec(host, commands, timeoutMinutes, useSudo)
}
