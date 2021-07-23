package memorystore

import (
	"database/sql"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
)

//LogRepository ..
type LogRepository struct {
	db    sql.DB
	store *Store
	logs  []*model.LogLine
}

// Create ...
func (r *LogRepository) AddLine(sd *model.LogLine) {
	r.logs = append(r.logs, sd)
	// TODO Доделать
}

//FindByMac ..
func (r *LogRepository) GetAll() []*model.LogLine {
	return r.logs
}

func (r *LogRepository) Flush() error {

	//TODO доделать
	return nil
}
