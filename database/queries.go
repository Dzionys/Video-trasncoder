package db

import (
	"fmt"
	"log"
)

func getInsertQuery(clms []string, tname string) string {
	query := fmt.Sprintf("INSERT INTO %v (", tname)
	val := "("
	for i, c := range clms {
		if i != len(clms)-1 {
			query += fmt.Sprintf("%v,", c)
			val += "?,"
		} else {
			query += fmt.Sprintf("%v", c)
			val += "?"
		}
	}
	val += ")"
	query += fmt.Sprintf(") VALUES %v", val)

	return query
}

func getSelectQuery(clms []string, tname string, key string) string {
	query := "SELECT "
	if len(clms) > 0 {
		for i, c := range clms {
			if i != len(clms)-1 {
				query += fmt.Sprintf("%v,", c)
			} else {
				query += fmt.Sprintf("%v ", c)
			}
		}
	} else {
		query += "* "
	}
	query += "FROM " + tname
	if key != "" {
		query += " WHERE " + key
	}

	return query
}

func getDeleteQuery(tname string, key string) string {
	query := fmt.Sprintf("DELETE FROM %v WHERE %v", tname, key)
	return query
}

func getUpdateQuery(clms []string, tname string, key string) string {
	query := fmt.Sprintf("UPDATE %v SET ", tname)
	for i, c := range clms {
		if i != len(clms)-1 {
			query += fmt.Sprintf("%v,", c)
		} else {
			query += fmt.Sprintf("%v ", c)
		}
	}
	query += "WHERE " + key

	return query
}

func getIdByName(tname string, name string) (int, error) {
	var id = -1
	row, err := DB.Query(fmt.Sprintf("SELECT Id FROM %v WHERE Name='%v'", tname, name))
	if err != nil {
		return id, err
	}
	for row.Next() {
		err = row.Scan(&id)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func prepareTable(table string) error {
	statement, err := DB.Prepare(table)
	if err != nil {
		log.Println("Error: failed to prepare database table")
		log.Println(err)
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

func runCustomQuery(query string) error {
	statement, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}
