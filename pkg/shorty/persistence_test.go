package shorty

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	databaseFile       = "/var/tmp/shorty-test.sqlite"
	short              = "abc"
	longURL            = "http://x.y.z"
	createdAt    int64 = 0
)

var persistence *Persistence

// DeleteDatabaseFile making sure database file is removed before starting
func DeleteDatabaseFile(t *testing.T) {
	t.Logf("Deleting database-file: '%s'", databaseFile)
	_ = os.Remove(databaseFile)
}

func TestPersistenceNew(t *testing.T) {
	var err error

	DeleteDatabaseFile(t)

	config := &Config{DatabaseFile: databaseFile}
	persistence, err = NewPersistence(config)

	assert.Nil(t, err)
}

func TestPersistenceWrite(t *testing.T) {
	shortened := &Shortened{Short: short, URL: longURL, CreatedAt: createdAt}

	err := persistence.Write(context.Background(), shortened)
	assert.Nil(t, err)

	// should return error on trying to re-insert
	err = persistence.Write(context.Background(), shortened)
	assert.Error(t, err)
}

func TestPersistenceRead(t *testing.T) {
	shortened, err := persistence.Read(context.Background(), short)

	assert.Nil(t, err)
	assert.Equal(t, short, shortened.Short)
	assert.Equal(t, longURL, shortened.URL)
	assert.Equal(t, createdAt, shortened.CreatedAt)
}
