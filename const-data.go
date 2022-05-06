package main

const defaultInsertData = `INSERT INTO stock(product,unit,amount,price) VALUES ('Apples','KG',1.0,3.5);
INSERT INTO stock(product,unit,amount,price) VALUES ('Bananas','KG',0.0,5.0);
INSERT INTO stock(product,unit,amount,price) VALUES ('Leek','KG',100.0,2.0);
INSERT INTO stock(product,unit,amount,price) VALUES ('oPhone 17','Piece(s)',5.0,1500.0);
INSERT INTO stock(product,unit,amount,price) VALUES ('OneDivide 18 Pro','Piece(s)',5.0,1000.0);
INSERT INTO stock(product,unit,amount,price) VALUES ('Paystation','Piece(s)',10.0,400.0);
INSERT INTO stock(product,unit,amount,price) VALUES ('Tony TV','Piece(s)',10.0,699.0);
INSERT INTO stock(product,unit,amount,price) VALUES ('LB TV','Piece(s)',5.0,999.0);
`
const createTable = `CREATE TABLE IF NOT EXISTS stock (
		id serial PRIMARY KEY,
		product VARCHAR ( 50 ) NOT NULL,
		unit VARCHAR ( 55 ) NOT NULL,
		amount decimal ( 10,2 ),
		price decimal ( 10,2 )
);	
`
