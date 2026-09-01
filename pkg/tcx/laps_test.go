package tcx

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLapBounds(t *testing.T) {
	t.Run("takes the first and last positioned trackpoint of each lap", func(t *testing.T) {
		db := TrainingCenterDatabase{Activities: []Activity{{Laps: []Lap{
			{Tracks: []Track{{Trackpoints: []Trackpoint{
				trackpoint("2026-08-30T13:13:30.000Z", 1, 2),
				trackpoint("2026-08-30T13:13:31.000Z", 3, 4),
				trackpoint("2026-08-30T13:13:32.000Z", 5, 6),
			}}}},
		}}}}

		bounds, skipped, err := db.LapBounds()
		require.NoError(t, err)
		assert.Empty(t, skipped)
		require.Len(t, bounds, 1)
		assert.Equal(t, 1, bounds[0].Number)
		assert.Equal(t, 1.0, bounds[0].First.Latitude)
		assert.Equal(t, 2.0, bounds[0].First.Longitude)
		assert.Equal(t, time.Date(2026, 8, 30, 13, 13, 30, 0, time.UTC), bounds[0].First.Timestamp)
		assert.Equal(t, 5.0, bounds[0].Last.Latitude)
		assert.Equal(t, 6.0, bounds[0].Last.Longitude)
	})

	t.Run("spans the tracks a paused lap is split across", func(t *testing.T) {
		db := TrainingCenterDatabase{Activities: []Activity{{Laps: []Lap{
			{Tracks: []Track{
				{Trackpoints: []Trackpoint{trackpoint("2026-08-30T13:13:30.000Z", 1, 2)}},
				{Trackpoints: []Trackpoint{trackpoint("2026-08-30T13:20:00.000Z", 3, 4)}},
			}},
		}}}}

		bounds, _, err := db.LapBounds()
		require.NoError(t, err)
		require.Len(t, bounds, 1)
		assert.Equal(t, 1.0, bounds[0].First.Latitude)
		assert.Equal(t, 3.0, bounds[0].Last.Latitude)
	})

	t.Run("ignores leading and trailing trackpoints without a position", func(t *testing.T) {
		db := TrainingCenterDatabase{Activities: []Activity{{Laps: []Lap{
			{Tracks: []Track{{Trackpoints: []Trackpoint{
				{Time: "2026-08-30T13:13:29.000Z"},
				trackpoint("2026-08-30T13:13:30.000Z", 1, 2),
				trackpoint("2026-08-30T13:13:31.000Z", 3, 4),
				{Time: "2026-08-30T13:13:32.000Z"},
			}}}},
		}}}}

		bounds, skipped, err := db.LapBounds()
		require.NoError(t, err)
		assert.Empty(t, skipped)
		require.Len(t, bounds, 1)
		assert.Equal(t, 1.0, bounds[0].First.Latitude)
		assert.Equal(t, 3.0, bounds[0].Last.Latitude)
	})

	t.Run("skips a lap that recorded no position, keeping the numbering of the rest", func(t *testing.T) {
		db := TrainingCenterDatabase{Activities: []Activity{{Laps: []Lap{
			{Tracks: []Track{{Trackpoints: []Trackpoint{trackpoint("2026-08-30T13:13:30.000Z", 1, 2)}}}},
			{Tracks: []Track{{Trackpoints: []Trackpoint{{Time: "2026-08-30T13:14:00.000Z"}}}}},
			{Tracks: []Track{{Trackpoints: []Trackpoint{trackpoint("2026-08-30T13:15:00.000Z", 3, 4)}}}},
		}}}}

		bounds, skipped, err := db.LapBounds()
		require.NoError(t, err)
		assert.Equal(t, []int{2}, skipped)
		require.Len(t, bounds, 2)
		assert.Equal(t, 1, bounds[0].Number)
		assert.Equal(t, 3, bounds[1].Number)
	})

	t.Run("numbers laps continuously across activities", func(t *testing.T) {
		lap := Lap{Tracks: []Track{{Trackpoints: []Trackpoint{trackpoint("2026-08-30T13:13:30.000Z", 1, 2)}}}}
		db := TrainingCenterDatabase{Activities: []Activity{
			{Laps: []Lap{lap, lap}},
			{Laps: []Lap{lap}},
		}}

		bounds, _, err := db.LapBounds()
		require.NoError(t, err)
		require.Len(t, bounds, 3)
		assert.Equal(t, []int{1, 2, 3}, []int{bounds[0].Number, bounds[1].Number, bounds[2].Number})
	})

	t.Run("reports an unparseable time against its lap", func(t *testing.T) {
		db := TrainingCenterDatabase{Activities: []Activity{{Laps: []Lap{
			{Tracks: []Track{{Trackpoints: []Trackpoint{trackpoint("2026-08-30T13:13:30.000Z", 1, 2)}}}},
			{Tracks: []Track{{Trackpoints: []Trackpoint{trackpoint("half past four", 3, 4)}}}},
		}}}}

		_, _, err := db.LapBounds()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lap 2")
		assert.Contains(t, err.Error(), `"half past four"`)
	})

	t.Run("returns nothing for a document with no laps", func(t *testing.T) {
		bounds, skipped, err := TrainingCenterDatabase{}.LapBounds()
		require.NoError(t, err)
		assert.Empty(t, bounds)
		assert.Empty(t, skipped)
	})
}

