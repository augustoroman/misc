package chunker

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Chunk cuts up a message into chunks of the specified size.  Each chunk is
// prefixed with the message id, chunk id, and chunk count:
//    Id:N:M:ChunkData
func Chunk(msgid, chunksize int, msg string) []string {
	N := (len(msg) + chunksize - 1) / chunksize
	chunks := make([]string, N)
	for i := 0; i < N; i++ {
		chunks[i] = chunk{
			mid:  messageId(msgid),
			cid:  chunkId(i),
			cnt:  chunkCount(N),
			data: msg[i*chunksize : min((i+1)*chunksize, len(msg))],
		}.String()
	}
	return chunks
}
