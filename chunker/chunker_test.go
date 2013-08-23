package chunker

import (
	"testing"
)

func TestChunking(t *testing.T) {
	var chunks []string

	chunks = Chunk(7, 3, "abcdefghi")
	if len(chunks) != 3 ||
		chunks[0] != "7:0:3:abc" ||
		chunks[1] != "7:1:3:def" ||
		chunks[2] != "7:2:3:ghi" {
		t.Error("Bad chunking, got: ", chunks)
	}

	chunks = Chunk(11, 3, "")
	if len(chunks) != 0 {
		t.Error("Bad chunking, got: ", chunks)
	}

	chunks = Chunk(13, 3, "a")
	if len(chunks) != 1 || chunks[0] != "13:0:1:a" {
		t.Error("Bad chunking, got: ", chunks)
	}
}
