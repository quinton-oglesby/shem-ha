import os
import mysql.connector

# MySQL database connection settings
db = mysql.connector.connect(
    host="localhost",
    user="root",
    password="KingArthur09052012?",
    database="discord_mimicry"
)

# Directory containing the text files
directory = './files/'

# iterate over files in the directory
for filename in os.listdir(directory):
    if filename.endswith(".txt"):
        # create table based on filename
        table_name = filename[:-4]  # remove the '.txt' extension
        cursor = db.cursor()
        cursor.execute(f"CREATE TABLE IF NOT EXISTS user_{table_name} (id INT AUTO_INCREMENT PRIMARY KEY, line TEXT)")
        
        # read lines from file and insert into table
        filepath = os.path.join(directory, filename)
        with open(filepath, "r") as file:
            lines = file.readlines()
            for line in lines:
                if line == '\n':
                    continue

                cursor.execute(f"INSERT INTO user_{table_name} (line) VALUES (%s)", (line,))        
        db.commit()
        print(f"Inserted {len(lines)} rows into {table_name} table")

# close database connection
db.close()
