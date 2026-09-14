package buildcard_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/LittleDrongo/buildcard"
)

func ExampleSnapshot() {
	var info buildcard.Info = buildcard.Snapshot()
	fmt.Println(info.PID > 0)
	// Output: true
}

func TestSnapshotOwnsArgs(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"app.exe", "--port", "8080"}
	info := buildcard.Snapshot()
	if len(info.Args) != 2 || info.Args[0] != "--port" || info.Args[1] != "8080" {
		t.Fatalf("unexpected arguments: %v", info.Args)
	}
	info.Args[0] = "changed"
	if os.Args[1] != "--port" {
		t.Fatal("snapshot mutated process arguments")
	}
}
