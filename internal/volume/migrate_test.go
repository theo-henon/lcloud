package volume

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPrepareProtocolsColumnExistingRows(t *testing.T) {
	dsn := "host=localhost user=lcloud password=changeme dbname=lcloud port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("postgres not available:", err)
	}

	// Simulate pre-protocols schema: drop column if present, insert row without protocols.
	_ = db.Exec(`ALTER TABLE volumes DROP COLUMN IF EXISTS protocols`)
	require.NoError(t, db.Exec(`
		INSERT INTO volumes (id, name, owner_id, disk_path, root_path, quota_bytes, used_bytes, filters, created_at, updated_at)
		VALUES (?, 'migrate-test', ?, '/disk', '/root', 0, 0, '{}', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, uuid.New(), uuid.New()).Error)

	require.NoError(t, PrepareProtocolsColumn(db))
	require.NoError(t, db.AutoMigrate(&Volume{}))

	var count int64
	require.NoError(t, db.Model(&Volume{}).Where("protocols IS NULL").Count(&count).Error)
	require.Zero(t, count)

	_ = db.Exec(`DELETE FROM volumes WHERE name = 'migrate-test'`)
}
