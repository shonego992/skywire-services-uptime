package node_checker

import (
	"github.com/SkycoinPro/skywire-services-uptime/src/database/postgres"

	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// store is node-checker related interface for dealing with database operations
type store interface {
	findNodes() ([]Node, error)
	findNode(key string) (Node, error)
	createNode(node *Node) error
	updateUptime(uptime *Uptime) error
	createUptime(uptime *Uptime) error
	updateNodeOnlineStatus(node *Node, status bool, currentTime time.Time) error
	updateAllNodesOnlineStatus(currentTime time.Time) error
	getLastUptimeForNode(nodeKey string) (Uptime, error)
	createMonthlyUptime(monthlyUptime *MonthlyUptime) error
}

// data implements store interface which uses GORM library
type data struct {
	db *gorm.DB
}

func DefaultData() data {
	return NewData(postgres.DB)
}

func NewData(database *gorm.DB) data {
	return data{
		db: database,
	}
}

func (u data) getLastUptimeForNode(nodeKey string) (Uptime, error) {
	var (
		uptime  Uptime
		dbError error
	)
	record := u.db.Where("node_id = ?", nodeKey).Last(&uptime)
	if record.RecordNotFound() {
		return Uptime{}, errCannotLoadDataFromDatabase
	}
	if errs := record.GetErrors(); len(errs) > 0 {
		for _, err := range errs {
			dbError = err
			log.Errorf("Error occurred while fetching uptime by key %v - %v", nodeKey, err)
		}
		return Uptime{}, dbError
	}

	return uptime, nil
}

func (u data) createUptime(uptime *Uptime) error {
	db := u.db.Begin()
	var dbError error
	for _, err := range db.Create(uptime).GetErrors() {
		dbError = err
		log.Error("Error while creating new uptime in DB %v", err)
	}
	if dbError != nil {
		db.Rollback()
		return dbError
	}
	db.Commit()

	return nil
}

func (u data) updateUptime(uptime *Uptime) error {
	db := u.db
	var dbError error
	for _, err := range db.Model(&uptime).Update("StartTime", uptime.StartTime).GetErrors() {
		dbError = err
		log.Error("Error while updating uptime in DB ", err)
	}
	if dbError != nil {
		return dbError
	}
	return nil
}

func (u data) updateNodeOnlineStatus(node *Node, status bool, currentTime time.Time) error {
	db := u.db.Begin()
	var dbError error
	for _, err := range db.Model(&node).UpdateColumns(Node{Online: status, LastCheck: currentTime, UpdatedAt: time.Now()}).GetErrors() {
		dbError = err
		log.Error("Error while updating node in DB ", err)
	}
	if dbError != nil {
		db.Rollback()
		return dbError
	}
	db.Commit()

	return nil
}

func (u data) createNode(node *Node) error {
	db := u.db.Begin()
	var dbError error
	for _, err := range db.Create(node).GetErrors() {
		dbError = err
		log.Error("Error while creating new node in DB ", err)
	}
	if dbError != nil {
		db.Rollback()
		return dbError
	}
	db.Commit()

	return nil
}

func (u data) findNodes() ([]Node, error) {
	var (
		nodes   []Node
		dbError error
	)

	record := u.db.Order("key ASC").Find(&nodes)
	if record.RecordNotFound() {
		return nil, errCannotLoadDataFromDatabase
	}
	if errs := record.GetErrors(); len(errs) > 0 {
		for _, err := range errs {
			dbError = err
			log.Error("Error occurred while fetching nodes - ", err)
		}
		return nil, dbError
	}

	return nodes, nil
}

func (u data) findNode(key string) (Node, error) {
	var (
		node    Node
		dbError error
	)
	record := u.db.Where("key = ?", key).Preload("MonthlyUptimes", func(db *gorm.DB) *gorm.DB { return db.Order("Monthly_Uptimes.id ASC") }).Preload("Uptimes", func(db *gorm.DB) *gorm.DB { return db.Order("Uptimes.id ASC") }).Find(&node)
	if record.RecordNotFound() {
		return Node{}, errCannotLoadDataFromDatabase
	}
	if errs := record.GetErrors(); len(errs) > 0 {
		for _, err := range errs {
			dbError = err
			log.Errorf("Error occurred while fetching node by key %v - %v", key, err)
		}
		return Node{}, dbError
	}

	return node, nil
}

func (u data) updateAllNodesOnlineStatus(currentTime time.Time) error {
	db := u.db.Begin()
	var dbError error
	for _, err := range db.Exec("UPDATE nodes set online=? where online=? and last_check < ?;", false, true, currentTime).GetErrors() {
		dbError = err
		log.Error("Error while updating nodes online status: ", err)
	}
	if dbError != nil {
		db.Rollback()
		return dbError
	}
	db.Commit()
	return nil
}

func (u data) createMonthlyUptime(monthlyUptime *MonthlyUptime) error {
	db := u.db.Begin()
	var dbError error
	for _, err := range db.Create(monthlyUptime).GetErrors() {
		dbError = err
		log.Error("Error while creating new monthly uptime in DB ", err)
	}
	if dbError != nil {
		db.Rollback()
		return dbError
	}
	db.Commit()

	return nil
}
