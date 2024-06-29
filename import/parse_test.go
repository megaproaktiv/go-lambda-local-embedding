package hugoembedding_test

import (
	"hugoembedding"
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

const testMarkdown_001 = "testdata/react-aws-authentication/index.md"

func TestParseNumberParagraphs(t *testing.T) {
	t.Logf("Read file: %s", testMarkdown_001)
	content, err := os.ReadFile(testMarkdown_001)
	assert.NilError(t, err)

	t.Logf("Call parse \n")
	chunks, err := hugoembedding.Parse(content)
	assert.NilError(t, err)

	assert.Equal(t, len(*chunks), 10)
}
