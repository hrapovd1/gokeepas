package cli

import (
	"testing"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_printList(t *testing.T) {
	tests := []struct {
		name    string
		resp    *pb.ListResponse
		jsonFmt bool
	}{
		{"empty", &pb.ListResponse{Keys: ""}, false},
		{"empty json", &pb.ListResponse{Keys: ""}, true},
		{"text out", &pb.ListResponse{Keys: "'one','two'"}, false},
		{"json out", &pb.ListResponse{Keys: "'one','two'"}, true},
	}
	var logger = zap.New(nil)
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			require.NoError(t, printList(tst.resp, tst.jsonFmt, logger))
		})
	}
}
