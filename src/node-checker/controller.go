package node_checker

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"strconv"

	"github.com/SkycoinPro/skywire-services-uptime/src/api"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const Nodes = "nodes"
const StartDate = "startDate"
const EndDate = "endDate"

// Controller is handling requests regarding Model
type Controller struct {
	nodeService Service
}

func DefaultController() Controller {
	return NewController(DefaultService())
}

func NewController(ns Service) Controller {
	return Controller{
		nodeService: ns,
	}
}

func (ctrl Controller) RegisterAPIs(public *gin.RouterGroup, closed *gin.RouterGroup) {
	publicUserGroup := public.Group("/info")

	publicUserGroup.GET("/updateNodeInfo", ctrl.updateNodeInfo)
	publicUserGroup.GET("/getNodeInfo", ctrl.getNodeInfo)
	publicUserGroup.GET("/getNodeInfoExport", ctrl.getPreviousMonthInfo)
	publicUserGroup.GET("/getAllUptimes", ctrl.getAllUptimes)
}

func (ctrl Controller) getAllUptimes(c *gin.Context) {
	var startDate, endDate time.Time
	params := c.Request.URL.Query()
	if len(params[StartDate]) <= 0 || len(params[EndDate]) <= 0 {
		startDate = time.Time{}
		endDate = time.Time{}
	} else {
		start, err1 := strconv.ParseInt(params[StartDate][0], 10, 64)
		end, err2 := strconv.ParseInt(params[EndDate][0], 10, 64)
		if err1 == nil && err2 == nil {
			startDate = time.Unix(start, 0)
			endDate = time.Unix(end, 0)
		}
	}
	response, err := ctrl.nodeService.exportAllNodesUptimes(startDate, endDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(200, response)

}

// @Summary Returns uptime info for previous month
// @Description Returns uptime info for nodes from the request
// @Tags nodes
// @Accept json
// @Produce json
// @Param nodes query string true "Node for checking of uptime status"
// @Success 200 {array} node_checker.NodeUptimeResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /info/getNodeInfoExport [get]
func (ctrl Controller) getPreviousMonthInfo(c *gin.Context) {

	params := c.Request.URL.Query()
	if len(params[Nodes]) <= 0 {
		log.Info("Node info requested for 0 nodes")
		c.AbortWithStatusJSON(http.StatusBadRequest, api.ErrorResponse{Error: "uptime service: zero nodes in request"})
		return
	}

	var startDate, endDate time.Time
	if len(params[StartDate]) <= 0 || len(params[EndDate]) <= 0 {
		startDate = time.Time{}
		endDate = time.Time{}
	} else {
		start, err1 := strconv.ParseInt(params[StartDate][0], 10, 64)
		end, err2 := strconv.ParseInt(params[EndDate][0], 10, 64)
		if err1 == nil && err2 == nil {
			startDate = time.Unix(start, 0)
			endDate = time.Unix(end, 0)
		}
	}

	nodesString := params[Nodes][0]
	//TODO consider returning some warning that node all keys were matched
	detail, err := ctrl.nodeService.getNodeInfoExport(strings.Split(nodesString, ","), startDate, endDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, detail)
}

// @Summary Returns uptime info
// @Description Returns uptime info for nodes from the request
// @Tags nodes
// @Accept json
// @Produce json
// @Param nodes query string true "Node for checking of uptime status"
// @Success 200 {array} node_checker.NodeUptimeResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /info/getNodeInfo [get]
func (ctrl Controller) getNodeInfo(c *gin.Context) {
	params := c.Request.URL.Query()
	if len(params[Nodes]) <= 0 {
		log.Info("Node info requested for 0 nodes")
		c.AbortWithStatusJSON(http.StatusBadRequest, api.ErrorResponse{Error: "uptime service: zero nodes in request"})
		return
	}
	nodesString := params[Nodes][0]
	//TODO consider returning some warning that node all keys were matched
	detail, err := ctrl.nodeService.getNodeInfo(strings.Split(nodesString, ","))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(200, detail)
}

// @Summary Updates node info
// @Description Updates nodes uptime info with up to date data
// @Tags nodes
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} api.ErrorResponse
// @Router /info/updateNodeInfo [get]
func (ctrl Controller) updateNodeInfo(c *gin.Context) {
	err := ctrl.nodeService.updateNodeInfo()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}
}

