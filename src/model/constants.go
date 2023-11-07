package model

type Priority int

const (
	HIGH_PRIORITY Priority = iota
	LOW_PRIORITY
)

const PRIORITY_LEVEL_COUNT int = 2

type Bitrate int

const (
	LOW_BITRATE    Bitrate = 3
	MEDIUM_BITRATE Bitrate = 5
	HIGH_BITRATE   Bitrate = 10
)
