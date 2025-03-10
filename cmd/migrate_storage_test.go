// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"code.gitea.io/gitea/models/packages"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	packages_module "code.gitea.io/gitea/modules/packages"
	"code.gitea.io/gitea/modules/storage"
	packages_service "code.gitea.io/gitea/services/packages"

	"github.com/stretchr/testify/assert"
)

func TestMigratePackages(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	creator := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	content := "package main\n\nfunc main() {\nfmt.Println(\"hi\")\n}\n"
	buf, err := packages_module.CreateHashedBufferFromReader(strings.NewReader(content), 1024)
	assert.NoError(t, err)
	defer buf.Close()

	v, f, err := packages_service.CreatePackageAndAddFile(&packages_service.PackageCreationInfo{
		PackageInfo: packages_service.PackageInfo{
			Owner:       creator,
			PackageType: packages.TypeGeneric,
			Name:        "test",
			Version:     "1.0.0",
		},
		Creator:           creator,
		SemverCompatible:  true,
		VersionProperties: map[string]string{},
	}, &packages_service.PackageFileCreationInfo{
		PackageFileInfo: packages_service.PackageFileInfo{
			Filename: "a.go",
		},
		Data:   buf,
		IsLead: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.NotNil(t, f)

	ctx := context.Background()

	p, err := os.MkdirTemp(os.TempDir(), "migrated_packages")
	assert.NoError(t, err)

	dstStorage, err := storage.NewLocalStorage(
		ctx,
		storage.LocalStorageConfig{
			Path: p,
		})
	assert.NoError(t, err)

	err = migratePackages(ctx, dstStorage)
	assert.NoError(t, err)

	entries, err := os.ReadDir(p)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(entries))
	assert.EqualValues(t, "01", entries[0].Name())
	assert.EqualValues(t, "tmp", entries[1].Name())
}
