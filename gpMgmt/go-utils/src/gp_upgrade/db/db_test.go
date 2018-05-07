package db

import (
	"os"

	"gp_upgrade/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("db connector", func() {
	Describe("NewDBConn", func() {
		Context("Database connection receives its constructor parameters", func() {
			It("gets the DBName from dbname argument, Port from masterPort, and Host from masterHost", func() {
				dbConnector := NewDBConn("localHost", 5432, "testdb")
				Expect(dbConnector.DBName).To(Equal("testdb"))
				Expect(dbConnector.Host).To(Equal("localHost"))
				Expect(dbConnector.Port).To(Equal(5432))
			})
		})
		Context("when dbname param is empty, but PGDATABASE env var set", func() {
			It("gets the DBName from PGDATABASE", func() {
				oldPgDatabase := os.Getenv("PGDATABASE")
				os.Setenv("PGDATABASE", "testdb")
				defer os.Setenv("PGDATABASE", oldPgDatabase)

				dbConnector := NewDBConn("localHost", 5432, "testdb")
				Expect(dbConnector.DBName).To(Equal("testdb"))
			})
		})
		Context("when dbname param empty and PGDATABASE env var empty", func() {
			It("has an empty database name", func() {
				oldPgDatabase := os.Getenv("PGDATABASE")
				os.Setenv("PGDATABASE", "")
				defer os.Setenv("PGDATABASE", oldPgDatabase)

				dbConnector := NewDBConn("localHost", 5432, "")
				Expect(dbConnector.DBName).To(Equal(""))
			})
		})
		Context("when Host parameter is empty, but PGHOST is set", func() {
			It("uses PGHOST value", func() {
				old := os.Getenv("PGHOST")
				os.Setenv("PGHOST", "foo")
				defer os.Setenv("PGHOST", old)

				dbConnector := NewDBConn("", 5432, "testdb")
				Expect(dbConnector.Host).To(Equal("foo"))
			})
		})
		Context("when Host parameter is empty and PGHOST is empty", func() {
			It("uses localHost", func() {
				old := os.Getenv("PGHOST")
				os.Setenv("PGHOST", "")
				defer os.Setenv("PGHOST", old)

				dbConnector := NewDBConn("", 5432, "")
				currentHost, _ := utils.GetHost()
				Expect(dbConnector.Host).To(Equal(currentHost))
			})
		})
		Context("when Port parameter is 0 and PGPORT is set", func() {
			It("uses PGPORT", func() {
				old := os.Getenv("PGPORT")
				os.Setenv("PGPORT", "777")
				defer os.Setenv("PGPORT", old)

				dbConnector := NewDBConn("", 0, "")
				Expect(dbConnector.Port).To(Equal(777))
			})
		})
		Context("when Port parameter is 0 and PGPORT is not set", func() {
			It("uses 15432", func() {
				old := os.Getenv("PGPORT")
				os.Setenv("PGPORT", "")
				defer os.Setenv("PGPORT", old)

				dbConnector := NewDBConn("", 0, "")
				Expect(dbConnector.Port).To(Equal(15432))
			})
		})
	})
})
