package node_checker

import "errors"

var errCannotLoadData = errors.New("node checker controller: cannot load data form url")
var errCannotFindNodes = errors.New("node checker controller: cannot find nodes")
var errCannotFindNodeWithKey = errors.New("node checker controller: cannot find node with key")
var errCannotLoadDataFromDatabase = errors.New("node checker controller: cannot load data from database")
var errUnableToProcessRequest = errors.New("node checker controller: cannot process request")