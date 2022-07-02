package robots

import (
	"bufio"
	"os"
	"testing"
)

// benchmarks the IsAllowed function
func BenchmarkIsAllowed(b *testing.B) {
	b.StopTimer()

	s, err := useOne()
	if err != nil {
		b.Log(err)
		b.Fail()
	}

	b.StartTimer()
	priv := s.IsAllowed("/private/", "*")
	if priv != false {
		b.Fail()
	}
}

// internal, loads a robots.txt file for use
func useOne() (RobotList, error) {
	f, err := os.Open("./examples/t.txt")
	if err != nil {
		return RobotList{}, err
	}

	b := bufio.NewReader(f)
	s, err := parse(b)
	if err != nil {
		return RobotList{}, err
	}
	return s, nil
}

// tests whether the IsAllowed function properly determines match strings
func TestIsAllowed(t *testing.T) {
	s, err := useOne()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	priv := s.IsAllowed("/private/", "*")
	user := s.IsAllowed("/user/hvlck", "*")
	if !(!priv && user) {
		t.Fail()
	}
}

// tests whether the parse function successfully parses a robots.txt file
func TestParse(t *testing.T) {
	s, err := useOne()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	proper := RobotList{
		robots: map[string]Robot{
			"*": {
				Allowed: map[string]bool{
					"/user/*":   true,
					"/private/": false,
				},
			},
		},
		sitemaps: []string{"https://example.com/"},
	}

	for k := range proper.robots {
		if ent, ok := proper.robots[k]; ok {
			for kk, vv := range ent.Allowed {
				if s.robots[k].Allowed[kk] != vv {
					t.Fail()
				}
			}
		}
	}

	for i, v := range proper.sitemaps {
		if s.sitemaps[i] != v {
			t.Fail()
		}
	}
}
