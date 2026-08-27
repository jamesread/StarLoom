package store

// CvarRow is a configuration variable row from the database.
type CvarRow struct {
	Key, MainType, Title, Description, Category, ValueString string
	Ordinal, ValueInt                                        int
}
