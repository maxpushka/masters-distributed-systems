from testcontainers.postgres import PostgresContainer
from testcontainers.mongodb import MongoDbContainer
from testcontainers.neo4j import Neo4jContainer
from sqlalchemy import create_engine, text
from pymongo import MongoClient
from neo4j import GraphDatabase
import json


def print_result(query_name, result):
    print(f"Query: {query_name}")
    print(json.dumps(result, indent=4))


def postgres_driver():
    with PostgresContainer("postgres:16") as postgres:
        engine = create_engine(postgres.get_connection_url())
        with engine.connect() as connection:
            # Run migrations

            connection.execute(text("""
            CREATE TABLE cities (
                id SERIAL PRIMARY KEY,
                name VARCHAR NOT NULL
            );
            CREATE TABLE users (
                id SERIAL PRIMARY KEY,
                login VARCHAR NOT NULL,
                password VARCHAR NOT NULL,
                city_id INT REFERENCES cities(id)
            );
            CREATE TABLE resumes (
                id SERIAL PRIMARY KEY,
                user_id INT REFERENCES users(id),
                summary TEXT NOT NULL
            );
            CREATE TABLE hobbies (
                id SERIAL PRIMARY KEY,
                name VARCHAR NOT NULL
            );
            CREATE TABLE resume_hobbies (
                resume_id INT REFERENCES resumes(id),
                hobby_id INT REFERENCES hobbies(id)
            );
            CREATE TABLE experience (
                id SERIAL PRIMARY KEY,
                user_id INT REFERENCES users(id),
                institution VARCHAR NOT NULL
            );
            """))
            # Insert data

            connection.execute(text("INSERT INTO cities (name) VALUES ('Kyiv');"))
            connection.execute(text("INSERT INTO cities (name) VALUES ('Lviv');"))
            connection.execute(text("INSERT INTO users (login, password, city_id) VALUES ('user1', 'pass', 1);"))
            connection.execute(text("INSERT INTO users (login, password, city_id) VALUES ('user2', 'pass', 2);"))
            connection.execute(text("INSERT INTO resumes (user_id, summary) VALUES (1, 'Experienced software developer...');"))
            connection.execute(text("INSERT INTO resumes (user_id, summary) VALUES (2, 'Data scientist...');"))
            connection.execute(text("INSERT INTO hobbies (name) VALUES ('football'), ('reading'), ('traveling');"))
            connection.execute(text("INSERT INTO resume_hobbies (resume_id, hobby_id) VALUES (1, 1), (1, 2), (2, 3);"))
            connection.execute(text("INSERT INTO experience (user_id, institution) VALUES (1, 'Company A');"))
            connection.execute(text("INSERT INTO experience (user_id, institution) VALUES (2, 'Company B');"))
            connection.execute(text("INSERT INTO experience (user_id, institution) VALUES (2, 'Company A');"))

            # Perform queries

            # Get Resume
            result = connection.execute(text("SELECT * FROM resumes WHERE user_id = 1;")).mappings().all()
            print_result("Get Resume", [dict(row) for row in result])

            # Get All Hobbies
            result = connection.execute(text("SELECT DISTINCT h.name FROM hobbies h "
                                             "JOIN resume_hobbies rh ON h.id = rh.hobby_id;")).mappings().all()
            print_result("Get All Hobbies", [dict(row) for row in result])

            # Get All Cities in Resumes
            result = connection.execute(text("SELECT DISTINCT c.name FROM cities c "
                                             "JOIN users u ON c.id = u.city_id "
                                             "JOIN resumes r ON r.user_id = u.id;")).mappings().all()
            print_result("Get All Cities in Resumes", [dict(row) for row in result])

            # Get Hobbies of Applicants Living in a Specific City
            city_name = 'Kyiv'
            result = connection.execute(text("SELECT DISTINCT h.name FROM hobbies h "
                                             "JOIN resume_hobbies rh ON h.id = rh.hobby_id "
                                             "JOIN resumes r ON r.id = rh.resume_id "
                                             "JOIN users u ON u.id = r.user_id "
                                             "JOIN cities c ON c.id = u.city_id "
                                             "WHERE c.name = :city_name;"), {'city_name': city_name}).mappings().all()
            print_result(f"Get Hobbies in {city_name}", [dict(row) for row in result])

            # Get Applicants who Worked at the Same Institution
            institution_name = 'Company A'
            result = connection.execute(text("SELECT DISTINCT u.id, u.login FROM users u "
                                             "JOIN experience e ON e.user_id = u.id "
                                             "WHERE e.institution = :institution_name;"), {'institution_name': institution_name}).mappings().all()
            print_result(f"Get Applicants from {institution_name}", [dict(row) for row in result])


