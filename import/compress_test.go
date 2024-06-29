package hugoembedding_test

import (
	"hugoembedding"
	"os"
	// "os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"gotest.tools/v3/assert"
)

func TestCompressChunks(t *testing.T) {

	inputTable := []hugoembedding.Chunk{
		{
			Chunk: aws.String(`Problem:
You want to develop a local browser app with the AWS SDK and you want to use your local AWS credentials. Although your current credentials are valid, the SDK does not accept them.`),
		},
		{
			Chunk: aws.String(`Solution:
Use cognito or use framework specific solutions to provide ENV variables to the SDK.
`),
		},
		{
			Chunk: aws.String(`Prerequisites:
Checkin on bash/ cli if your credentials are valid:
`),
		},
		{
			Chunk: aws.String(`aws sts get-caller-identity
`),
		},
	}

	// chunkLength := []int{187, 95, 67, 38}
	// 187+95=282
	// 187+95+67=349
	// 187+95+67+38=387

	// TestCase1
	// Size is larger than input 1, so 1 and 2 should be combined
	input := []hugoembedding.Chunk{inputTable[0], inputTable[1]}
	size := 200
	result, err := hugoembedding.CompressChunks(&input, size)
	assert.NilError(t, err)
	combindedChunk := *(inputTable[0].Chunk) + *(inputTable[1].Chunk)
	expected := []hugoembedding.Chunk{
		{
			Chunk: &combindedChunk,
		},
	}
	assert.DeepEqual(t, *expected[0].Chunk, *(*result)[0].Chunk)
	assert.Equal(t, len(expected), len((*result)))

}

func TestCompressLists(t *testing.T) {
	path := "testdata/how-to-use-pow/index.md"
	markdownFileContent, err := os.ReadFile(path)
	assert.NilError(t, err)
	chunks, err := hugoembedding.Parse(markdownFileContent)
	assert.NilError(t, err)
	assert.Equal(t, len(*chunks), 4)
}
