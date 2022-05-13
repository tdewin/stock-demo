package main

//all made to work even if there is no db

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func initdb(ctx *context.Context, pool *pgxpool.Pool, insertData string) error {
	var rerr error
	fmt.Println("Creating table if not exists")
	_, cerr := pool.Exec(*ctx, createTable)

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
				_, cerr := pool.Exec(context.Background(), insertData)
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
	fmt.Println("v 1.1")

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

	//use ?adminkey=unlock to do editing
	adminkey := "unlock"
	testenvadmin := os.Getenv("ADMINKEY")
	if testenvadmin != "" {
		adminkey = testenvadmin
	}

	//data that is insert when calling init
	insertData := defaultInsertData
	content, err := ioutil.ReadFile("/var/stockdb/initinsert.psql")
	if err == nil {
		insertData = string(content)
	} else {
		fmt.Println("Loaded default insert query")
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
	fmt.Printf("DB string postgres://%s:***@%s:%s/%s\n", username, server, port, dbname)

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

	//when first deploying, run this to create the table
	http.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err = initdb(&ctx, pool, insertData)
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
	//admin function
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
			defer br.Close()
			msg := NewMessage("Admin set executed")
			msg.Refresh = 2
			terr = tmpl.Execute(w, msg)
			if terr != nil {
				fmt.Fprint(w, "Internal Error, could not render")
			}
		}
	})
	//buying items, lowering the stock
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
					defer br.Close()
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
	//main function
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

		rows, err := pool.Query(ctx, `SELECT id, product, unit, amount, price  FROM stock ORDER BY id`)

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
				var product, unit string
				var amount, price float64
				err := rows.Scan(&id, &product, &unit, &amount, &price)
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
