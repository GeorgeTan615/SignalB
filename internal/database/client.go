package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Database interface {
	InsertTicker(ctx context.Context, symbol, class string) error
	GetTickers(ctx context.Context) ([]Ticker, error)
	GetTickerClassBySymbol(ctx context.Context, tickerSymbol string) (string, error)
	GetTickersByTimeframe(ctx context.Context, timeframe string) ([]*Ticker, error)
	IsTickerRegistered(ctx context.Context, tickerSymbol string) bool

	InsertBinding(ctx context.Context, tickerSymbol, timeframe, strategy string) error
	GetBindingsByTicker(ctx context.Context, tickerSymbol string) ([]Binding, error)
	GetBindingsByTimeframe(ctx context.Context, timeframe string) ([]Binding, error)

	DeletePriceData(ctx context.Context, tickerSymbol, timeframe string, limit int) error
	InsertPriceData(ctx context.Context, timeframe string, data []PriceData) error
	GetPriceByTicker(ctx context.Context, tickerSymbol, timeframe string) ([]float64, error)

	Close()
	Ping(ctx context.Context) error
}

type DBClient struct {
	DB *sql.DB
}

func newDBClient(db *sql.DB) *DBClient {
	return &DBClient{
		DB: db,
	}
}

func (d *DBClient) InsertTicker(ctx context.Context, symbol, class string) error {
	query := `insert into ticker (symbol, class) values (?,?)`

	_, err := d.DB.ExecContext(ctx, query, symbol, class)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBClient) GetTickers(ctx context.Context) ([]Ticker, error) {
	query := `select symbol, class from ticker`

	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var tickers []Ticker
	for rows.Next() {
		var ticker Ticker

		err = rows.Scan(&ticker.Symbol, &ticker.Class)
		if err != nil {
			return nil, err
		}

		tickers = append(tickers, ticker)
	}

	return tickers, err
}

func (d *DBClient) GetTickerClassBySymbol(ctx context.Context, tickerSymbol string) (string, error) {
	query := fmt.Sprintf(`select class 
					from ticker
					where symbol = '%s'`, tickerSymbol)

	var class string

	err := d.DB.QueryRowContext(ctx, query).Scan(&class)
	if err != nil {
		return "", err
	}

	return class, nil
}

func (d *DBClient) IsTickerRegistered(ctx context.Context, tickerSymbol string) bool {
	checkQuery := `select count(symbol) 
							from ticker 
							where symbol = ?`

	var count int
	err := d.DB.QueryRowContext(ctx, checkQuery, tickerSymbol).Scan(&count)

	return err == nil && count > 0
}

func (d *DBClient) GetBindingsByTicker(ctx context.Context, tickerSymbol string) ([]Binding, error) {
	query := fmt.Sprintf(`select ticker_symbol, timeframe, strategy 
								from binding 
								where ticker_symbol = '%s'`, tickerSymbol)

	return d.getBindingsWithQuery(ctx, query)
}

func (d *DBClient) GetBindingsByTimeframe(ctx context.Context, timeframe string) ([]Binding, error) {
	query := fmt.Sprintf(`select ticker_symbol, timeframe, strategy 
								from binding 
								where timeframe = '%s'`, timeframe)

	return d.getBindingsWithQuery(ctx, query)
}

func (d *DBClient) getBindingsWithQuery(ctx context.Context, query string) ([]Binding, error) {
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var results []Binding
	for rows.Next() {
		var binding Binding

		if err := rows.Scan(&binding.TickerSymbol, &binding.Timeframe, &binding.Strategy); err != nil {
			return nil, err
		}
		results = append(results, binding)
	}

	return results, nil
}

func (d *DBClient) InsertBinding(ctx context.Context, tickerSymbol, timeframe, strategy string) error {
	registerQuery := `insert into binding (ticker_symbol, timeframe, strategy) values (?,?,?)`
	_, err := d.DB.ExecContext(ctx, registerQuery, tickerSymbol, timeframe, strategy)
	return err
}

func (d *DBClient) DeletePriceData(ctx context.Context, tickerSymbol, timeframe string, limit int) error {
	table := "price_" + strings.ToLower(timeframe)

	delQuery := fmt.Sprintf(`delete 
						from %s
						where (ticker_symbol,time) in (
							select ticker_symbol, time
							from %s
							where ticker_symbol = ?
							order by time 
							limit ?)`,
		table, table)

	_, err := d.DB.ExecContext(ctx, delQuery, tickerSymbol, limit)
	return err
}

func (d *DBClient) InsertPriceData(ctx context.Context, timeframe string, data []PriceData) error {
	table := "price_" + strings.ToLower(timeframe)
	insQuery := `insert into %s (ticker_symbol,time,price) values ('%s','%s',%.2f);`

	var builder strings.Builder

	for i := len(data) - 1; i > -1; i-- {
		currData := data[i]
		timeString := currData.Time.Format("2006-01-02 15:04:05")
		nxtQuery := fmt.Sprintf(insQuery, table, currData.TickerSymbol, timeString, currData.Price)
		builder.WriteString(nxtQuery)
	}

	finalQuery := builder.String()
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(finalQuery)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (d *DBClient) GetTickersByTimeframe(ctx context.Context, timeframe string) ([]*Ticker, error) {
	query := `select distinct t.symbol, t.class
					from ticker t join binding b on t.symbol = b.ticker_symbol
					where b.timeframe = ?`

	rows, err := d.DB.QueryContext(ctx, query, timeframe)
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var tickers []*Ticker
	for rows.Next() {
		var ticker Ticker

		err = rows.Scan(&ticker.Symbol, &ticker.Class)
		if err != nil {
			return nil, err
		}

		tickers = append(tickers, &ticker)
	}

	return tickers, nil
}

func (d *DBClient) GetPriceByTicker(ctx context.Context, tickerSymbol, timeframe string) ([]float64, error) {
	query := fmt.Sprintf(`select price
							from price_%s
							where ticker_symbol = ?
							order by time`, strings.ToLower(timeframe))

	rows, err := d.DB.QueryContext(ctx, query, tickerSymbol)
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var prices []float64

	for rows.Next() {
		var price float64

		err := rows.Scan(&price)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}

	return prices, nil
}

func (d *DBClient) Close() {
	d.DB.Close()
}

func (d *DBClient) Ping(ctx context.Context) error {
	query := `select symbol from ticker limit 1`

	//nolint:rowserrcheck
	res, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	defer res.Close()
	return nil
}
