// Package tcx models the subset of the Garmin Training Center Database (TCX)
// format needed to recover lap boundaries. Laps do not survive Garmin Connect's
// GPX export, which flattens an activity to a single track segment, so the TCX
// export is the only readily available source for them.
package tcx

import "encoding/xml"

// TrainingCenterDatabase is the root of a TCX document. Only the elements
// needed to locate lap boundaries are modelled; everything else, including the
// namespaced extension elements, is ignored. encoding/xml matches on local
// name, so the default TCX namespace needs no handling here.
type TrainingCenterDatabase struct {
	XMLName    xml.Name   `xml:"TrainingCenterDatabase"`
	Activities []Activity `xml:"Activities>Activity"`
}

type Activity struct {
	Sport string `xml:"Sport,attr"`
	// ID is the activity's <Id>, an ISO-8601 timestamp of its start.
	ID   string `xml:"Id"`
	Laps []Lap  `xml:"Lap"`
}

type Lap struct {
	StartTime string `xml:"StartTime,attr"`
	// Tracks usually holds a single track, but a lap that was paused part way
	// through is split across one track per recorded stretch.
	Tracks []Track `xml:"Track"`
}

type Track struct {
	Trackpoints []Trackpoint `xml:"Trackpoint"`
}

type Trackpoint struct {
	Time string `xml:"Time"`
	// Position is absent for trackpoints recorded without a GPS fix.
	Position *Position `xml:"Position"`
	// Altitude is absent for devices that record no elevation.
	Altitude *float64 `xml:"AltitudeMeters"`
}

type Position struct {
	Latitude  float64 `xml:"LatitudeDegrees"`
	Longitude float64 `xml:"LongitudeDegrees"`
}
