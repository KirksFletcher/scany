package pgxscan

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/KirksFletcher/scany/dbscan"
)

// Querier is something that pgxscan can query and get the pgx.Rows from.
// For example, it can be: *pgxpool.Pool, *pgx.Conn or pgx.Tx.
type Querier interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
}

var (
	_ Querier = &pgxpool.Conn{}
	_ Querier = &pgxpool.Pool{}
	_ Querier = &pgx.Conn{}
	_ Querier = pgx.Tx(nil)
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")


func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake  = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// Select is a high-level function that queries rows from Querier and calls the ScanAll function.
// See ScanAll for details.
func Select(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "scany: query multiple result rows")
	}
	err = ScanAll(dst, rows)
	return errors.WithStack(err)
}

// Basic Insert function to allow for inserting structs
func Insert(ctx context.Context, db Querier, data interface{}, table string, additionalQuery string) error {

	fields := reflect.TypeOf(data)
	values := reflect.ValueOf(data)

	num := fields.NumField()
	var dbCols []string
	var dbVals []string

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)

		val, exists := field.Tag.Lookup("pgx")

		if exists {
			dbCols = append(dbCols, val)
		}else{
			dbCols = append(dbCols, toSnakeCase(field.Name))
		}

		var v string

		switch value.Kind() {
		case reflect.String:
			v = "'" + value.String() + "'"
		case reflect.Int:
			v = strconv.FormatInt(value.Int(), 10)
		case reflect.Int8:
			v = strconv.FormatInt(value.Int(), 10)
		case reflect.Int32:
			v = strconv.FormatInt(value.Int(), 10)
		case reflect.Int64:
			v = strconv.FormatInt(value.Int(), 10)
		case reflect.Float64:
			v = fmt.Sprintf("%f", value.Float())
		case reflect.Float32:
			v = fmt.Sprintf("%f", value.Float())
		default:
			return errors.Wrap(errors.New("type: " + value.Kind().String() + " unsupported"), "scany: this type not yet supported")
		}

		dbVals = append(dbVals, v)

	}

	sql := "INSERT INTO " + table + " (" + strings.Join(dbCols, ", ") + ") VALUES (" + strings.Join(dbVals, ", ") + ") " + additionalQuery

	_, err := db.Query(ctx, sql)
	if err != nil {
		return errors.Wrap(err, "scany: insertion error")
	}

	return errors.WithStack(err)
}

// Get is a high-level function that queries rows from Querier and calls the ScanOne function.
// See ScanOne for details.
func Get(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "scany: query one result row")
	}
	err = ScanOne(dst, rows)
	return errors.WithStack(err)
}

// ScanAll is a wrapper around the dbscan.ScanAll function.
// See dbscan.ScanAll for details.
func ScanAll(dst interface{}, rows pgx.Rows) error {
	err := dbscan.ScanAll(dst, NewRowsAdapter(rows))
	return errors.WithStack(err)
}

// ScanOne is a wrapper around the dbscan.ScanOne function.
// See dbscan.ScanOne for details. If no rows are found it
// returns a pgx.ErrNoRows error.
func ScanOne(dst interface{}, rows pgx.Rows) error {
	err := dbscan.ScanOne(dst, NewRowsAdapter(rows))
	if dbscan.NotFound(err) {
		return errors.WithStack(pgx.ErrNoRows)
	}
	return errors.WithStack(err)
}

// NotFound is a helper function to check if an error
// is `pgx.ErrNoRows`.
func NotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// RowScanner is a wrapper around the dbscan.RowScanner type.
// See dbscan.RowScanner for details.
type RowScanner struct {
	*dbscan.RowScanner
}

// NewRowScanner returns a new RowScanner instance.
func NewRowScanner(rows pgx.Rows) *RowScanner {
	ra := NewRowsAdapter(rows)
	return &RowScanner{RowScanner: dbscan.NewRowScanner(ra)}
}

// ScanRow is a wrapper around the dbscan.ScanRow function.
// See dbscan.ScanRow for details.
func ScanRow(dst interface{}, rows pgx.Rows) error {
	err := dbscan.ScanRow(dst, NewRowsAdapter(rows))
	return errors.WithStack(err)
}

// RowsAdapter makes pgx.Rows compliant with the dbscan.Rows interface.
// See dbscan.Rows for details.
type RowsAdapter struct {
	pgx.Rows
}

// NewRowsAdapter returns a new RowsAdapter instance.
func NewRowsAdapter(rows pgx.Rows) *RowsAdapter {
	return &RowsAdapter{Rows: rows}
}

// Columns implements the dbscan.Rows.Columns method.
func (ra RowsAdapter) Columns() ([]string, error) {
	columns := make([]string, len(ra.Rows.FieldDescriptions()))
	for i, fd := range ra.Rows.FieldDescriptions() {
		columns[i] = string(fd.Name)
	}
	return columns, nil
}

// Close implements the dbscan.Rows.Close method.
func (ra RowsAdapter) Close() error {
	ra.Rows.Close()
	return nil
}