// TestUnmarshal covers the parts of the mapping that the struct tags alone
// decide: the namespaced document, the nested element paths and the optional
// elements.
func TestUnmarshal(t *testing.T) {
	const document = `<?xml version="1.0" encoding="UTF-8"?>
<TrainingCenterDatabase
  xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2"
  xmlns:ns3="http://www.garmin.com/xmlschemas/ActivityExtension/v2">
  <Activities>
    <Activity Sport="Biking">
      <Id>2026-08-30T13:13:30.000Z</Id>
      <Lap StartTime="2026-08-30T13:13:30.000Z">
        <TotalTimeSeconds>2603.254</TotalTimeSeconds>
        <TriggerMethod>Manual</TriggerMethod>
        <Track>
          <Trackpoint>
            <Time>2026-08-30T13:13:30.000Z</Time>
            <Position>
              <LatitudeDegrees>39.41267179325223</LatitudeDegrees>
              <LongitudeDegrees>-105.7585683837533</LongitudeDegrees>
            </Position>
            <AltitudeMeters>3033.800048828125</AltitudeMeters>
            <Extensions>
              <ns3:TPX><ns3:Speed>3.2</ns3:Speed></ns3:TPX>
            </Extensions>
          </Trackpoint>
          <Trackpoint>
            <Time>2026-08-30T13:13:31.000Z</Time>
          </Trackpoint>
        </Track>
      </Lap>
    </Activity>
  </Activities>
</TrainingCenterDatabase>`

	var db TrainingCenterDatabase
	require.NoError(t, xml.Unmarshal([]byte(document), &db))

	require.Len(t, db.Activities, 1)
	activity := db.Activities[0]
	assert.Equal(t, "Biking", activity.Sport)
	assert.Equal(t, "2026-08-30T13:13:30.000Z", activity.ID)

	require.Len(t, activity.Laps, 1)
	lap := activity.Laps[0]
	assert.Equal(t, "2026-08-30T13:13:30.000Z", lap.StartTime)
	require.Len(t, lap.Tracks, 1)
	require.Len(t, lap.Tracks[0].Trackpoints, 2)

	positioned := lap.Tracks[0].Trackpoints[0]
	require.NotNil(t, positioned.Position)
	assert.Equal(t, 39.41267179325223, positioned.Position.Latitude)
	assert.Equal(t, -105.7585683837533, positioned.Position.Longitude)
	require.NotNil(t, positioned.Altitude)
	assert.Equal(t, 3033.800048828125, *positioned.Altitude)

	// a trackpoint recorded without a GPS fix carries neither
	assert.Nil(t, lap.Tracks[0].Trackpoints[1].Position)
	assert.Nil(t, lap.Tracks[0].Trackpoints[1].Altitude)
}

func trackpoint(time string, lat, lon float64) Trackpoint {
	return Trackpoint{
		Time:     time,
		Position: &Position{Latitude: lat, Longitude: lon},
	}
}
