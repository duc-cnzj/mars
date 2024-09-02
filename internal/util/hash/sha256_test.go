package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashWithValidData(t *testing.T) {
	data := "Hello, World!"
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	hash := Hash(data)

	assert.Equal(t, expectedHash, hash)
}

func TestHashWithEmptyData(t *testing.T) {
	data := ""
	expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	hash := Hash(data)

	assert.Equal(t, expectedHash, hash)
}

func TestHashWithDifferentDataSameLength(t *testing.T) {
	data1 := "Hello, World!"
	data2 := "World, Hello!"

	hash1 := Hash(data1)
	hash2 := Hash(data2)

	assert.NotEqual(t, hash1, hash2)
}
