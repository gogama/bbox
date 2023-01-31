package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// For now it just reads lines in format LON LAT, interprets them
// points, and adds them to orb.
//
// TODO: Should eventually be able to read WKT, WKB, GeoJSON
func main() {
	s := bufio.NewScanner(os.Stdin)
	var err error
	var n int64
	if !s.Scan() {
		if err = s.Err(); err != nil {
			exitWithError(err, n, nil)
		}
		n++
		return
	}
	var p orb.Point
	if p, err = toPoint(s.Bytes()); err != nil {
		exitWithError(err, n, s.Bytes())
	}
	var b orb.Bound = p.Bound()
	for s.Scan() {
		if err = s.Err(); err != nil {
			exitWithError(err, n, nil)
		} else if p, err = toPoint(s.Bytes()); err != nil {
			exitWithError(err, n, s.Bytes())
		} else {
			b = b.Extend(p)
		}
	}
	emitBBox(os.Stdout, b)
}

func exitWithError(err error, n int64, text []byte) {
	if text == nil {
		fmt.Fprintf(os.Stderr, "%d: %s\n", n+1, err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "%d: %s\n  [%s]\n", n+1, err.Error(), text)
	}
	os.Exit(1)
}
func toPoint(text []byte) (orb.Point, error) {
	var i, j int
	// Skip leading whitespace.
	for i < len(text) && isSpace(text[i]) {
		i++
	}
	if i == len(text) {
		return orb.Point{}, errors.New("blank line")
	}
	// Parse longitude.
	for j = i + 1; j < len(text) && !isSpace(text[j]); j++ {
	}
	lon, err := strconv.ParseFloat(string(text[i:j]), 64)
	if err != nil {
		return orb.Point{}, fmt.Errorf("bad longitude (at column %d): %s", i+1, err.Error())
	}
	// Skip additional whitespace.
	for i = j + 1; i < len(text) && isSpace(text[i]); i++ {
	}
	if i >= len(text) {
		return orb.Point{}, errors.New("missing latitude")
	}
	// Parse latitude.
	for j = i + 1; j < len(text) && !isSpace(text[j]); j++ {
	}
	lat, err := strconv.ParseFloat(string(text[i:j]), 64)
	if err != nil {
		return orb.Point{}, fmt.Errorf("bad latitude (at column %d): %s", i+1, err.Error())
	}
	// Parse trailing whitespace.
	for i = j + 1; i < len(text) && isSpace(text[i]); i++ {
	}
	if i < len(text) {
		return orb.Point{}, fmt.Errorf("unexpected trailing text (at column %d)", j+1)
	}
	// Return the point.
	return orb.Point{lon, lat}, nil
}

func isSpace(x byte) bool {
	switch x {
	case ' ', '\t':
		return true
	default:
		return false
	}
}
func emitBBox(out io.Writer, b orb.Bound) {
	f := geojson.NewFeature(b.ToPolygon())
	f.BBox = geojson.NewBBox(b)
	p, err := f.MarshalJSON()
	if err != nil {
		panic(err)
	}
	out.Write(p)
}
