package interop

import (
	"testing"
)

func TestFullScore(t *testing.T) {
	f1 := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\interop\test_data\gtc\H6E2Y9U2J7U6Y2U_202151180137_R12C02.gtc`
	f2 := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\interop\test_data\gtc\H9S5P6E2D4S2T5J_202687840174_R02C01.gtc`
	s, err := Score(f1, f2)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(`score`, s)
}
