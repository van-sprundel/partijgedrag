etl can be ran using a crontab
```bash
0 2 * * * [path]/etl >> /var/log/etl.log 2>&1
```

## TODO
- db storage
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

type DatabaseStorage struct {
    db *sql.DB
}

func (ds *DatabaseStorage) Store(category string, data []byte) error {
    _, err := ds.db.Exec(
        "", //query to store data
        category, data,
    )
    return err
}
```
