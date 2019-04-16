// Copyright (c) 2019 Sylabs, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sylabs/slurm-operator/pkg/slurm/local"

	"github.com/stretchr/testify/require"
)

func newTestServer(_ *testing.T) (*httptest.Server, func()) {
	slurmClient := &local.Client{}

	router := NewSlurmRouter(slurmClient)
	srv := httptest.NewServer(router)
	return srv, func() {
		srv.CloseClientConnections()
		srv.Close()
	}
}

func TestApi_Open(t *testing.T) {
	testFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())

	fileContent := []byte(`
Hello!
This is a test output from SLURM!
`)

	_, err = testFile.Write(fileContent)
	require.NoError(t, err)
	require.NoError(t, testFile.Close())

	srv, cleanup := newTestServer(t)
	defer cleanup()

	t.Run("no path", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/open", srv.URL), nil)
		require.NoError(t, err)
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "no path query parameter is found\n", string(content))
	})

	t.Run("non existent file", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/open?path=/foo/bar", srv.URL), nil)
		require.NoError(t, err)
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("all ok", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/open?path=%s", srv.URL, testFile.Name()), nil)
		require.NoError(t, err)
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, fileContent, content, "unexpected file content")
	})
}
