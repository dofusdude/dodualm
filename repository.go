package main

import (
	"context"
	"database/sql"
	"os"
	"path"
	"sync"

	"github.com/charmbracelet/log"
	_ "github.com/mattn/go-sqlite3"
)

var repositoryMutex = sync.Mutex{}
var DatabaseName = "almanax.db"

type Repository struct {
	Db  *sql.DB
	ctx context.Context
}

func NewDatabaseRepository(ctx context.Context, workdir string) *Repository {
	repo := Repository{}
	repo.Init(ctx, workdir)
	return &repo
}

func (r *Repository) Init(ctx context.Context, workdir string) error {
	dbpath := path.Join(workdir, DatabaseName)
	// check if the file exists
	_, err := os.Stat(dbpath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("Database does not exist, creating")
			file, err := os.Create(dbpath)
			if err != nil {
				return err
			}
			file.Close()
		} else {
			return err
		}
	} else {
		log.Info("Found database")
	}

	sqliteDatabase, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return err
	}
	r.Db = sqliteDatabase
	r.ctx = ctx

	return nil
}

func (r *Repository) Deinit() {
	r.Db.Close()
	r.Db = nil
}

func (r *Repository) GetAlmanaxByDateRangeAndNameID(from, to, nameID string) ([]MappedAlmanax, error) {
	query := `
		SELECT
			a.id, a.bonus_id, a.tribute_id, a.date, a.reward_kamas, a.created_at, a.updated_at, a.deleted_at,
			b.id, b.bonus_type_id, b.description_en, b.description_fr, b.description_es, b.description_de, b.description_it, b.description_pt,
			bt.id, bt.name_id, bt.name_en, bt.name_fr, bt.name_es, bt.name_de, bt.name_it, bt.name_pt,
			t.id, t.item_name_en, t.item_name_fr, t.item_name_es, t.item_name_de, t.item_name_it, t.item_name_pt,
			t.item_icon, t.item_sd, t.item_hq, t.item_hd, t.item_ankama_id, t.item_subtype, t.item_doduapi_uri, t.quantity
		FROM almanax AS a
		JOIN bonus AS b ON a.bonus_id = b.id
		JOIN bonus_types AS bt ON b.bonus_type_id = bt.id
		JOIN tribute AS t ON a.tribute_id = t.id
		WHERE a.date >= ? AND a.date <= ? AND bt.name_id = ? AND a.deleted_at IS NULL
		ORDER BY a.date ASC`

	rows, err := r.Db.Query(query, from, to, nameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MappedAlmanax

	for rows.Next() {
		var denorm MappedAlmanax
		var deletedAt sql.NullTime

		err := rows.Scan(
			&denorm.Almanax.ID, &denorm.Almanax.BonusID, &denorm.Almanax.TributeID, &denorm.Almanax.Date,
			&denorm.Almanax.RewardKamas, &denorm.Almanax.CreatedAt, &denorm.Almanax.UpdatedAt, &deletedAt,
			&denorm.Bonus.ID, &denorm.Bonus.BonusTypeID, &denorm.Bonus.DescriptionEn, &denorm.Bonus.DescriptionFr,
			&denorm.Bonus.DescriptionEs, &denorm.Bonus.DescriptionDe, &denorm.Bonus.DescriptionIt, &denorm.Bonus.DescriptionPt,
			&denorm.BonusType.ID, &denorm.BonusType.NameID, &denorm.BonusType.NameEn, &denorm.BonusType.NameFr,
			&denorm.BonusType.NameEs, &denorm.BonusType.NameDe, &denorm.BonusType.NameIt, &denorm.BonusType.NamePt,
			&denorm.Tribute.ID, &denorm.Tribute.ItemNameEn, &denorm.Tribute.ItemNameFr, &denorm.Tribute.ItemNameEs, &denorm.Tribute.ItemNameDe, &denorm.Tribute.ItemNameIt, &denorm.Tribute.ItemNamePt,
			&denorm.Tribute.ItemIcon, &denorm.Tribute.ItemSd, &denorm.Tribute.ItemHq, &denorm.Tribute.ItemHd,
			&denorm.Tribute.ItemAnkamaID, &denorm.Tribute.ItemSubtype,
			&denorm.Tribute.ItemDoduapiUri, &denorm.Tribute.Quantity)

		if err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			denorm.Almanax.DeletedAt = &deletedAt.Time
		}

		result = append(result, denorm)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetAlmanaxByDateRange(from, to string) ([]MappedAlmanax, error) {
	query := `
		SELECT
			a.id, a.bonus_id, a.tribute_id, a.date, a.reward_kamas, a.created_at, a.updated_at, a.deleted_at,
			b.id, b.bonus_type_id, b.description_en, b.description_fr, b.description_es, b.description_de, b.description_it, b.description_pt,
			bt.id, bt.name_id, bt.name_en, bt.name_fr, bt.name_es, bt.name_de, bt.name_it, bt.name_pt,
			t.id, t.item_name_en, t.item_name_fr, t.item_name_es, t.item_name_de, t.item_name_it, t.item_name_pt,
			t.item_icon, t.item_sd, t.item_hq, t.item_hd, t.item_ankama_id, t.item_subtype, t.item_doduapi_uri, t.quantity
		FROM almanax AS a
		JOIN bonus AS b ON a.bonus_id = b.id
		JOIN bonus_types AS bt ON b.bonus_type_id = bt.id
		JOIN tribute AS t ON a.tribute_id = t.id
		WHERE a.date >= ? AND a.date <= ? AND a.deleted_at IS NULL
		ORDER BY a.date ASC`

	rows, err := r.Db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MappedAlmanax

	for rows.Next() {
		var denorm MappedAlmanax
		var deletedAt sql.NullTime

		err := rows.Scan(
			&denorm.Almanax.ID, &denorm.Almanax.BonusID, &denorm.Almanax.TributeID, &denorm.Almanax.Date,
			&denorm.Almanax.RewardKamas, &denorm.Almanax.CreatedAt, &denorm.Almanax.UpdatedAt, &deletedAt,
			&denorm.Bonus.ID, &denorm.Bonus.BonusTypeID, &denorm.Bonus.DescriptionEn, &denorm.Bonus.DescriptionFr,
			&denorm.Bonus.DescriptionEs, &denorm.Bonus.DescriptionDe, &denorm.Bonus.DescriptionIt, &denorm.Bonus.DescriptionPt,
			&denorm.BonusType.ID, &denorm.BonusType.NameID, &denorm.BonusType.NameEn, &denorm.BonusType.NameFr,
			&denorm.BonusType.NameEs, &denorm.BonusType.NameDe, &denorm.BonusType.NameIt, &denorm.BonusType.NamePt,
			&denorm.Tribute.ID, &denorm.Tribute.ItemNameEn, &denorm.Tribute.ItemNameFr, &denorm.Tribute.ItemNameEs, &denorm.Tribute.ItemNameDe, &denorm.Tribute.ItemNameIt, &denorm.Tribute.ItemNamePt,
			&denorm.Tribute.ItemIcon, &denorm.Tribute.ItemSd, &denorm.Tribute.ItemHq, &denorm.Tribute.ItemHd,
			&denorm.Tribute.ItemAnkamaID, &denorm.Tribute.ItemSubtype,
			&denorm.Tribute.ItemDoduapiUri, &denorm.Tribute.Quantity)

		if err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			denorm.Almanax.DeletedAt = &deletedAt.Time
		}

		result = append(result, denorm)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) Create(almanax *Almanax) (int64, error) {
	query := `
		INSERT INTO almanax (bonus_id, tribute_id, date, reward_kamas, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`
	result, err := r.Db.Exec(query, almanax.BonusID, almanax.TributeID, almanax.Date, almanax.RewardKamas)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateAlmanax(almanax *Almanax) error {
	query := `
		UPDATE almanax
		SET bonus_id = ?, tribute_id = ?, date = ?, reward_kamas = ?, updated_at = datetime('now')
		WHERE id = ?`
	_, err := r.Db.Exec(query, almanax.BonusID, almanax.TributeID, almanax.Date, almanax.RewardKamas, almanax.ID)
	return err
}

func (r *Repository) CreateBonusType(bonusType *BonusType) (int64, error) {
	query := `INSERT INTO bonus_types (name_id, name_en, name_fr, name_es, name_de, name_it, name_pt, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`
	result, err := r.Db.Exec(query, bonusType.NameID, bonusType.NameEn, bonusType.NameFr, bonusType.NameEs,
		bonusType.NameDe, bonusType.NameIt, bonusType.NamePt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
