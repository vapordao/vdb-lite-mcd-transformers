package queries

import (
	"math/rand"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/fakes"

	"github.com/vulcanize/mcd_transformers/test_config"
	"github.com/vulcanize/mcd_transformers/transformers/component_tests/queries/test_helpers"
	"github.com/vulcanize/mcd_transformers/transformers/events/spot_poke"
	"github.com/vulcanize/mcd_transformers/transformers/shared"
	"github.com/vulcanize/mcd_transformers/transformers/test_data"
)

var _ = Describe("all poke events query", func() {
	var (
		db                 *postgres.DB
		spotPokeRepo       spot_poke.SpotPokeRepository
		headerRepo         repositories.HeaderRepository
		beginningTimeRange int64
		endingTimeRange    int64
	)

	BeforeEach(func() {
		db = test_config.NewTestDB(test_config.NewTestNode())
		test_config.CleanTestDB(db)

		beginningTimeRange = int64(test_helpers.GetRandomInt(1558710000, 1558720000))
		endingTimeRange = int64(test_helpers.GetRandomInt(1558720001, 1558730000))
		headerRepo = repositories.NewHeaderRepository(db)
		spotPokeRepo = spot_poke.SpotPokeRepository{}
		spotPokeRepo.SetDB(db)
		rand.Seed(GinkgoRandomSeed())
	})

	AfterEach(func() {
		closeErr := db.Close()
		Expect(closeErr).NotTo(HaveOccurred())
	})

	It("returns poke events in different blocks between a time range", func() {
		fakeHeaderOne := fakes.GetFakeHeaderWithTimestamp(beginningTimeRange, int64(test_data.EthSpotPokeLog.BlockNumber))
		headerID, err := headerRepo.CreateOrUpdateHeader(fakeHeaderOne)
		Expect(err).NotTo(HaveOccurred())

		spotPoke := generateSpotPoke(test_helpers.FakeIlk.Hex, 1)
		ilkIdBlockOne, err := shared.GetOrCreateIlk(spotPoke.Ilk, db)
		err = spotPokeRepo.Create(headerID, []interface{}{spotPoke})
		Expect(err).NotTo(HaveOccurred())

		fakeHeaderTwo := fakes.GetFakeHeaderWithTimestamp(endingTimeRange, fakeHeaderOne.BlockNumber+1)
		anotherHeaderID, err := headerRepo.CreateOrUpdateHeader(fakeHeaderTwo)
		Expect(err).NotTo(HaveOccurred())

		anotherSpotPoke := generateSpotPoke(test_helpers.AnotherFakeIlk.Hex, 1)
		anotherIlkId, err := shared.GetOrCreateIlk(anotherSpotPoke.Ilk, db)
		Expect(err).NotTo(HaveOccurred())
		err = spotPokeRepo.Create(anotherHeaderID, []interface{}{anotherSpotPoke})
		Expect(err).NotTo(HaveOccurred())

		expectedValues := []test_helpers.PokeEvent{
			{
				IlkId: strconv.Itoa(anotherIlkId),
				Val:   anotherSpotPoke.Value,
				Spot:  anotherSpotPoke.Spot,
			},
			{
				IlkId: strconv.Itoa(ilkIdBlockOne),
				Val:   spotPoke.Value,
				Spot:  spotPoke.Spot,
			},
		}

		var dbPokeEvents []test_helpers.PokeEvent
		err = db.Select(&dbPokeEvents, `SELECT ilk_id, val, spot FROM api.all_poke_events($1, $2)`, beginningTimeRange, endingTimeRange)
		Expect(err).NotTo(HaveOccurred())
		Expect(dbPokeEvents).To(Equal(expectedValues))
	})

	It("returns poke events with transactions in the same block", func() {
		fakeHeaderOne := fakes.GetFakeHeaderWithTimestamp(beginningTimeRange, int64(test_data.EthSpotPokeLog.BlockNumber))
		headerID, err := headerRepo.CreateOrUpdateHeader(fakeHeaderOne)
		Expect(err).NotTo(HaveOccurred())

		spotPoke := generateSpotPoke(test_helpers.FakeIlk.Hex, 1)
		ilkIdBlockOne, err := shared.GetOrCreateIlk(spotPoke.Ilk, db)
		err = spotPokeRepo.Create(headerID, []interface{}{spotPoke})
		Expect(err).NotTo(HaveOccurred())

		anotherSpotPoke := generateSpotPoke(test_helpers.AnotherFakeIlk.Hex, 1)
		anotherSpotPoke.TransactionIndex = spotPoke.TransactionIndex + 1
		anotherIlkId, err := shared.GetOrCreateIlk(anotherSpotPoke.Ilk, db)
		Expect(err).NotTo(HaveOccurred())
		err = spotPokeRepo.Create(headerID, []interface{}{anotherSpotPoke})
		Expect(err).NotTo(HaveOccurred())

		expectedValues := []test_helpers.PokeEvent{
			{
				IlkId: strconv.Itoa(ilkIdBlockOne),
				Val:   spotPoke.Value,
				Spot:  spotPoke.Spot,
			},
			{
				IlkId: strconv.Itoa(anotherIlkId),
				Val:   anotherSpotPoke.Value,
				Spot:  anotherSpotPoke.Spot,
			},
		}

		var dbPokeEvents []test_helpers.PokeEvent
		err = db.Select(&dbPokeEvents, `SELECT ilk_id, val, spot FROM api.all_poke_events($1, $2)`, beginningTimeRange, endingTimeRange)
		Expect(err).NotTo(HaveOccurred())
		Expect(dbPokeEvents).To(ConsistOf(expectedValues))
	})

	It("ignores poke events not in time range", func() {
		fakeHeaderOne := fakes.GetFakeHeaderWithTimestamp(beginningTimeRange, int64(test_data.EthSpotPokeLog.BlockNumber))
		headerID, err := headerRepo.CreateOrUpdateHeader(fakeHeaderOne)
		Expect(err).NotTo(HaveOccurred())

		spotPoke := generateSpotPoke(test_helpers.FakeIlk.Hex, 1)
		ilkIdBlockOne, err := shared.GetOrCreateIlk(spotPoke.Ilk, db)
		err = spotPokeRepo.Create(headerID, []interface{}{spotPoke})
		Expect(err).NotTo(HaveOccurred())

		fakeHeaderTwo := fakes.GetFakeHeaderWithTimestamp(endingTimeRange+1, fakeHeaderOne.BlockNumber+1)
		anotherHeaderID, err := headerRepo.CreateOrUpdateHeader(fakeHeaderTwo)
		Expect(err).NotTo(HaveOccurred())

		anotherSpotPoke := generateSpotPoke(test_helpers.AnotherFakeIlk.Hex, 1)
		_, err = shared.GetOrCreateIlk(anotherSpotPoke.Ilk, db)
		Expect(err).NotTo(HaveOccurred())
		err = spotPokeRepo.Create(anotherHeaderID, []interface{}{anotherSpotPoke})
		Expect(err).NotTo(HaveOccurred())

		expectedValues := []test_helpers.PokeEvent{
			{
				IlkId: strconv.Itoa(ilkIdBlockOne),
				Val:   spotPoke.Value,
				Spot:  spotPoke.Spot,
			},
		}

		var dbPokeEvents []test_helpers.PokeEvent
		err = db.Select(&dbPokeEvents, `SELECT ilk_id, val, spot FROM api.all_poke_events($1, $2)`, beginningTimeRange, endingTimeRange)
		Expect(err).NotTo(HaveOccurred())
		Expect(dbPokeEvents).To(Equal(expectedValues))
	})

	It("limits results to latest blocks if max_results argument is provided", func() {
		fakeHeaderOne := fakes.GetFakeHeaderWithTimestamp(beginningTimeRange, int64(test_data.EthSpotPokeLog.BlockNumber))
		headerID, headerOneErr := headerRepo.CreateOrUpdateHeader(fakeHeaderOne)
		Expect(headerOneErr).NotTo(HaveOccurred())

		oldSpotPoke := generateSpotPoke(test_helpers.FakeIlk.Hex, 1)
		ilkId, ilkErr := shared.GetOrCreateIlk(oldSpotPoke.Ilk, db)
		Expect(ilkErr).NotTo(HaveOccurred())
		oldSpotPokeErr := spotPokeRepo.Create(headerID, []interface{}{oldSpotPoke})
		Expect(oldSpotPokeErr).NotTo(HaveOccurred())

		fakeHeaderTwo := fakes.GetFakeHeaderWithTimestamp(endingTimeRange, fakeHeaderOne.BlockNumber+1)
		anotherHeaderID, headerTwoErr := headerRepo.CreateOrUpdateHeader(fakeHeaderTwo)
		Expect(headerTwoErr).NotTo(HaveOccurred())

		recentSpotPoke := generateSpotPoke(test_helpers.FakeIlk.Hex, 2)
		recentSpotPokeErr := spotPokeRepo.Create(anotherHeaderID, []interface{}{recentSpotPoke})
		Expect(recentSpotPokeErr).NotTo(HaveOccurred())

		maxResults := 1
		var dbPokeEvents []test_helpers.PokeEvent
		selectErr := db.Select(&dbPokeEvents, `SELECT ilk_id, val, spot FROM api.all_poke_events($1, $2, $3)`,
			beginningTimeRange, endingTimeRange, maxResults)
		Expect(selectErr).NotTo(HaveOccurred())

		Expect(dbPokeEvents).To(ConsistOf(test_helpers.PokeEvent{
			IlkId: strconv.Itoa(ilkId),
			Val:   recentSpotPoke.Value,
			Spot:  recentSpotPoke.Spot,
		}))
	})

	It("uses default arguments when none are passed in", func() {
		_, err := db.Exec(`SELECT * FROM api.all_poke_events()`)
		Expect(err).NotTo(HaveOccurred())
	})
})

func generateSpotPoke(ilk string, seed int) spot_poke.SpotPokeModel {
	spotPoke := test_data.SpotPokeModel
	spotPoke.Ilk = ilk
	spotPoke.Value = strconv.Itoa(1 + seed)
	spotPoke.Spot = strconv.Itoa(2 + seed)
	return spotPoke
}