func (ctrl Controller) testFuncForMonthlyUptimes() error {
	sumForAllNodesUptimes := []int{}
	var totalStartTime int
	var totalMonthlyStartTimeForNode int
	sumForAllNodesUptimesAfterFuncCall := []int{}
	nodesThatAreNotSame := []string{}
	allNodes, err := ctrl.nodeService.getNodes()
	if err != nil {
		return err
	}
	for _, node := range allNodes {
		nodeWithUptimes, erro := ctrl.nodeService.getNode(node.Key)
		if erro != nil {
			log.Error("Cannot find node with %v key ", node.Key)
		}
		totalStartTime = 0
		for _, uptime := range nodeWithUptimes.Uptimes {
			totalStartTime += uptime.StartTime
		}
		sumForAllNodesUptimes = append(sumForAllNodesUptimes, totalStartTime)
	}
	ctrl.GetUptimesForPreviousMonths()

	for _, node := range allNodes {
		nodeWithUptimes, erro := ctrl.nodeService.getNode(node.Key)
		if erro != nil {
			log.Error("Cannot find node with %v key ", node.Key)
		}
		totalStartTime = 0
		for _, uptime := range nodeWithUptimes.Uptimes {
			totalStartTime += uptime.StartTime
		}
		sumForAllNodesUptimesAfterFuncCall = append(sumForAllNodesUptimesAfterFuncCall, totalStartTime)
	}

	if len(sumForAllNodesUptimes) == len(sumForAllNodesUptimesAfterFuncCall) {
		for i, j := 0, 0; i < len(sumForAllNodesUptimes); i++ {
			if sumForAllNodesUptimes[i] != sumForAllNodesUptimesAfterFuncCall[j] {
				nodesThatAreNotSame = append(nodesThatAreNotSame, "wrong")
			}
			j++
		}
	} else {
		log.Error("Number of nodes is different after function call")
	}

	file, err := os.Create("startTimeDifference.csv")
	if err != nil {
		log.Debug("Cannot create csv file")
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	var headline = []string{"node_key", "difference"}
	err = writer.Write(headline)
	if err != nil {
		log.Debug("Cannot write csv file")
	}

	allDifferences := []UptimeAndMonthlyUptimeDifference{}
	sortedAllDifferences := []UptimeAndMonthlyUptimeDifference{}
	for _, node := range allNodes {
		nodeWithMonthlyUptimes, erro := ctrl.nodeService.getNode(node.Key)
		if erro != nil {
			log.Error("Cannot find node with %v key ", node.Key)
		}
		totalStartTime = 0
		totalMonthlyStartTimeForNode = 0
		for _, uptime := range nodeWithMonthlyUptimes.Uptimes {
			totalStartTime += uptime.StartTime
		}
		for _, monthlyUptime := range nodeWithMonthlyUptimes.MonthlyUptimes {
			totalMonthlyStartTimeForNode += monthlyUptime.TotalStartTime
		}
		if totalStartTime != totalMonthlyStartTimeForNode {
			log.Error("Sum of all uptimes for node and all monthly uptimes are not same")
			difference := totalStartTime - totalMonthlyStartTimeForNode
			uptimeAndMonthlyUptimeDifference := UptimeAndMonthlyUptimeDifference{
				NodeId:     node.Key,
				Difference: difference,
			}
			allDifferences = append(allDifferences, uptimeAndMonthlyUptimeDifference)
		}
	}
	for 0 < len(allDifferences) {
		min := 0
		for j := 1; j < len(allDifferences); j++ {
			if allDifferences[j].Difference > allDifferences[min].Difference {
				min = j
			}
		}
		sortedAllDifferences = append(sortedAllDifferences, allDifferences[min])
		allDifferences = append(allDifferences[:min], allDifferences[min+1:]...)
	}
	for _, sortedDifference := range sortedAllDifferences {
		differenceString := strconv.Itoa(sortedDifference.Difference)
		values := []string{sortedDifference.NodeId, differenceString}
		err = writer.Write(values)
		if err != nil {
			log.Debug("Cannot insert values into csv file")
		}
	}
	return nil
}

func (ctrl Controller) GetUptimesForPreviousMonths() {

	nodes, err := ctrl.nodeService.getNodes()
	if err != nil {
		log.Fatal("Cannot get any node")
	}

	var startDate, endDate time.Time
	currentLocation := endDate.Location()

	for _, node := range nodes {
		nodeKeys := []string{node.Key}
		for i := 2; i <= 12; i++ {
			convertStartMonth := time.Month(i)
			startDate = time.Date(2019, convertStartMonth, 1, 0, 0, 0, 0, currentLocation)

			convertEndMonth := time.Month(i + 1)
			endDate = time.Date(2019, convertEndMonth, 1, 0, 0, 0, 0, currentLocation)
			details, err := ctrl.nodeService.getNodeInfoExport(nodeKeys, startDate, endDate)
			if err != nil {
				log.Error("Cannot get node info")
			}
			for _, detail := range details {
				if detail.Uptime != 0 {
					uptimeInt := int(detail.Uptime)
					downtimeInt := int(detail.Downtime)
					err := ctrl.nodeService.createUptimesForPastMonths(detail.Key, i, 2019, uptimeInt, detail.Percentage, downtimeInt)
					if err != nil {
						log.Error("Cannot make new monthly uptime record", err)
					}
				}
			}
		}
	}
}

func (ctrl Controller) RunningRoutine() {
	diff := viper.GetDuration("server.refresh-interval")
	jobTicker := &jobTicker{}
	//ctrl.nodeService.updateNodeInfo()
	//ctrl.compareCSV()
	//ctrl.compareCsvWithMU()
	//ctrl.getUptimesForPreviousMonths()
	//ctrl.testFuncForMonthlyUptimes()
	jobTicker.updateTimer(diff)
	for {
		<-jobTicker.timer.C
		log.Info("Scheduler triggered, current time: ", time.Now())
		//ctrl.nodeService.updateNodeInfo()
		jobTicker.updateTimer(diff)
	}
}

type exportDate struct {
	StartDate int64
	EndDate   int64
}

func (ctrl Controller) compareCSV() error {

	var sliceOfConvertedCsvFiles1 []CSVDifferences
	var sliceOfConvertedCsvFiles2 []CSVDifferences
	var csvFileDifferences []CSVDifferences
	f, err := os.Open("C:/Users/Nesa/go-workspace/src/github.com/SkycoinPro/skywire-services-uptime/minerExport_1_12_2019-1_1_2020_all (1).csv")
	defer f.Close()
	if err != nil {
		return err
	}
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}
	header := true
	for _, line := range lines {
		if header {
			header = false
			continue
		}
		s, _ := strconv.ParseFloat(line[5], 64)
		csvDiff1 := CSVDifferences{
			NodeID: line[2],
			Uptime: s,
		}
		sliceOfConvertedCsvFiles1 = append(sliceOfConvertedCsvFiles1, csvDiff1)
		fmt.Println(csvDiff1)

	}

	file, err := os.Open("C:/Users/Nesa/go-workspace/src/github.com/SkycoinPro/skywire-services-uptime/minerExport_1_12_2019-1_1_2020_all.csv")
	defer file.Close()
	if err != nil {
		return err
	}
	lines2, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}
	header2 := true
	for _, line := range lines2 {
		if header2 {
			header2 = false
			continue
		}
		s, _ := strconv.ParseFloat(line[5], 64)
		csvDiff2 := CSVDifferences{
			NodeID: line[2],
			Uptime: s,
		}
		sliceOfConvertedCsvFiles2 = append(sliceOfConvertedCsvFiles2, csvDiff2)
		fmt.Println(csvDiff2)
	}

	newFile, err := os.Create("csvDifference.csv")
	if err != nil {
		log.Debug("Cannot create csv file")
	}
	defer newFile.Close()
	writer := csv.NewWriter(newFile)
	var headline = []string{"node_key", "difference"}
	err = writer.Write(headline)
	if err != nil {
		log.Debug("Cannot write csv file")
	}

	for i := 0; i <= len(sliceOfConvertedCsvFiles1)-1; i++ {
		diff := sliceOfConvertedCsvFiles1[i].Uptime - sliceOfConvertedCsvFiles2[i].Uptime
		if sliceOfConvertedCsvFiles1[i].NodeID != sliceOfConvertedCsvFiles2[i].NodeID {
			log.Info("node key from first file %v and node key from second file %v are not same", sliceOfConvertedCsvFiles1[i].NodeID, sliceOfConvertedCsvFiles2[i].NodeID)
		}
		if diff == 0 {
			continue
		}
		csvFileDiff := CSVDifferences{
			NodeID: sliceOfConvertedCsvFiles1[i].NodeID,
			Uptime: diff,
		}
		csvFileDifferences = append(csvFileDifferences, csvFileDiff)
	}

	for _, csvFileDifference := range csvFileDifferences {
		diffFloatToString := fmt.Sprintf("%f", csvFileDifference.Uptime)
		values := []string{csvFileDifference.NodeID, diffFloatToString}
		err = writer.Write(values)
		if err != nil {
			log.Debug("Cannot insert values into csv file")
		}
	}

	return nil
}

