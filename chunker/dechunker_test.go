package chunker

import (
	"testing"
)

func TestChunkParsing(t *testing.T) {
	c, err := ParseChunk([]byte("123:047:00170:1:kjasdf:13!~ 0909 @#$)(`-"))
	if err != nil {
		t.Fatal(err)
	}
	if c.mid != 123 || c.cid != 47 || c.cnt != 170 {
		t.Fatal("Wrong params:", c)
	}
	if c.data != "1:kjasdf:13!~ 0909 @#$)(`-" {
		t.Fatal("Wrong data:", c)
	}
}

func TestChunkParsingMalformed(t *testing.T) {
	if c, err := ParseChunk([]byte("")); err == nil {
		t.Fatal("Expected error, got:", c)
	}
	if c, err := ParseChunk([]byte("a:b:c:d")); err == nil {
		t.Fatal("Expected error, got:", c)
	}
	if c, err := ParseChunk([]byte(":::::")); err == nil {
		t.Fatal("Expected error, got:", c)
	}
	if c, err := ParseChunk([]byte("1:0x2:3:asdf")); err == nil {
		t.Fatal("Expected error, got:", c)
	}
	if c, err := ParseChunk([]byte("123:47:17x:1:kjasdf:13!~ 0909 @#$)(`-")); err == nil {
		t.Fatal("Expected error, got:", c)
	}
	if c, err := ParseChunk([]byte("123:47:17:data")); err == nil { // cid > cnt
		t.Fatal("Expected error, got:", c)
	}
}

func TestMessageDechunking(t *testing.T) {
	var md MessageDechunker

	if md.AddChunk(chunk{1, 2, 3, "ghi"}) != false {
		t.Error("Expected incomplete dechunker")
	}
	if md.AddChunk(chunk{1, 0, 3, "abc"}) != false {
		t.Error("Expected incomplete dechunker")
	}
	if md.AddChunk(chunk{1, 1, 3, "def"}) != true {
		t.Error("Expected complete dechunker")
	}
	if res := md.Assemble(); res != "abcdefghi" {
		t.Error("Bad assembly: ", res)
	}
}

func TestMessageDechunkingMultipleMessages(t *testing.T) {
	var md MessageDechunker

	if md.AddChunk(chunk{1, 0, 1, "abc"}) != true || md.Assemble() != "abc" {
		t.Fail()
	}
	if md.AddChunk(chunk{2, 0, 1, "def"}) != true || md.Assemble() != "def" {
		t.Fail()
	}
	if md.AddChunk(chunk{3, 1, 2, "def"}) != false {
		t.Fail()
	}
	if md.AddChunk(chunk{3, 0, 2, "abc"}) != true || md.Assemble() != "abcdef" {
		t.Fail()
	}
	if md.AddChunk(chunk{4, 0, 2, "abc"}) != false {
		t.Fail()
	}
	if md.AddChunk(chunk{5, 1, 2, "def"}) != false {
		t.Fail()
	}
	md.Reset()
	if md.AddChunk(chunk{5, 0, 2, "def"}) != false {
		t.Fail()
	}
}
