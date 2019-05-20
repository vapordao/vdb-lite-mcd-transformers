package queries

import (
	"database/sql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/mcd_transformers/test_config"
	helper "github.com/vulcanize/mcd_transformers/transformers/component_tests/queries/test_helpers"
	"github.com/vulcanize/mcd_transformers/transformers/shared/constants"
	"github.com/vulcanize/mcd_transformers/transformers/storage/vat"
	"github.com/vulcanize/mcd_transformers/transformers/test_data"
	"github.com/vulcanize/vulcanizedb/libraries/shared/storage/utils"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/fakes"
	"math/rand"
	"strconv"
)

var _ = Describe("Urn history query", func() {
	var (
		db         *postgres.DB
		vatRepo    vat.VatStorageRepository
		headerRepo repositories.HeaderRepository
		fakeUrn    string
	)

	BeforeEach(func() {
		db = test_config.NewTestDB(test_config.NewTestNode())
		headerRepo = repositories.NewHeaderRepository(db)
		vatRepo = vat.VatStorageRepository{}
		vatRepo.SetDB(db)

		fakeUrn = test_data.RandomString(5)
	})

	It("returns a reverse chronological history for the given ilk and urn", func() {
		blockOne := rand.Int()
		timestampOne := int(rand.Int31())
		urnSetupData := helper.GetUrnSetupData(blockOne, timestampOne)
		urnMetadata := helper.GetUrnMetadata(helper.FakeIlk.Hex, fakeUrn)
		helper.CreateUrn(urnSetupData, urnMetadata, vatRepo, headerRepo)

		inkBlockOne := urnSetupData.Ink
		artBlockOne := urnSetupData.Art

		expectedRatioBlockOne := helper.GetExpectedRatio(inkBlockOne, urnSetupData.Spot, artBlockOne, urnSetupData.Rate)
		expectedTimestampOne := helper.GetExpectedTimestamp(timestampOne)
		expectedUrnBlockOne := helper.UrnState{
			UrnGuy:      fakeUrn,
			IlkName:     helper.FakeIlk.Name,
			BlockHeight: blockOne,
			Ink:         strconv.Itoa(inkBlockOne),
			Art:         strconv.Itoa(artBlockOne),
			Ratio:       sql.NullString{String: strconv.FormatFloat(expectedRatioBlockOne, 'f', 8, 64), Valid: true},
			Safe:        expectedRatioBlockOne >= 1,
			Created:     sql.NullString{String: expectedTimestampOne, Valid: true},
			Updated:     sql.NullString{String: expectedTimestampOne, Valid: true},
		}

		// New block
		blockTwo := blockOne + 1
		timestampTwo := timestampOne + 1
		createFakeHeader(blockTwo, timestampTwo, headerRepo)

		// Relevant ink diff in block two
		inkBlockTwo := rand.Int()
		err := vatRepo.Create(blockTwo, fakes.FakeHash.String(), urnMetadata.UrnInk, strconv.Itoa(inkBlockTwo))
		Expect(err).NotTo(HaveOccurred())

		// Irrelevant art diff in block two
		wrongUrn := test_data.RandomString(5)
		wrongArt := strconv.Itoa(rand.Int())
		wrongMetadata := utils.GetStorageValueMetadata(vat.UrnArt,
			map[utils.Key]string{constants.Ilk: helper.FakeIlk.Hex, constants.Guy: wrongUrn}, utils.Uint256)
		err = vatRepo.Create(blockOne, fakes.FakeHash.String(), wrongMetadata, wrongArt)
		Expect(err).NotTo(HaveOccurred())

		expectedRatioBlockTwo := helper.GetExpectedRatio(inkBlockTwo, urnSetupData.Spot, artBlockOne, urnSetupData.Rate)
		expectedTimestampTwo := helper.GetExpectedTimestamp(timestampTwo)
		expectedUrnBlockTwo := helper.UrnState{
			UrnGuy:      fakeUrn,
			IlkName:     helper.FakeIlk.Name,
			BlockHeight: blockTwo,
			Ink:         strconv.Itoa(inkBlockTwo),
			Art:         strconv.Itoa(artBlockOne),
			Ratio:       sql.NullString{String: strconv.FormatFloat(expectedRatioBlockTwo, 'f', 8, 64), Valid: true},
			Safe:        expectedRatioBlockTwo >= 1,
			Created:     sql.NullString{String: expectedTimestampOne, Valid: true},
			Updated:     sql.NullString{String: expectedTimestampTwo, Valid: true},
		}

		// New block
		blockThree := blockTwo + 1
		timestampThree := timestampTwo + 1
		createFakeHeader(blockThree, timestampThree, headerRepo)

		// Relevant art diff in block three
		artBlockThree := 0
		err = vatRepo.Create(blockThree, fakes.FakeHash.String(), urnMetadata.UrnArt, strconv.Itoa(artBlockThree))
		Expect(err).NotTo(HaveOccurred())

		expectedTimestampThree := helper.GetExpectedTimestamp(timestampThree)
		expectedUrnBlockThree := helper.UrnState{
			UrnGuy:      fakeUrn,
			IlkName:     helper.FakeIlk.Name,
			BlockHeight: blockThree,
			Ink:         strconv.Itoa(inkBlockTwo),
			Art:         strconv.Itoa(artBlockThree),
			Ratio:       sql.NullString{Valid: false}, // 0 art => null ratio
			Safe:        true,                         // 0 art => safe urn
			Created:     sql.NullString{String: expectedTimestampOne, Valid: true},
			Updated:     sql.NullString{String: expectedTimestampThree, Valid: true},
		}

		var result []helper.UrnState
		dbErr := db.Select(&result,
			`SELECT * FROM api.all_urn_states($1, $2, $3)`,
			helper.FakeIlk.Name, fakeUrn, blockThree)
		Expect(dbErr).NotTo(HaveOccurred())

		// Reverse chronological order
		helper.AssertUrn(result[0], expectedUrnBlockThree)
		helper.AssertUrn(result[1], expectedUrnBlockTwo)
		helper.AssertUrn(result[2], expectedUrnBlockOne)
	})
})

func createFakeHeader(blockNumber, timestamp int, headerRepo repositories.HeaderRepository) {
	fakeHeaderOne := fakes.GetFakeHeader(int64(blockNumber))
	fakeHeaderOne.Timestamp = strconv.Itoa(timestamp)

	_, headerErr := headerRepo.CreateOrUpdateHeader(fakeHeaderOne)
	Expect(headerErr).NotTo(HaveOccurred())
}
