package main

import "testing"

func TestExtractUser(t *testing.T) {
	type testCase struct {
		user string
		ok   bool
	}
	cases := map[string]testCase{
		"/trck": testCase{
			user: "",
			ok:   false,
		},
		"/trck/": testCase{
			user: "",
			ok:   false,
		},
		"/trck/user123": testCase{
			user: "user123",
			ok:   true,
		},
		"/trck/user123/": testCase{
			user: "",
			ok:   false,
		},
		"/trck/user123/user321": testCase{
			user: "",
			ok:   false,
		},
	}
	for path, testCase := range cases {
		user, err := extractUser(path)
		if user != testCase.user {
			t.Errorf("Expected user %s but got %s", testCase.user, user)
		}
		isOk := err == nil
		if !isOk && testCase.ok {
			t.Errorf("Expected an error")
		} else if isOk && !testCase.ok {
			t.Errorf("Unexpected error")
		}
	}
}
