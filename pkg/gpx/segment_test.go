package gpx_test

import (
	"math"
	"testing"

	"github.com/glynternet/gpx/pkg/gpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// route builds n points with distinct, ordered coordinates so that chunk
// boundaries and ordering are identifiable in assertions.
func route(n int) []gpxgo.GPXPoint {
	points := make([]gpxgo.GPXPoint, n)
	for i := range points {
		points[i] = gpxgo.GPXPoint{Point: gpxgo.Point{Latitude: float64(i), Longitude: float64(i)}}
	}
	return points
}

// lats reduces points to their latitudes, which route sets to the point index.
func lats(points []gpxgo.GPXPoint) []float64 {
	out := make([]float64, len(points))
	for i, p := range points {
		out[i] = p.Latitude
	}
	return out
}

func TestSplitPoints(t *testing.T) {
	t.Run("rejects zero chunks", func(t *testing.T) {
		_, err := gpx.SplitPoints(route(10), 0, 0)
		assert.EqualError(t, err, "chunks must be greater than 0")
	})

	t.Run("rejects preoverlap outside 0-100, NaN included", func(t *testing.T) {
		for _, percentage := range []float32{
			-1, 101,
			float32(math.NaN()),
			float32(math.Inf(1)),
			float32(math.Inf(-1)),
		} {
			_, err := gpx.SplitPoints(route(10), 5, percentage)
			assert.EqualError(t, err, "preoverlapPercentage must be between 0-100, inclusive")
		}
	})

	t.Run("chunks the route", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			points int
			chunks uint
			want   [][]float64
		}{
			{
				name: "one segment per chunk", points: 6, chunks: 5,
				want: [][]float64{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}},
			},
			{
				name: "segments divide exactly", points: 7, chunks: 3,
				want: [][]float64{{0, 1, 2}, {2, 3, 4}, {4, 5, 6}},
			},
			{
				name: "segments divide with a remainder", points: 8, chunks: 3,
				want: [][]float64{{0, 1, 2}, {2, 3, 4}, {4, 5, 6, 7}},
			},
			{
				name: "fewer segments than chunks", points: 3, chunks: 5,
				want: [][]float64{{0, 1}, {1, 2}},
			},
			{
				name: "single chunk keeps the whole route", points: 4, chunks: 1,
				want: [][]float64{{0, 1, 2, 3}},
			},
			{
				name: "single segment", points: 2, chunks: 5,
				want: [][]float64{{0, 1}},
			},
			{
				name: "single point route", points: 1, chunks: 5,
				want: [][]float64{{0}},
			},
			{
				name: "empty route", points: 0, chunks: 5,
				want: nil,
			},
			{
				// A chunks value too large for an int must not collapse the split.
				name: "more chunks than a uint can hold segments", points: 6, chunks: ^uint(0),
				want: [][]float64{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				points := route(tc.points)
				out, err := gpx.SplitPoints(points, tc.chunks, 0)
				require.NoError(t, err)

				var got [][]float64
				for _, chunk := range out {
					got = append(got, lats(chunk))
				}
				assert.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("covers every point of a route that does not divide exactly", func(t *testing.T) {
		const points, chunks = 1326, 5
		out, err := gpx.SplitPoints(route(points), chunks, 0)
		require.NoError(t, err)
		require.Len(t, out, chunks)

		// Concatenating the chunks, less each shared boundary point, must
		// reproduce the route exactly: nothing dropped, nothing duplicated.
		var got []float64
		for i, chunk := range out {
			if i == 0 {
				got = append(got, lats(chunk)...)
				continue
			}
			assert.Equal(t, out[i-1][len(out[i-1])-1], chunk[0],
				"chunk %d must begin on the point chunk %d ended on", i, i-1)
			got = append(got, lats(chunk)[1:]...)
		}
		assert.Equal(t, lats(route(points)), got)
	})

	t.Run("preoverlap reaches each chunk back over its predecessor", func(t *testing.T) {
		// 12 segments over 3 chunks is 4 segments each; 50% reaches back 2.
		out, err := gpx.SplitPoints(route(13), 3, 50)
		require.NoError(t, err)

		var got [][]float64
		for _, chunk := range out {
			got = append(got, lats(chunk))
		}
		assert.Equal(t, [][]float64{
			{0, 1, 2, 3, 4},          // nothing to reach back into
			{2, 3, 4, 5, 6, 7, 8},    // reaches back to 2
			{6, 7, 8, 9, 10, 11, 12}, // reaches back to 6
		}, got)
	})
}

// FuzzSplitPoints exercises the properties that must hold for any route and
// chunk count: the chunks run the length of the route in order, each is a
// contiguous run of it, and none carries more than one segment more than another.
func FuzzSplitPoints(f *testing.F) {
	for _, seed := range []struct {
		points     int
		chunks     uint
		preoverlap float32
	}{
		{6, 5, 0}, {1326, 5, 0}, {14, 5, 0}, {3, 5, 0}, {2, 1, 0},
		{1, 5, 0}, {0, 5, 0}, {13, 3, 50}, {100, 7, 100}, {50, 1, 0},
	} {
		f.Add(seed.points, seed.chunks, seed.preoverlap)
	}

	f.Fuzz(func(t *testing.T, numPoints int, chunks uint, preoverlap float32) {
		// Bound the route to a size worth building; the properties do not depend
		// on it being large.
		if numPoints < 0 || numPoints > 5000 {
			t.Skip()
		}

		out, err := gpx.SplitPoints(route(numPoints), chunks, preoverlap)
		if err != nil {
			return // rejections are documented behaviour
		}

		if numPoints == 0 {
			assert.Empty(t, out)
			return
		}

		wantChunks := 1
		if numPoints > 1 {
			wantChunks = int(min(uint(numPoints-1), chunks))
		}
		require.Len(t, out, wantChunks)

		assert.Equal(t, float64(0), out[0][0].Latitude, "first chunk must start at the route start")
		last := out[len(out)-1]
		assert.Equal(t, float64(numPoints-1), last[len(last)-1].Latitude, "last chunk must end at the route end")

		shortest, longest := math.MaxInt, 0
		for i, chunk := range out {
			require.NotEmpty(t, chunk, "chunk %d is empty", i)

			// route numbers its points by index, so a chunk of it must be a run of
			// consecutive latitudes.
			for j, p := range chunk {
				require.Equal(t, chunk[0].Latitude+float64(j), p.Latitude,
					"chunk %d is not a contiguous run of the route at %d", i, j)
			}
			if i > 0 {
				prev := out[i-1]
				assert.Greater(t, chunk[len(chunk)-1].Latitude, prev[len(prev)-1].Latitude,
					"chunk %d must advance beyond chunk %d", i, i-1)
			}
			segments := len(chunk) - 1
			shortest, longest = min(shortest, segments), max(longest, segments)
		}

		if numPoints > 1 {
			assert.GreaterOrEqual(t, shortest, 1, "every chunk must span at least one segment")
		}
		if preoverlap == 0 {
			// Reaching back over a predecessor deliberately makes chunks uneven, so
			// even spread is a property of the un-overlapped split.
			assert.LessOrEqual(t, longest-shortest, 1, "segment counts must differ by at most one")

			for i := 1; i < len(out); i++ {
				prev := out[i-1]
				assert.Equal(t, prev[len(prev)-1].Latitude, out[i][0].Latitude,
					"chunk %d must begin on the point chunk %d ended on", i, i-1)
			}
		}
	})
}
