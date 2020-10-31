package main

import (
	"fmt"
	"sort"
	"testing"
)

type name struct {
	name string
	want string
}

type langs struct {
	user string
	repo string
	want []string
}

type contributors struct {
	user       string
	repo       string
	wantContri []string
	wantCommi  int
}

type info struct {
	params []string
	want   githubInfo
}

//TestCheckPath function for the CheckPath function
func TestCheckPath(t *testing.T) {
	validURI := []string{"github.com", "TheBluestOfDudes", "GameJam-Shooting-Stars"}
	invalidURIOne := []string{"Something", "SomethingElse", "SomethingSomething"}

	ok, _ := checkPath(validURI)
	if !ok {
		t.Errorf("checkPath(%s) = %t; want true", validURI, ok)
	}
	ok, _ = checkPath(invalidURIOne)
	if ok {
		t.Errorf("checkPath(%s) = %t; want false", invalidURIOne, ok)
	}
}

func TestGetName(t *testing.T) {
	names := []name{
		{"TheBluestOfDudes", "TheBluestOfDudes"},
		{"apache", "The Apache Software Foundation"},
		{"TheBluestOfDudesssssssss", ""},
	}

	for _, test := range names {
		testname := "Testing " + test.name
		t.Run(testname, func(t *testing.T) {
			n, _ := getName(test.name)
			if n != test.want {
				t.Errorf("got %s, want %s", n, test.want)
			}
		})
	}
}

func TestGetLanguages(t *testing.T) {
	languages := []langs{
		{"TheBluestOfDudes", "GameJam-Shooting-Stars", []string{"JavaScript", "HTML"}},
		{"TheBluestOfDudesssssssss", "NoMyRepo", []string{"message", "documentation_url"}},
	}
	for _, test := range languages {
		testname := fmt.Sprintf("Testing (%s/%s)", test.user, test.repo)
		t.Run(testname, func(t *testing.T) {
			langs, _ := getLanguages(test.user, test.repo)
			ok := compare(test.want, langs)
			if !ok {
				t.Errorf("got %s; want %s.", langs, test.want)
			}
		})
	}
}

func TestGetContributor(t *testing.T) {
	contris := []contributors{
		{"TheBluestOfDudes", "TurnBasedCombat", []string{"TheBluestOfDudes"}, 9},
		{"TheBluestOfDudesssssssss", "NoMyRepo", []string{}, 0},
	}
	for _, test := range contris {
		testname := fmt.Sprintf("Testing (%s/%s)", test.user, test.repo)
		t.Run(testname, func(t *testing.T) {
			contributes, commis, _ := getContributor(test.user, test.repo)
			ok := compare(test.wantContri, contributes)
			if !ok {
				t.Errorf("got %t on comparing contributor slices, want true.", ok)
			} else if commis != test.wantCommi {
				t.Errorf("got %d as contriubutions; want %d.", commis, test.wantCommi)
			} else if !ok && (commis != test.wantCommi) {
				t.Errorf("got contributors: %s and commits: %d; want %s contributors and %d commits.", contributes, commis, test.wantContri, test.wantCommi)
			}
		})
	}
}

func TestGetInfo(t *testing.T) {
	infoAr := []info{
		{[]string{"github.com", "TheBluestOfDudes", "GameJam-Shooting-Stars"}, githubInfo{
			"GameJam-Shooting-Stars",
			"TheBluestOfDudes",
			[]string{"TheBluestOfDudes", "VegardSkaansar"},
			10,
			[]string{"JavaScript", "HTML"},
		}},
		{[]string{"github.com", "TheBluestOfDudesssssssss", "NoMyRepo"}, githubInfo{
			"NoMyRepo",
			"Could not find owner name",
			[]string{"Could not find contributors"},
			0,
			[]string{"Could not find languages"},
		}},
	}
	for _, test := range infoAr {
		testname := fmt.Sprintf("Testing (%s/%s)", test.params[1], test.params[2])
		t.Run(testname, func(t *testing.T) {
			inf := getInfo(test.params)
			if inf.Owner != test.want.Owner {
				t.Errorf("got %s; want %s.", inf.Owner, test.want.Owner)
			} else if inf.Project != test.want.Project {
				t.Errorf("got %s; want %s.", inf.Project, test.want.Project)
			} else if inf.Commits != test.want.Commits {
				t.Errorf("got %d; want %d.", inf.Commits, test.want.Commits)
			} else if !compare(test.want.Languages, inf.Languages) {
				t.Errorf("got %s; want %s.", inf.Languages, test.want.Languages)
			} else if !compare(test.want.TopCommitter, inf.TopCommitter) {
				t.Errorf("got %s; want %s.", inf.TopCommitter, test.want.TopCommitter)
			}
		})
	}
}

func compare(one []string, two []string) bool {
	ok := true
	sort.Strings(one)
	sort.Strings(two)
	if len(one) == len(two) {
		for i, v := range one {
			if v != two[i] {
				ok = false
			}
		}
	} else {
		ok = false
	}
	return ok
}
