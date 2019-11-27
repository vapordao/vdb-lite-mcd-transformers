package test_helpers

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/makerdao/vulcanizedb/pkg/core"

	"github.com/makerdao/vulcanizedb/pkg/datastore/postgres"
	"github.com/makerdao/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/makerdao/vulcanizedb/pkg/fakes"
	. "github.com/onsi/gomega"
)

func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

func CreateHeader(timestamp int64, blockNumber int, db *postgres.DB) core.Header {
	return CreateHeaderWithHash(strconv.Itoa(rand.Int()), timestamp, blockNumber, db)
}

func CreateHeaderWithHash(hash string, timestamp int64, blockNumber int, db *postgres.DB) core.Header {
	fakeHeader := fakes.GetFakeHeaderWithTimestamp(timestamp, int64(blockNumber))
	fakeHeader.Hash = hash
	headerRepo := repositories.NewHeaderRepository(db)
	headerId, headerErr := headerRepo.CreateOrUpdateHeader(fakeHeader)
	Expect(headerErr).NotTo(HaveOccurred())
	fakeHeader.Id = headerId
	return fakeHeader
}