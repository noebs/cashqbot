package main

import (
	"reflect"
	"testing"
)

func Test_necFormatter(t *testing.T) {
	type args struct {
		bInfo map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := necFormatter(tt.args.bInfo); got != tt.want {
				t.Errorf("necFormatter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_split(t *testing.T) {

	tests := []struct {
		name string
		args string
		want []string
	}{
		{"test splitter", "a b", []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := split(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dispatch(t *testing.T) {

	tests := []struct {
		name string
		args string
		want [][]string
	}{
		{"Test list of three items", "a 10 b 20 b 30", [][]string{{"a", "10"}, {"b", "20"}, {"b", "30"}}},
		{"list of one", "a 10", [][]string{{"a", "10"}}},
		{"list of three", "04203594959 1 04203594959 1 04203594959 1         ",
			[][]string{{"04203594959", "1"}, {"04203594959", "1"}, {"04203594959", "1"}}},
		{"Test list of two items", "a 10 b 20", [][]string{{"a", "10"}, {"b", "20"}}},
		{"Test list of three items", "a 10 b 20 b 30 c 2000 f 3232 d 11", [][]string{{"a", "10"},
			{"b", "20"}, {"b", "30"}, {"c", "2000"}, {"f", "3232"}, {"d", "11"}},
		},
		{"Test actual data", "04203594959 1 04203594959 1 04203594959 1",
			[][]string{{"04203594959", "1"}, {"04203594959", "1"}, {"04203594959", "1"}}},
		{"test long input", "04203594959 1 04203594959 1 04203594959 1 04203594959 1 04203594959 1",
			[][]string{{"04203594959", "1"}, {"04203594959", "1"}, {"04203594959", "1"}, {"04203594959", "1"}, {"04203594959", "1"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dispatch(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dispatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateDate(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"32323232", "2323233232"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateDate(); got != tt.want {
				t.Errorf("generateDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOdd(t *testing.T) {

	tests := []struct {
		name string
		args int
		want bool
	}{
		{"Even number", 4, false},
		{"Odd number", 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOdd(tt.args); got != tt.want {
				t.Errorf("isOdd() = %v, want %v", got, tt.want)
			}
		})
	}
}
