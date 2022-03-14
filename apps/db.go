package apps

import (
	"fmt"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/lnbits/lnbits/utils"
	"github.com/lucsky/cuid"
	"gorm.io/gorm/clause"
)

func DBGet(wallet, app, model, key string) (map[string]interface{}, error) {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Model:    model,
		Key:      key,
	}

	result := storage.DB.First(&item)
	if result.Error != nil {
		return nil, result.Error
	}

	if err := fillComputedValues(item); err != nil {
		return item.Value, fmt.Errorf("failed to compute: %w", err)
	}

	return item.Value, nil
}

func DBList(wallet, app, model, startkey, endkey string) ([]models.AppDataItem, error) {
	q := storage.DB.
		Where(&models.AppDataItem{WalletID: wallet, App: app, Model: model})

	if startkey != "" {
		q = q.Where("key > ?", startkey)
	}
	if endkey != "" {
		q = q.Where("key < ?", endkey)
	}

	var items []models.AppDataItem
	result := q.Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, item := range items {
		if err := fillComputedValues(item); err != nil {
			return items, fmt.Errorf("failed to compute: %w", err)
		}
	}

	return items, nil
}

func DBSet(wallet, app, model, key string, value map[string]interface{}) error {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Model:    model,
		Key:      key,
		Value:    value,
	}

	settings, err := GetAppSettings(app)
	if err != nil {
		return fmt.Errorf("failed to get app on model.set: %w", err)
	}
	if err := settings.getModel(model).validateItem(item); err != nil {
		j, _ := utils.JSONMarshal(value)
		return fmt.Errorf("invalid value %s for model %s: %w", string(j), model, err)
	}

	result := storage.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "app"}, {Name: "wallet_id"}, {Name: "model"}, {Name: "key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&item)

	if result.Error != nil {
		return result.Error
	}

	SendItemSSE(item)
	return nil
}

func DBAdd(wallet, app, model string, value map[string]interface{}) (string, error) {
	key := cuid.Slug()
	err := DBSet(wallet, app, model, key, value)
	if err != nil {
		return "", err
	}
	return key, nil
}

func DBUpdate(wallet, app, model, key string, updates map[string]interface{}) error {
	value, err := DBGet(wallet, app, model, key)
	if err != nil {
		return fmt.Errorf("failed to get %s: %w", key, err)
	}

	for k, v := range updates {
		value[k] = v
	}

	return DBSet(wallet, app, model, key, value)
}

func DBDelete(wallet, app, model, key string) error {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Model:    model,
		Key:      key,
	}
	result := storage.DB.Delete(&models.AppDataItem{}, item)

	if result.Error != nil {
		return result.Error
	}

	// an item with an empty .Value means it was deleted
	SendItemSSE(item)

	return nil
}
