package hugoembedding

// CombineChunks, so that the lenght of the combined chunks
// is less than the size parameter
func CompressChunks(chunks *[]Chunk, size int) (*[]Chunk, error) {
	// Range an chunks and split it to the size
	startChunk := true
	endFlag := false
	combinedChunk := ""
	resultChunks := []Chunk{}
	for i, chunk := range *chunks {

		if startChunk {
			combinedChunk = *chunk.Chunk
			startChunk = false
		} else {
			combinedChunk = combinedChunk + *chunk.Chunk
		}
		if i == len(*chunks)-1 {
			endFlag = true
		}

		if len(combinedChunk) > size || endFlag {
			startChunk = true
			// Append the combined chunk to resultChunks
			line := combinedChunk
			resultChunks = append(resultChunks, Chunk{Chunk: &line})
			// Reset the combinedChunk
			combinedChunk = ""

		}

	}
	return &resultChunks, nil
}
