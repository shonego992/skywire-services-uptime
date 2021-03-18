package node_checker

import (
	"time"
)

type Node struct {
	Key            string          `gorm:"primary_key" json:"key"`
	Online         bool            `json:"online"`
	LastCheck      time.Time       `json:"lastCheck"`
	CreatedAt      time.Time       `json:"-"`
	UpdatedAt      time.Time       `json:"-"`
	DeletedAt      *time.Time      `json:"-"`
	Uptimes        []Uptime        `json:"-" gorm:"foreignkey:NodeId; PRELOAD:true"`
	MonthlyUptimes []MonthlyUptime `json:"-" gorm:"foreignkey:NodeId; PRELOAD:true"`
}

type Uptime struct {
	Id        uint       `gorm:"primary_key" json:"key"`
	NodeId    string     `json:"-"`
	StartTime int        `json:"startTime"` // represents running time since the last start
	CreatedAt time.Time  `json:"-"`         // holds the actual time when node was started, not when it's persisted in our database
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}

type MonthlyUptime struct {
	Id             uint       `gorm:"primary_key" json:"key"`
	NodeId         string     `json:"nodeId"`
	Month          int        `json:"month"`
	Year           int        `json:"year"`
	TotalStartTime int        `json:"totalStartTime"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"disabled,omitempty"`
	Percentage     float64    `json:"percentage"`
	Downtime       int        `json:"downtime"`
	LastStartTime  int        `json:"lastStartTime"`
}
