// chunker provides simple message chunking/dechunking.
//
// Messages are assigned unique ids and split into chunks.  Each chunk is of
// the format:
//
//   Id:N:M:Data
//
// where:
//
//   Id   = message id
//   N    = chunk index
//   M    = number of chunks
//   Data = data for this chunk
//
// and Id, N, M are ints.
//
// For example, the message "abcdefghi" with id=77 broken into chunks of size 5 would
// result in the chunks:
//
//   77:0:2:abcde
//   77:1:2:fghi
//
package chunker

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type messageId int
type chunkId int
type chunkCount int

type chunk struct {
	mid  messageId
	cid  chunkId
	cnt  chunkCount
	data string
}

func (c chunk) MessageId() int { return int(c.mid) }
func (c chunk) ChunkId() int   { return int(c.cid) }
func (c chunk) NumChunks() int { return int(c.cnt) }
func (c chunk) Data() string   { return c.data }

func (c chunk) String() string {
	return fmt.Sprintf("%d:%d:%d:%s", c.mid, c.cid, c.cnt, c.data)
}

func crop(data []byte, N int) []byte {
	if N > len(data) {
		return data
	}
	return data[:N]
}

func recoverToError(retErr *error) {
	if err := recover(); err != nil {
		*retErr = err.(error)
	}
}

// Parse a message chunk.
func ParseChunk(data []byte) (c chunk, e error) {
	splits := bytes.SplitN(data, []byte(":"), 4)
	if len(splits) != 4 {
		return chunk{}, fmt.Errorf("Could not split data: %s", crop(data, 30))
	}
	defer recoverToError(&e)
	must := func(v int, e error) int {
		if e != nil {
			panic(e)
		}
		return v
	}

	mid := must(strconv.Atoi(string(splits[0])))
	cid := must(strconv.Atoi(string(splits[1])))
	cnt := must(strconv.Atoi(string(splits[2])))

	if cid >= cnt {
		return chunk{}, fmt.Errorf("Chunk id >= count: %d >= %d.  Data: %s", cid, cnt, crop(data, 30))
	}

	return chunk{
		mid:  messageId(mid),
		cid:  chunkId(cid),
		cnt:  chunkCount(cnt),
		data: string(splits[3]),
	}, nil
}

// MessageDechunker reassembles message chunks.  It keeps state for only a
// single message at a time.  Adding chunks for a different message id will
// reset the dechunker for the new messsage id.
type MessageDechunker struct {
	chunks map[chunkId]string
	cnt    chunkCount
	mid    messageId
}

// AddChunk accumulates the chunk for the message id specified in the chunk.
// If the chunk doesn't match the last message id of the dechunker, the
// dechunker is re-initialized for the new message.  This returns true if the
// chunk completes the message (which may be returned by Assemble()).
func (m *MessageDechunker) AddChunk(c chunk) (complete bool) {
	if m.chunks == nil || m.mid != c.mid || m.cnt != c.cnt {
		m.chunks = make(map[chunkId]string)
		m.cnt = c.cnt
		m.mid = c.mid
	}
	m.chunks[c.cid] = c.data
	return m.isComplete()
}

func (m *MessageDechunker) isComplete() bool {
	return chunkCount(len(m.chunks)) == m.cnt
}

// Assemble returns the current message with all known chunks combined.  If
// chunks are missing (the message is not yet complete), the returned message
// will silently ignore them.
func (m *MessageDechunker) Assemble() string {
	var msg []string
	for i := chunkId(0); i < chunkId(m.cnt); i++ {
		msg = append(msg, m.chunks[i])
	}
	return strings.Join(msg, "")
}

// Reset clears the dechunker state.
func (m *MessageDechunker) Reset() {
	m.chunks = nil
}
