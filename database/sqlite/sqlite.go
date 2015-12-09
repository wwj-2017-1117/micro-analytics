package sqlite

import (
	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/database/errors"
	"github.com/GitbookIO/micro-analytics/database/structures"
)

type SQLite struct {
	DBManager *DBManager
	directory string
}

func NewSimpleDriver(driverOpts database.DriverOpts) *SQLite {
	manager := NewManager(ManagerOpts{driverOpts})
	return &SQLite{
		DBManager: manager,
		directory: driverOpts.Directory,
	}
}

func (driver *SQLite) Query(params structures.Params) (*structures.Analytics, error) {
	// Construct DBPath
	dbPath := DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	driver.DBManager.RequestDB <- dbPath
	db := <-driver.DBManager.SendDB

	// If value is in Cache, return directly
	cached, inCache := driver.DBManager.Cache.Get(params.URL)
	if inCache {
		if response, ok := cached.(*structures.Analytics); ok {
			driver.DBManager.UnlockDB <- NewUnlock(dbPath)
			return response, nil
		}
	}

	// Return query result
	analytics, err := db.Query(params.TimeRange)
	if err != nil {
		return nil, &errors.InternalError
	}

	// Unlock DB
	driver.DBManager.UnlockDB <- NewUnlock(dbPath)

	// Store response in Cache before sending
	driver.DBManager.Cache.Add(params.URL, analytics)

	return analytics, nil
}

func (driver *SQLite) GroupBy(params structures.Params) (*structures.Aggregates, error) {
	// Construct DBPath
	dbPath := DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	driver.DBManager.RequestDB <- dbPath
	db := <-driver.DBManager.SendDB

	// If value is in Cache, return directly
	cached, inCache := driver.DBManager.Cache.Get(params.URL)
	if inCache {
		if response, ok := cached.(*structures.Aggregates); ok {
			driver.DBManager.UnlockDB <- NewUnlock(dbPath)
			return response, nil
		}
	}

	// Check for unique query parameter to call function accordingly
	var analytics *structures.Aggregates

	if params.Unique {
		analytics, err = db.GroupByUniq(params.Property, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	} else {
		analytics, err = db.GroupBy(params.Property, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	}

	// Unlock DB
	driver.DBManager.UnlockDB <- NewUnlock(dbPath)

	// Store response in Cache before sending
	driver.DBManager.Cache.Add(params.URL, analytics)

	return analytics, nil
}

func (driver *SQLite) OverTime(params structures.Params) (*structures.Intervals, error) {
	// Construct DBPath
	dbPath := DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	driver.DBManager.RequestDB <- dbPath
	db := <-driver.DBManager.SendDB

	// If value is in Cache, return directly
	cached, inCache := driver.DBManager.Cache.Get(params.URL)
	if inCache {
		if response, ok := cached.(*structures.Intervals); ok {
			driver.DBManager.UnlockDB <- NewUnlock(dbPath)
			return response, nil
		}
	}

	// Check for unique query parameter to call function accordingly
	var analytics *structures.Intervals

	if params.Unique {
		analytics, err = db.OverTimeUniq(params.Interval, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	} else {
		analytics, err = db.OverTime(params.Interval, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	}

	// Unlock DB
	driver.DBManager.UnlockDB <- NewUnlock(dbPath)

	// Store response in Cache before sending
	driver.DBManager.Cache.Add(params.URL, analytics)

	return analytics, nil
}

func (driver *SQLite) Push(params structures.Params, analytic structures.Analytic) error {
	// Construct DBPath
	dbPath := DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Get DB from manager
	driver.DBManager.RequestDB <- dbPath
	db := <-driver.DBManager.SendDB

	// Insert data if everything's OK
	err := db.Insert(analytic)

	// Unlock DB
	driver.DBManager.UnlockDB <- NewUnlock(dbPath)

	if err != nil {
		return &errors.InsertFailed
	}

	return nil
}

func (driver *SQLite) Delete(params structures.Params) error {
	// Construct DBPath
	dbPath := DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return &errors.InvalidDatabaseName
	}

	// Delete full DB directory
	err = driver.DBManager.DeleteDB(dbPath)
	return err
}