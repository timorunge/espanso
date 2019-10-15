package espanso

import "testing"

var testDict = []string{
	"trigger", "replace",
	"foo", "bar",
	":br", "Best Regards,\nJon Snow",
	"newline", "\n",
	"quot", "\"",
}

func TestDictToMatches(t *testing.T) {
	i := 0
	for _, match := range DictToMatches(testDict) {
		if toRaw(testDict[i]) != match.Trigger() {
			t.Error(
				"Expected result is", testDict[i],
				"but got", match.Trigger(),
			)
		}
		if toRaw(testDict[i+1]) != match.Replace() {
			t.Error(
				"Expected result is", testDict[i+1],
				"but got", match.Replace(),
			)
		}
		i += 2
	}
}
