package robots

import (
	"bufio"
	"os"
	"testing"
)

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

func TestParse(t *testing.T) {
	_, err := useOne()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
