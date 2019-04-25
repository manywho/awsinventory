package data

import "github.com/itmecho/awsinventory/internal/inventory"

// MapperFunc takes an inventory row, performs some action, and returns an error
type MapperFunc func(inventory.Row) error

// MapRows takes a MapperFunc as an argument and runs it against each stored row
func (d *Data) MapRows(mapper MapperFunc) {
	for _, row := range d.rows {
		if err := mapper(row); err != nil {
			d.log.Error(err)
		}
	}
}
