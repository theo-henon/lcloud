package volume

import (
	"fmt"

	"gorm.io/gorm"
)

const defaultProtocolsJSON = `{"webdav":{"enabled":false},"ftp":{"enabled":false}}`

// PrepareProtocolsColumn adds the protocols JSON column safely on PostgreSQL when
// volumes already exist. GORM AutoMigrate with NOT NULL fails on backfill; this
// runs first with nullable add + UPDATE + NOT NULL + DEFAULT.
// On a fresh install the volumes table does not exist yet — AutoMigrate creates it.
func PrepareProtocolsColumn(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	if !db.Migrator().HasTable(&Volume{}) {
		return nil
	}
	if db.Migrator().HasColumn(&Volume{}, "Protocols") {
		return backfillNullProtocols(db)
	}
	if err := db.Exec(`ALTER TABLE volumes ADD COLUMN protocols jsonb`).Error; err != nil {
		return fmt.Errorf("add protocols column: %w", err)
	}
	if err := backfillNullProtocols(db); err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE volumes ALTER COLUMN protocols SET NOT NULL`).Error; err != nil {
		return fmt.Errorf("set protocols not null: %w", err)
	}
	if err := db.Exec(`ALTER TABLE volumes ALTER COLUMN protocols SET DEFAULT '` + defaultProtocolsJSON + `'::jsonb`).Error; err != nil {
		return fmt.Errorf("set protocols default: %w", err)
	}
	return nil
}

func backfillNullProtocols(db *gorm.DB) error {
	return db.Exec(`UPDATE volumes SET protocols = '`+defaultProtocolsJSON+`'::jsonb WHERE protocols IS NULL`).Error
}
