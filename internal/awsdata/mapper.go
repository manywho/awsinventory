package awsdata

import "github.com/manywho/awsinventory/internal/inventory"

// MapperFunc takes an inventory row, performs some action, and returns an error
type MapperFunc func(inventory.Row) error

// MapRows takes a MapperFunc as an argument and runs it against each row
func (d *AWSData) MapRows(mapper MapperFunc) {
	for row := range d.rows {
		if err := mapper(row); err != nil {
			d.log.Error(err)
		}
	}
}
