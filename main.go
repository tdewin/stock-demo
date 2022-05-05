package main

//all made to work even if there is no db

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func initdb(ctx *context.Context, pool *pgxpool.Pool) error {
	var rerr error
	fmt.Println("Creating table if not exists")
	_, cerr := pool.Exec(*ctx, `
CREATE TABLE IF NOT EXISTS stock (
		id serial PRIMARY KEY,
		product VARCHAR ( 50 ) UNIQUE NOT NULL,
		department VARCHAR ( 50 ) NOT NULL,
		unit VARCHAR ( 55 ) NOT NULL,
		amount decimal ( 10,2 ),
		price decimal ( 10,2 )
);	
	`)

	if cerr != nil {
		rerr = fmt.Errorf("db query failed: %v", cerr)
	} else {
		var count int
		cerr = pool.QueryRow(*ctx, `SELECT count(*) FROM stock;	`).Scan(&count)
		if cerr != nil {
			rerr = fmt.Errorf("select count failed: %v", cerr)
		} else {
			fmt.Println(count, "rows in table")
			if count == 0 {
				_, cerr := pool.Exec(context.Background(), `
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('Apples','Fruits','KG',1.0,3.5);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('Bananas','Fruits','KG',0.0,5.0);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('Leek','Vegetables','KG',100.0,2.0);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('oPhone 17','Electronics','Piece(s)',5.0,1500.0);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('OneDivide 18 Pro','Electronics','Piece(s)',5.0,1000.0);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('Paystation','Electronics','Piece(s)',10.0,400.0);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('Tony TV','Electronics','Piece(s)',10.0,699.0);
	INSERT INTO stock(product,department,unit,amount,price) VALUES ('LB TV','Electronics','Piece(s)',5.0,999.0);
				`)
				if cerr != nil {
					rerr = fmt.Errorf("insert random data query failed: %v", cerr)
				} else {
					fmt.Println("Inserted random data")
				}
			}
		}
	}

	return rerr
}

