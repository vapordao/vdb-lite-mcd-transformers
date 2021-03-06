// VulcanizeDB
// Copyright © 2018 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package new_cdp

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/makerdao/vdb-mcd-transformers/transformers/shared"
	"github.com/makerdao/vdb-mcd-transformers/transformers/shared/constants"
	"github.com/makerdao/vulcanizedb/libraries/shared/factories/event"
	"github.com/makerdao/vulcanizedb/pkg/core"
	"github.com/makerdao/vulcanizedb/pkg/datastore/postgres"
	"github.com/makerdao/vulcanizedb/pkg/eth"
)

type Converter struct{}

func (Converter) toEntities(contractAbi string, logs []core.HeaderSyncLog) ([]NewCdpEntity, error) {
	var entities []NewCdpEntity
	for _, log := range logs {
		var entity NewCdpEntity
		address := log.Log.Address
		abi, err := eth.ParseAbi(contractAbi)
		if err != nil {
			return nil, err
		}
		contract := bind.NewBoundContract(address, abi, nil, nil, nil)

		err = contract.UnpackLog(&entity, "NewCdp", log.Log)
		if err != nil {
			return nil, err
		}

		entity.LogID = log.ID
		entity.HeaderID = log.HeaderID

		entities = append(entities, entity)
	}

	return entities, nil
}

func (converter Converter) ToModels(abi string, logs []core.HeaderSyncLog, _ *postgres.DB) ([]event.InsertionModel, error) {
	entities, entityErr := converter.toEntities(abi, logs)
	if entityErr != nil {
		return nil, fmt.Errorf("NewCDP converter couldn't convert logs to entities: %v", entityErr)
	}

	var models []event.InsertionModel
	for _, newCdpEntity := range entities {
		model := event.InsertionModel{
			SchemaName: constants.MakerSchema,
			TableName:  constants.NewCdpTable,
			OrderedColumns: []event.ColumnName{
				event.HeaderFK, event.LogFK, constants.UsrColumn, constants.OwnColumn, constants.CdpColumn,
			},
			ColumnValues: event.ColumnValues{
				event.HeaderFK:      newCdpEntity.HeaderID,
				event.LogFK:         newCdpEntity.LogID,
				constants.UsrColumn: newCdpEntity.Usr.Hex(),
				constants.OwnColumn: newCdpEntity.Own.Hex(),
				constants.CdpColumn: shared.BigIntToString(newCdpEntity.Cdp),
			},
		}
		models = append(models, model)
	}
	return models, nil
}
