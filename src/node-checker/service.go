package node_checker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"math"

	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Service provides access to User related data
type Service struct {
	db store
}

// DefaultService prepares new instance of Service
func DefaultService() Service {
	return NewService(DefaultData())
}

// NewService prepares new instance of Service
func NewService(nodeStore store) Service {
	return Service{
		db: nodeStore,
	}
}

func (ns *Service) exportAllNodesUptimes(startDate time.Time, endDate time.Time) ([]NodeUptimeResponse, error) {
	allNodes, err := ns.db.findNodes()
	if err != nil {
		return nil, errCannotFindNodes
	}
	var allNodeKeys []string
	for _, node := range allNodes {
		allNodeKeys = append(allNodeKeys, node.Key)
	}
	response, err := ns.getNodeInfoExport(allNodeKeys, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (ns *Service) getNodeInfoExport(nodeKeys []string, startDate time.Time, endDate time.Time) ([]NodeUptimeResponse, error) {
	var exportStart, exportEnd time.Time
	var currentYear int
	var thisMonth time.Month
	currentLocation := exportEnd.Location()

	if startDate.IsZero() || endDate.IsZero() {
		currentTime := time.Now()
		currentYear, thisMonth, _ = currentTime.Date()
		previousMonth := int(thisMonth)
		if previousMonth == 1 {
			previousMonth = 12
		} else {
			previousMonth--
		}
		previousMonthConverted := time.Month(previousMonth)
		exportStart = time.Date(currentYear, previousMonthConverted, 1, 0, 0, 0, 0, currentLocation)
		exportEnd = time.Date(currentYear, thisMonth, 1, 0, 0, 0, 0, currentLocation)
	} else {
		exportEnd = endDate
		exportStart = startDate
	}

	var results []NodeUptimeResponse
	for _, nodeString := range nodeKeys {
		nodeString = strings.TrimSpace(nodeString)
		uptimeSum := 0
		firstPeriodPastMonth := true
		lastPeriodActualMonth := Uptime{}
		dbNode, err := ns.db.findNode(nodeString)
		if err != nil {
			if err == errCannotLoadDataFromDatabase {
				log.Warn("Missing records for node ", nodeString)
				continue
			}
			log.Error("Unable to read data from the db due to error ", err)
			return nil, errCannotLoadData
		}
		for i := len(dbNode.Uptimes) - 1; i >= 0; i-- {
			uptime := dbNode.Uptimes[i]
			if exportEnd.Before(uptime.CreatedAt) {
				continue
			} else if uptime.UpdatedAt.After(exportEnd) {
				uptime.StartTime = int(exportEnd.Sub(uptime.CreatedAt).Seconds())
			}
			if exportStart.Before(uptime.CreatedAt) {
				uptimeSum += uptime.StartTime
				lastPeriodActualMonth = uptime
			} else if firstPeriodPastMonth {
				var actualMonthTime = 0
				if lastPeriodActualMonth.NodeId != "" {
					pastMonthPartOfStartTIme := int(exportStart.Sub(uptime.CreatedAt).Seconds())
					actualMonthTime = uptime.StartTime - pastMonthPartOfStartTIme
				} else {
					activeDate := uptime.CreatedAt.Add(time.Second * time.Duration(uptime.StartTime))
					if activeDate.After(exportStart) {
						actualMonthTime = int(activeDate.Sub(exportStart).Seconds())
					}
				}
				if actualMonthTime < 0 {
					actualMonthTime = 0
				}
				firstPeriodPastMonth = false
				uptimeSum = uptimeSum + actualMonthTime
			} else {
				break
			}
		}
		floatUptime := float64(uptimeSum)
		var duration time.Duration
		duration = exportEnd.Sub(exportStart)
		durationInSeconds := float64(duration) / float64(time.Second)
		if floatUptime > durationInSeconds {
			floatUptime = durationInSeconds
		}
		result := NodeUptimeResponse{
			Key:        nodeString,
			Uptime:     floatUptime,
			Downtime:   toFixed(durationInSeconds-floatUptime, 0),
			Percentage: floatUptime / durationInSeconds * 100,
			Online:     dbNode.Online,
		}
		results = append(results, result)

	}
	return results, nil
}

func (ns *Service) getNodeInfo(nodeKeys []string) ([]NodeUptimeResponse, error) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	var results []NodeUptimeResponse
	for _, nodeString := range nodeKeys {
		nodeString = strings.TrimSpace(nodeString)
		uptimeSum := 0
		firstPeriodPastMonth := true
		lastPeriodActualMonth := Uptime{}
		dbNode, err := ns.db.findNode(nodeString)
		if err != nil {
			if err == errCannotLoadDataFromDatabase {
				log.Warn("Missing records for node ", nodeString)
				continue
			}
			log.Error("Unable to read data from the db due to error ", err)
			return nil, errCannotLoadData
		}
		for i := len(dbNode.Uptimes) - 1; i >= 0; i-- {
			uptime := dbNode.Uptimes[i]
			if firstOfMonth.Before(uptime.CreatedAt) {
				uptimeSum += uptime.StartTime
				lastPeriodActualMonth = uptime
			} else if firstPeriodPastMonth {
				var actualMonthTime = 0
				if lastPeriodActualMonth.NodeId != "" {
					var downtimeSum = 0
					secondsBetweenPeriods := int(lastPeriodActualMonth.CreatedAt.Sub(uptime.CreatedAt).Seconds())
					if secondsBetweenPeriods > uptime.StartTime {
						downtimeSum = secondsBetweenPeriods - uptime.StartTime
					}
					actualMonthTime = int(lastPeriodActualMonth.CreatedAt.Sub(firstOfMonth).Seconds()) - downtimeSum

				} else {
					activeDate := uptime.CreatedAt.Add(time.Second * time.Duration(uptime.StartTime))
					if activeDate.After(firstOfMonth) {
						actualMonthTime = int(activeDate.Sub(firstOfMonth).Seconds())
					}
				}
				if actualMonthTime < 0 {
					actualMonthTime = 0
				}
				firstPeriodPastMonth = false
				uptimeSum = uptimeSum + actualMonthTime
			}
		}

		floatUptime := float64(uptimeSum)
		duration := time.Since(firstOfMonth)
		durationInSeconds := float64(duration) / float64(time.Second)
		if floatUptime > durationInSeconds {
			floatUptime = durationInSeconds
		}
		result := NodeUptimeResponse{
			Key:        nodeString,
			Uptime:     floatUptime,
			Downtime:   toFixed(durationInSeconds-floatUptime, 0),
			Percentage: floatUptime / durationInSeconds * 100,
			Online:     dbNode.Online,
		}
		results = append(results, result)
	}
	return results, nil
}

func (ns *Service) updateNodeInfo() error {
	start := time.Now()
	log.Info("Starting update process for nodes uptime")
	uptimeThreshold := int(viper.GetDuration("server.uptime-threshold").Seconds())
	res, err := getDataFromAPI()
	if err != nil {
		log.Error("Unable to fetch the data from the external API")
		return err
	}

	elapsed := time.Since(start)
	log.Infof("Pulled data from %v", elapsed)
	currentTime := time.Now()
	var totalTime time.Duration
	nodes, err := ns.db.findNodes()
	if err != nil && err != errCannotLoadDataFromDatabase {
		return err
	}
	for _, resUptime := range *res {
		if resUptime.StartTime > uptimeThreshold { // skipping records smaller than configured threshold
			dbNode, err := findNode(resUptime.Key, nodes)
			if err != nil {
				ns.createNewNode(resUptime, currentTime)
			} else {
				totalTime += ns.updateNode(dbNode, resUptime, currentTime)
			}
		}
	}

	log.Infof("Total time reading from db %v", totalTime)
	ns.db.updateAllNodesOnlineStatus(currentTime)

	log.Info("Done with updating")
	return nil
}

func (ns *Service) updateNode(node Node, def NodeDef, currentTime time.Time) time.Duration {
	err := ns.db.updateNodeOnlineStatus(&node, true, currentTime)
	if err != nil {
		log.Debug("Error updating online status for node with key: ", node.Key)
	}

	start := time.Now()
	lastUptime, err := ns.db.getLastUptimeForNode(node.Key)

	elapsed := time.Since(start)

	createNewRecord := false
	if err != nil {
		if err != errCannotLoadDataFromDatabase {
			log.Errorf("Unable to find uptime for node %v due to error %v", node.Key, err)
		}

		// We just don't have current node in db yet
		createNewRecord = true
	}

	//creating new record if we:
	createNewRecord = createNewRecord || // have not found one in DB, or
		def.StartTime <= lastUptime.StartTime || // new running time is smaller than previously recorded, or
		lastUptime.CreatedAt.Add(time.Duration(def.StartTime+UptimeDifferenceOffsetInSeconds)*time.Second).Before(currentTime) // dealing with old record we should leave as is in past

	if createNewRecord {
		// if running time is less or equal than before means that node was restarted
		newUptime := Uptime{
			NodeId:    node.Key,
			StartTime: def.StartTime,
			CreatedAt: currentTime.Add(time.Duration(-def.StartTime) * time.Second),
		}
		err := ns.db.createUptime(&newUptime)
		if err != nil {
			log.Info("Error creating new uptime for node with key: ", node.Key)
		}
	} else {
		// if running time increased means that same uptime should be kept
		lastUptime.StartTime = def.StartTime
		ns.db.updateUptime(&lastUptime)
	}
	return elapsed
}

func (ns *Service) createNewNode(def NodeDef, currentTime time.Time) {
	uptime := Uptime{
		NodeId:    def.Key,
		StartTime: def.StartTime,
		CreatedAt: currentTime.Add(time.Duration(-def.StartTime) * time.Second),
	}
	node := Node{
		Key:       def.Key,
		LastCheck: currentTime,
		Uptimes:   []Uptime{uptime},
		Online:    true,
	}

	if err := ns.db.createNode(&node); err != nil {
		log.Errorf("Unable to create a new node %v from received data %v", node, def)
		return
	}
}

func (ns *Service) createUptimesForPastMonths(nodeKey string, month int, year int, startTime int, percentage float64, downtime int) error {
	monthlyUptime := MonthlyUptime{
		NodeId:         nodeKey,
		Month:          month,
		Year:           year,
		TotalStartTime: startTime,
		Percentage:     percentage,
		Downtime:       downtime,
	}
	lastUptime, err := ns.db.getLastUptimeForNode(nodeKey)
	if err != nil {
		log.Error("Cannot get last uptime for node with key %v", nodeKey)
	}
	monthlyUptime.LastStartTime = lastUptime.StartTime
	erro := ns.db.createMonthlyUptime(&monthlyUptime)
	if erro != nil {
		log.Error("Cannot create uptimes for past months", err)
	} else {
		log.Info("Monthly uptime with key %v is created for %v. month", nodeKey, month)
	}
	return nil
}

func (ns *Service) getNodes() ([]Node, error) {
	nodes, err := ns.db.findNodes()
	if err != nil {
		log.Fatal("Cannot fetch nodes from db", err)
	}
	return nodes, nil
}

func (ns *Service) getNode(nodeKey string) (Node, error) {
	node, err := ns.db.findNode(nodeKey)
	if err != nil {
		log.Error("Cannot find node with %v key ", nodeKey)
	}
	return node, nil
}

func getDataFromAPI() (*NodeResponse, error) {
	var apiString = viper.GetString("server.node-check-api")
	response, err := http.Get(apiString)
	if err != nil {
		return nil, errCannotLoadData
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errCannotLoadData
	}

	uptimes, err := extractUptimesFromURL(contents)
	if err != nil {
		return nil, err
	}
	return uptimes, nil
}

func extractUptimesFromURL(body []byte) (*NodeResponse, error) {
	var s = new(NodeResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		return nil, errCannotLoadData
	}
	return s, err
}

func findNode(key string, allNodes []Node) (Node, error) {
	index := 0
	if len(allNodes) == 0 {
		return Node{}, errCannotFindNodeWithKey
	}
	if allNodes[index].Key > key {
		return Node{}, errCannotFindNodeWithKey
	}
	low, mid, high := 0, 0, len(allNodes)
	if allNodes[low].Key == key {
		return allNodes[low], nil
	}
	for (low + 1) != high {
		mid = low + (high-low)/2
		if allNodes[mid].Key == key {
			return allNodes[mid], nil
		} else if allNodes[mid].Key < key {
			low = mid
		} else {
			high = mid
		}
	}
	return Node{}, errCannotFindNodeWithKey
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// UptimeDifferenceOffsetInSeconds used to distingush beteen old uptime records that should be left unchanged.
// Calcuated based on time needed to complete one round of record storing
var UptimeDifferenceOffsetInSeconds = 200

type NodeResponse []NodeDef

type NodeDef struct {
	Key       string `json:"key"`
	StartTime int    `json:"start_time"`
}

type NodeUptimeResponse struct {
	Key        string
	Uptime     float64
	Downtime   float64
	Percentage float64
	Online     bool
}

type UptimeAndMonthlyUptimeDifference struct {
	NodeId     string
	Difference int
}
