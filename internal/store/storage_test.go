package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "protodex-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storage, err := New(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, storage)
	err = storage.Close()
	assert.NoError(t, err)
}

func TestCreatePackage(t *testing.T) {
	storage := setupTestStorage(t)
	defer cleanupTestStorage(t, storage)

	pkgStore := storage.Package()

	pkg, err := pkgStore.CreatePackage("test-package", "A test package", "user123", []string{"tag1", "tag2"})
	require.NoError(t, err)

	assert.NotEmpty(t, pkg.ID)
	assert.Equal(t, "test-package", pkg.Name)
	assert.Equal(t, "A test package", pkg.Description)
	assert.Equal(t, []string{"tag1", "tag2"}, pkg.Tags)
	assert.Equal(t, "user123", pkg.OwnerID)
}

func TestGetPackage(t *testing.T) {
	storage := setupTestStorage(t)
	defer cleanupTestStorage(t, storage)

	pkgStore := storage.Package()

	// Create a package
	createdPkg, err := pkgStore.CreatePackage("test-package", "A test package", "user123", []string{"tag1"})
	require.NoError(t, err)

	// Get the package
	pkg, err := pkgStore.GetPackage("test-package")
	require.NoError(t, err)

	assert.Equal(t, createdPkg.ID, pkg.ID)
	assert.Equal(t, "test-package", pkg.Name)
	assert.Equal(t, "A test package", pkg.Description)
	assert.Equal(t, []string{"tag1"}, pkg.Tags)
	assert.Equal(t, "user123", pkg.OwnerID)

	// Test non-existent package
	_, err = pkgStore.GetPackage("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListPackages(t *testing.T) {
	storage := setupTestStorage(t)
	defer cleanupTestStorage(t, storage)

	pkgStore := storage.Package()

	// Create public packages
	_, err := pkgStore.CreatePackage("public1", "Public package 1", "user1", []string{})
	require.NoError(t, err)
	_, err = pkgStore.CreatePackage("public2", "Public package 2", "user2", []string{})
	require.NoError(t, err)

	packages, err := pkgStore.ListPackages()
	require.NoError(t, err)
	assert.Equal(t, 2, len(packages))
}

// Helper functions for test setup
func setupTestStorage(t *testing.T) Store {
	tmpDir, err := os.MkdirTemp("", "protodex-test-")
	require.NoError(t, err)

	storage, err := New(tmpDir)
	require.NoError(t, err)
	err = storage.Init()
	require.NoError(t, err)

	t.Cleanup(func() {
		storage.Close()
		os.RemoveAll(tmpDir)
	})

	return storage
}

func cleanupTestStorage(_ *testing.T, storage Store) {
	storage.Close()
}
