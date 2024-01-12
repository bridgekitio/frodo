package codec_test

import "time"

type testStruct struct {
	String       string
	Int          int
	Int8         int8
	Float64      float64
	Bool         bool
	User         *testStructUser
	RemappedUser *testStructUser `json:"alias"`
}

type testStructUser struct {
	ID         string `json:",omitempty"`
	Name       string `json:"goes_by"`
	Ignore     string `json:"-"`
	AuditTrail testStructTimestamp
}

type testStructTimestamp struct {
	Deleted  bool
	Created  time.Time
	Modified time.Time
}
