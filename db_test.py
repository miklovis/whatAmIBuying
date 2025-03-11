import sqlite3

con = sqlite3.connect("test_database.db")
cur = con.cursor()

try:
    cur = con.execute("CREATE TABLE Purchases(id INTEGER PRIMARY KEY, name VARCHAR(128), price FLOAT, datetime DATETIME)")
except sqlite3.OperationalError:
    pass

res = cur.execute("SELECT name FROM sqlite_master")

print(res.fetchone())