func main() {
	fmt.Println("v 1.0")

	flag.Parse()
	/*
		docker container run --name stockdb -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=notsecure -e POSTGRES_DB=stock -d postgres
		set POSTGRES_SERVER=localhost
		set POSTGRES_USER=root
		set POSTGRES_PASSWORD=notsecure
		set POSTGRES_DB=stock
		set POSTGRES_PORT=5432
		docker container run --name stockfrontend -p 14000:8080 -e POSTGRES_SERVER=stockdb -e POSTGRES_USER=root -e POSTGRES_PASSWORD=notsecure -e POSTGRES_PORT=5432 -e POSTGRES_DB=stock -d tdewin/stock-demo
	*/

	adminkey := "unlock"
	testenvadmin := os.Getenv("ADMINKEY")
	if testenvadmin != "" {
		adminkey = testenvadmin
	}

	server := os.Getenv("POSTGRES_SERVER")
	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	dburl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, server, port, dbname)
	fmt.Printf("db string postgres://%s:***@%s:%s/%s\n", username, server, port, dbname)
	config, err := pgxpool.ParseConfig(dburl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse config: %v\n", err)

	}
	//lazy loading so can start with errors
	config.LazyConnect = true

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to lazy load db: %v\n", err)
	}
	defer pool.Close()

	tmpl, terr := template.New("msg").Parse(msghtml)
	//panic should never happen but okay
	if terr != nil {
		panic(terr)
	}

	tmplproduct, terr := template.New("productpage").Parse(mainhtml)
	//panic should never happen but okay
	if terr != nil {
		panic(terr)
	}

	http.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err = initdb(&ctx, pool)
		if err == nil {
			msg := NewMessage("Init OK")

			terr = tmpl.Execute(w, msg)
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			terr = tmpl.Execute(w, NewError(err.Error()))
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		}
	})
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		r.ParseForm()

		transactions := &pgx.Batch{}
		deleteids := []string{}

		for k, v := range r.PostForm {
			test := strings.Split(k, "-")
			if len(test) == 2 && len(v) > 0 {
				if test[0] == "setstock" {
					transactions.Queue(`UPDATE stock SET amount = $2 WHERE id=$1;`, test[1], v[0])
				} else if test[0] == "setprice" {
					transactions.Queue(`UPDATE stock SET price = $2 WHERE id=$1;`, test[1], v[0])
				} else if test[0] == "setproduct" {
					transactions.Queue(`UPDATE stock SET product = $2 WHERE id=$1;`, test[1], v[0])
				} else if test[0] == "setunit" {
					transactions.Queue(`UPDATE stock SET unit = $2 WHERE id=$1;`, test[1], v[0])
				} else if test[0] == "setkeep" && v[0] == "delete" {
					deleteids = append(deleteids, test[1])
				}
			}
			//fmt.Println(k, v)
		}
		for _, delid := range deleteids {
			transactions.Queue(`DELETE FROM stock WHERE id=$1;`, delid)
		}
		br := pool.SendBatch(ctx, transactions)
		_, err := br.Exec()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			terr = tmpl.Execute(w, NewError(err.Error()))
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		} else {
			msg := NewMessage("Admin set executed")
			msg.Redirect = "../"
			msg.Refresh = 2
			terr = tmpl.Execute(w, msg)
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		}
	})
	http.HandleFunc("/buy", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		r.ParseForm()

		idindex := []string{}

		buymap := make(map[string]*BuyTransaction)

		for k, v := range r.PostForm {
			test := strings.Split(k, "-")
			if len(test) == 2 && test[0] == "qty" {
				buyid := test[1]
				qty, _ := strconv.ParseFloat(v[0], 64)
				if qty > 0 {
					idindex = append(idindex, buyid)
					buymap[buyid] = &BuyTransaction{qty, 0, 0, "", false, 0, 0}
				}
			}
		}

		rows, err := pool.Query(ctx, `SELECT id,product,unit,amount,price FROM stock WHERE id = ANY ($1);`, idindex)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			terr = tmpl.Execute(w, NewError(err.Error()))
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		} else {
			defer rows.Close()

			processedanything := false
			total := 0.0
			for rows.Next() {
				var id int
				var product, unit string
				var amount, price float64
				err := rows.Scan(&id, &product, &unit, &amount, &price)
				if err != nil {
					fmt.Println(err)
				} else {
					strid := fmt.Sprintf("%d", id)
					processedanything = true

					buymap[strid].Product = product
					buymap[strid].StockCalculated = true
					buymap[strid].Price = price

					if buymap[strid].Request > amount {
						buymap[strid].Bought = amount
						buymap[strid].Remaining = 0
					} else {
						buymap[strid].Bought = buymap[strid].Request
						buymap[strid].Remaining = amount - buymap[strid].Request
					}

					buymap[strid].Sum = (buymap[strid].Bought) * price
					total = total + buymap[strid].Sum
				}
			}

			if processedanything {
				transactions := &pgx.Batch{}
				for buyid, transaction := range buymap {
					if transaction.StockCalculated {
						transactions.Queue(`UPDATE stock SET amount = $2 WHERE id=$1;`, buyid, transaction.Remaining)
					}

				}
				br := pool.SendBatch(ctx, transactions)
				_, err := br.Exec()

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					terr = tmpl.Execute(w, NewError(err.Error()))
					if terr != nil {
						fmt.Fprint(w, "Internal Error, could not render")
					}
				} else {
					for k, transaction := range buymap {
						println(k, transaction.Bought, transaction.Product)
					}
					msg := NewMessage("Thanks for buying")
					msg.BuyTable = &buymap
					msg.BuySum = total
					terr = tmpl.Execute(w, msg)
					if terr != nil {
						fmt.Fprint(w, "Internal Error, could not render")
					}
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				errmsg := NewError("We were not able to find any of your products or you didnt buy anything")
				errmsg.Redirect = "./"
				errmsg.Refresh = 3
				terr = tmpl.Execute(w, errmsg)
				if terr != nil {
					fmt.Fprint(w, "Internal Error, could not render")
				}
			}

		}

	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		isAdmin := false
		r.ParseForm()
		testadmin := r.Form.Get("adminkey")
		if testadmin == adminkey {
			isAdmin = true
			fmt.Println("Admin mode unlocked")
		}

		rows, err := pool.Query(ctx, `SELECT id, product, department, unit, amount, price  FROM stock ORDER BY id`)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			terr = tmpl.Execute(w, NewError(err.Error()))
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		} else {
			defer rows.Close()
			allstock := []Stock{}

			for rows.Next() {
				var id int
				var product, department, unit string
				var amount, price float64
				err := rows.Scan(&id, &product, &department, &unit, &amount, &price)
				if err != nil {
					fmt.Println(err)
				}

				allstock = append(allstock, Stock{id, product, fmt.Sprintf("%.2f %s", amount, unit), amount, unit, price})
			}

			terr = tmplproduct.Execute(w, StockMessage{allstock, isAdmin})
			if terr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Internal Error, could not render")
			}
		}

	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
