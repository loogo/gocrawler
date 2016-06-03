package database

import (
	"database/sql"
	"log"
)

// MySQL Instance
type MySQL struct {
	DataSourceName string
}

func (mysql *MySQL) db() (*sql.DB, error) {
	db, err := sql.Open("mysql", mysql.DataSourceName)
	return db, err
}

// CreateDb create database and table
func (mysql *MySQL) CreateDb() {
	db, err := mysql.db()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
    create table hc_product(
        id INT(10) NOT NULL AUTO_INCREMENT, 
        product_id VARCHAR(64), 
        name VARCHAR(64), 
        spec VARCHAR(64), 
        img VARCHAR(64), 
        price VARCHAR(64),
        pri VARCHAR(64),
        img_id VARCHAR(64),
        PRIMARY KEY (id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci;
    `

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		sqlStmt = "delete from hc_product;"
		db.Exec(sqlStmt)
	}
}

// Insert new data
func (mysql *MySQL) Insert(args ...interface{}) {
	db, err := mysql.db()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into hc_product(name,img,price,spec,product_id,pri,img_id) values(?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

// func todo() {
// 	db, err := db()
// 	rows, err := db.Query("select id,name from foo")

// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var id int
// 		var name string
// 		err = rows.Scan(&id, &name)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println(id, name)
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	stmt, err = db.Prepare("select name from foo where id = ?")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close()
// 	var name string
// 	err = stmt.QueryRow("3").Scan(&name)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println(name)

// 	_, err = db.Exec("delete from foo")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	_, err = db.Exec("insert into foo(id,name) values(1,'foo'),(2,'bar'),(3,'baz')")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rows, err = db.Query("select id,name from foo")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var id int
// 		var name string
// 		err = rows.Scan(&id, &name)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println(id, name)
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