func (ctrl Controller) compareCsvWithMU() error {

	mus := []CSVDifferences{}
	sliceOfConvertedCsvFiles2 := []CSVDifferences{}

	allNodes, err := ctrl.nodeService.getNodes()
	if err != nil {
		return err
	}
	for _, node := range allNodes {
		nodeWithUptimes, erro := ctrl.nodeService.getNode(node.Key)
		if erro != nil {
			log.Error("Cannot find node with %v key ", node.Key)
		}
		for _, nodeMu := range nodeWithUptimes.MonthlyUptimes {
			if nodeMu.Month == 12 {
				compare := CSVDifferences{
					NodeID: nodeMu.NodeId,
					Uptime: float64(nodeMu.TotalStartTime),
				}
				mus = append(mus, compare)
			}
		}
	}
	file, err := os.Open("C:/Users/Nesa/go-workspace/src/github.com/SkycoinPro/skywire-services-uptime/minerExport_1_12_2019-1_1_2020_all (2).csv")
	defer file.Close()
	if err != nil {
		return err
	}
	lines2, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}
	header2 := true
	for _, line := range lines2 {
		if header2 {
			header2 = false
			continue
		}
		i, _ := strconv.ParseFloat(line[5], 64)
		if i > 0 {
			csvDiff2 := CSVDifferences{
				NodeID: line[2],
				Uptime: i,
			}
			sliceOfConvertedCsvFiles2 = append(sliceOfConvertedCsvFiles2, csvDiff2)
			fmt.Println(csvDiff2)
		}
	}
	for i := 0; i <= len(sliceOfConvertedCsvFiles2)-1; i++ {
		//diff := mus[i].Uptime - sliceOfConvertedCsvFiles2[i].Uptime
		if mus[i].NodeID != sliceOfConvertedCsvFiles2[i].NodeID {
			log.Info("node key from first file %v and node key from second file %v are not same", mus[i].NodeID, sliceOfConvertedCsvFiles2[i].NodeID)
		}
	}
	return nil
}

type CSVDifferences struct {
	NodeID string
	Uptime float64
}
