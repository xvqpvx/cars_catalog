package repository

import (
	"car_catalog/internal/dto"
	"car_catalog/internal/model"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CarRepositoryImpl struct {
	conn *pgxpool.Pool
}

func NewCarRepository(conn *pgxpool.Pool) CarRepository {
	return &CarRepositoryImpl{
		conn: conn,
	}
}

func (c *CarRepositoryImpl) AddCars(ctx context.Context, cars []model.Car) error {
	entries := [][]any{}
	columns := []string{
		"mark", "model", "year", "reg_num",
		"owner_name", "owner_surname", "owner_patronymic",
	}
	tableName := "car"

	for _, car := range cars {
		entries = append(entries, []any{
			car.Mark, car.Model, car.Year, car.RegNum,
			car.OwnerName, car.OwnerSurname, car.OwnerPatronymic,
		})
	}

	tx, err := c.conn.Begin(ctx)
	if err != nil {
		log.Printf("[ERROR] Repo - AddCars - Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	copyCount, err := c.conn.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(entries),
	)
	if err != nil {
		return fmt.Errorf("[ERROR] Repo - AddCars - error copying into %s table: %w", tableName, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("[ERROR] Repo - AddCars - Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction")
	}
	log.Printf("[INFO] Repo - AddCars - Transaction committed successfully")

	log.Printf("[INFO] Repo - AddCars - New cars recorded, %d rows inserted \n", copyCount)
	return nil
}

func (c *CarRepositoryImpl) DeleteCar(ctx context.Context, carId int) error {
	query := `DELETE FROM car
	WHERE id = $1`

	tx, err := c.conn.Begin(ctx)
	if err != nil {
		log.Printf("[ERROR] Repo - DeleteCar - Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	commandTag, err := c.conn.Exec(ctx, query, carId)
	if err != nil {
		log.Printf("[ERROR] Repo - DeleteCar - Error executing delete query: %v", err)
		return err
	}
	if commandTag.RowsAffected() <= 0 {
		err := fmt.Errorf("no rows found to update")
		log.Printf("[ERROR] Repo - DeleteCar - Error executing delete query: %v", err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("[ERROR] Repo - AddCars - Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction")
	}
	log.Println("[INFO] Repo - DeleteCar - Transaction committed successfully")

	log.Println("[INFO] Repo - DeleteCar - Car deleted successfuly")
	return nil
}

func (c *CarRepositoryImpl) GetCarById(ctx context.Context, carId int) (model.Car, error) {
	query := `SELECT mark, model, year, reg_num
	FROM car
	WHERE id = $1`

	var car model.Car
	err := c.conn.QueryRow(ctx, query, carId).Scan(&car.Mark, &car.Model, &car.Year, &car.RegNum)
	if err != nil {
		log.Printf("[ERROR] Repo - GetCarById - Error executing select query: %v", err)
		return model.Car{}, err
	}
	log.Printf("[DEBUG] Repo - GetCarById - Сar details: %+v", car)
	car.CarId = carId

	log.Println("[INFO] Repo - GetCarById - Car founded successfuly")
	return car, nil
}

func (c *CarRepositoryImpl) GetCars(ctx context.Context, limit int, mark, carModel, year string, cursors dto.Cursors) ([]model.Car, dto.Cursors, error) {

	if limit == 0 {
		return []model.Car{}, dto.Cursors{}, errors.New("limit cannot be zero")
	}
	if cursors.Next != "" && cursors.Prev != "" {
		return []model.Car{}, dto.Cursors{}, errors.New("two cursors cannot be provided at the same time")
	}

	values := make([]interface{}, 0, 8)
	rowsLeftQuery := "SELECT COUNT(*) FROM car c"
	pagination := ""

	if cursors.Next != "" {
		f := fmt.Sprintf("c.id > $%d", len(values)+1)
		rowsLeftQuery += fmt.Sprintf(" WHERE %s", f)
		pagination += fmt.Sprintf(
			"WHERE %s AND ($%d = '' OR mark = $%d) AND ($%d = '' OR model = $%d) AND ($%d = '' OR year = $%d::int) ORDER BY id ASC LIMIT $%d",
			f, len(values)+2, len(values)+2, len(values)+3, len(values)+3, len(values)+4, len(values)+4, len(values)+5)

		values = append(values, cursors.Next, mark, carModel, year, limit)
	}

	if cursors.Prev != "" {
		f := fmt.Sprintf("c.id < $%d", len(values)+1)
		rowsLeftQuery += fmt.Sprintf(" WHERE %s", f)
		pagination += fmt.Sprintf("WHERE %s AND ($%d = '' OR mark = $%d) AND ($%d = '' OR model = $%d) AND ($%d = '' OR year = $%d::int) ORDER BY id DESC LIMIT $%d",
			f, len(values)+2, len(values)+2, len(values)+3, len(values)+3, len(values)+4, len(values)+4, len(values)+5)
		values = append(values, cursors.Prev, mark, carModel, year, limit)
	}

	if cursors.Next == "" && cursors.Prev == "" {
		pagination = fmt.Sprintf(
			"WHERE ($%d = '' OR mark = $%d) AND ($%d = '' OR model = $%d) AND ($%d = '' OR year = $%d::int) ORDER BY c.id ASC LIMIT $%d",
			len(values)+1, len(values)+1, len(values)+2, len(values)+2, len(values)+3, len(values)+3, len(values)+4)
		values = append(values, mark, carModel, year, limit)
	}

	stmt := fmt.Sprintf(`
	WITH c AS (
		SELECT * FROM car c %s
	)
	SELECT id, mark, model, year, reg_num,
	(%s) AS rows_left,
	(SELECT COUNT(*) FROM car) AS total
	FROM c
	ORDER BY id ASC
	`, pagination, rowsLeftQuery)

	log.Printf("[DEBUG] Repo - GetCars - Pagination: %s", pagination)
	log.Printf("[DEBUG] Repo - GetCars - RowsLeftQuery: %s", rowsLeftQuery)
	log.Printf("[DEBUG] Repo - GetCars - Statement: %s", stmt)
	log.Printf("[DEBUG] Repo - GetCars - Values: %+v", values)

	rows, err := c.conn.Query(ctx, stmt, values...)
	if err != nil {
		log.Printf("[ERROR] Repo - GetCars - Error executing select query: %v", err)
		return []model.Car{}, dto.Cursors{}, err
	}
	defer rows.Close()

	var (
		rowsLeft int
		total    int
		cars     []model.Car
	)
	for rows.Next() {
		var car model.Car

		err := rows.Scan(&car.CarId, &car.Mark, &car.Model, &car.Year, &car.RegNum, &rowsLeft, &total)
		if err != nil {
			log.Printf("[ERROR] Repo - GetCars - Error scanning row: %v", err)
			return []model.Car{}, dto.Cursors{}, err
		}
		cars = append(cars, car)
	}
	var (
		prevCursor string
		nextCursor string
	)

	switch {
	case rowsLeft < 0:
	case cursors.Prev == "" && cursors.Next == "":
		nextCursor = fmt.Sprint((cars[len(cars)-1].CarId))

	case cursors.Next != "" && rowsLeft == len(cars):
		prevCursor = fmt.Sprint(cars[0].CarId)

	case cursors.Prev != "" && rowsLeft == len(cars):
		nextCursor = fmt.Sprint(cars[len(cars)-1].CarId)

	case cursors.Prev != "" && total == rowsLeft:
		prevCursor = fmt.Sprint(cars[0].CarId)

	default:
		nextCursor = fmt.Sprint(cars[len(cars)-1].CarId)
		prevCursor = fmt.Sprint(cars[0].CarId)
	}
	log.Printf("[INFO] Repo - GetCars - Got %d records from the database", len(cars))

	return cars, dto.Cursors{Prev: prevCursor, Next: nextCursor}, nil
}

func (c *CarRepositoryImpl) UpdateCar(ctx context.Context, car model.Car) error {
	query := `UPDATE car
	SET mark = $1, model = $2, year = $3, reg_num = $4
	WHERE id = $5`

	tx, err := c.conn.Begin(ctx)
	if err != nil {
		log.Printf("[ERROR] Repo - UpdateCar - Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	commandTag, err := c.conn.Exec(ctx, query, car.Mark, car.Model, car.Year, car.RegNum, car.CarId)
	if err != nil {
		log.Printf("[ERROR] Repo - UpdateCar - Error executing delete query: %v", err)
		return err
	}
	if commandTag.RowsAffected() <= 0 {
		err := fmt.Errorf("no rows found to update")
		log.Printf("[ERROR] Repo - UpdateCar - Error executing delete query: %v", err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("[ERROR] Repo - UpdateCar - Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction")
	}
	log.Printf("[INFO] Repo - UpdateCar - Transaction committed successfully")

	log.Printf("[INFO] Repo - UpdateCar - Car updated successfuly")
	return nil
}