def mongodb_driver():
    with MongoDbContainer("mongo:8") as mongo:
        client = MongoClient(mongo.get_connection_url())
        db = client.test

        # No migrations required

        # Insert data
        db.users.insert_many([
            {
                "login": "user1",
                "password": "pass",
                "city": "Kyiv",
                "resumes": [{
                    "_id": 1,
                    "summary": "Experienced software developer...",
                    "hobbies": ["football", "reading"]
                }],
                "experience": [{"institution": "Company A"}]
            },
            {
                "login": "user2",
                "password": "pass",
                "city": "Lviv",
                "resumes": [{
                    "_id": 2,
                    "summary": "Data scientist...",
                    "hobbies": ["traveling"]
                }],
                "experience": [{"institution": "Company B"}, {"institution": "Company A"}]
            }
        ])

        # Perform queries

        # Get Resume of a specific user
        result = db.users.find({"login": "user1"}, {"resumes": 1, "_id": 0})
        print_result("Get Resume", list(result))

        # Get All Hobbies in Resumes
        result = db.users.distinct("resumes.hobbies")
        print_result("Get All Hobbies", result)

        # Get All Cities in Resumes
        result = db.users.distinct("city", {"resumes": {"$exists": True}})
        print_result("Get All Cities in Resumes", result)

        # Get Hobbies of Applicants Living in a Specific City
        city_name = "Kyiv"
        result = db.users.distinct("resumes.hobbies", {"city": city_name})
        print_result(f"Get Hobbies in {city_name}", result)

        # Get Applicants who Worked at the Same Institution
        institution_name = "Company A"
        result = db.users.find(
            {"experience.institution": institution_name},
            {"login": 1, "_id": 0}
        )
        print_result(f"Get Applicants from {institution_name}", list(result))


def neo4j_driver():
    with Neo4jContainer() as neo4j, neo4j.get_driver() as driver, driver.session() as session:
        # Run migrations
        session.run("CREATE CONSTRAINT FOR (u:User) REQUIRE u.login IS UNIQUE;")
        session.run("CREATE CONSTRAINT FOR (c:City) REQUIRE c.name IS UNIQUE;")
        session.run("CREATE CONSTRAINT FOR (h:Hobby) REQUIRE h.name IS UNIQUE;")
        session.run("CREATE CONSTRAINT FOR (i:Institution) REQUIRE i.name IS UNIQUE;")

        # Insert data
        session.run("CREATE (c1:City {name: 'Kyiv'});")
        session.run("CREATE (c2:City {name: 'Lviv'});")
        session.run("CREATE (u1:User {login: 'user1', password: 'pass'});")
        session.run("CREATE (u2:User {login: 'user2', password: 'pass'});")
        session.run("CREATE (h1:Hobby {name: 'football'});")
        session.run("CREATE (h2:Hobby {name: 'reading'});")
        session.run("CREATE (h3:Hobby {name: 'traveling'});")
        session.run("CREATE (i1:Institution {name: 'Company A'});")
        session.run("CREATE (i2:Institution {name: 'Company B'});")

        session.run("MATCH (u1:User {login: 'user1'}), (c1:City {name: 'Kyiv'}) CREATE (u1)-[:LIVES_IN]->(c1);")
        session.run("MATCH (u2:User {login: 'user2'}), (c2:City {name: 'Lviv'}) CREATE (u2)-[:LIVES_IN]->(c2);")
        session.run("MATCH (u1:User {login: 'user1'}), (h1:Hobby {name: 'football'}) CREATE (u1)-[:HAS_HOBBY]->(h1);")
        session.run("MATCH (u1:User {login: 'user1'}), (h2:Hobby {name: 'reading'}) CREATE (u1)-[:HAS_HOBBY]->(h2);")
        session.run("MATCH (u2:User {login: 'user2'}), (h3:Hobby {name: 'traveling'}) CREATE (u2)-[:HAS_HOBBY]->(h3);")
        session.run("MATCH (u1:User {login: 'user1'}), (i1:Institution {name: 'Company A'}) CREATE (u1)-[:WORKED_AT]->(i1);")
        session.run("MATCH (u2:User {login: 'user2'}), (i1:Institution {name: 'Company A'}) CREATE (u2)-[:WORKED_AT]->(i1);")
        session.run("MATCH (u2:User {login: 'user2'}), (i2:Institution {name: 'Company B'}) CREATE (u2)-[:WORKED_AT]->(i2);")

        # Perform queries

        # 1. Get Resume (returns user info and hobbies as resume is not a distinct entity in this model)
        result = session.run("""
            MATCH (u:User {login: 'user1'})-[:HAS_HOBBY]->(h:Hobby)
            RETURN u.login AS login, COLLECT(h.name) AS hobbies
        """).data()
        print_result("Get Resume", result)

        # 2. Get All Hobbies
        result = session.run("""
            MATCH (h:Hobby)
            RETURN DISTINCT h.name AS hobby_name
        """).data()
        print_result("Get All Hobbies", result)

        # 3. Get All Cities in Resumes (users with associated cities)
        result = session.run("""
            MATCH (u:User)-[:LIVES_IN]->(c:City)
            RETURN DISTINCT c.name AS city_name
        """).data()
        print_result("Get All Cities in Resumes", result)

        # 4. Get Hobbies of Applicants Living in a Specific City
        city_name = 'Kyiv'
        result = session.run("""
            MATCH (u:User)-[:LIVES_IN]->(c:City {name: $city_name})-[:HAS_HOBBY]->(h:Hobby)
            RETURN DISTINCT h.name AS hobby_name
        """, {'city_name': city_name}).data()
        print_result(f"Get Hobbies in {city_name}", result)

        # 5. Get Applicants who Worked at the Same Institution
        institution_name = 'Company A'
        result = session.run("""
            MATCH (u:User)-[:WORKED_AT]->(i:Institution {name: $institution_name})
            RETURN DISTINCT u.login AS user_login
        """, {'institution_name': institution_name}).data()
        print_result(f"Get Applicants from {institution_name}", result)


if __name__ == "__main__":
    print("PostgreSQL Test:")
    postgres_driver()

    print("\nMongoDB Test:")
    mongodb_driver()

    print("\nNeo4j Test:")
    neo4j_driver()

