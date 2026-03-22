package market

import (
	"database/sql"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type Repository interface {
	Close()
}

type MysqlRepository struct {
	db     *sql.DB
	logger util.Logger
}

func NewMysqlRepository(url string, logger util.Logger) (Repository, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		logger.Database().Fatal().Err(err).Msg("Database connection failed")
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		logger.Database().Fatal().Err(err).Msg("Database connection failed")
		return nil, err
	}
	logger.Database().Info().Msg("Database connection established")
	return &MysqlRepository{db: db, logger: logger}, nil
}

func (repository *MysqlRepository) Close() {
	repository.db.Close()
}

func 
