package cmd

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "os"
    "strconv"

    _ "github.com/go-sql-driver/mysql"
    "github.com/spf13/cobra"
)

var cacheFile = "cache.json"
var cache = make(map[string]int)

var mysqlHost string
var mysqlPort string
var mysqlDatabase string
var mysqlUser string
var mysqlPassword string
var mysqlTables []string
var tablesMaxChanges []int

var rootCmd = &cobra.Command{
    Use:   "mysql-monitor",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {

        if len(mysqlTables) != len(tablesMaxChanges) {
            log.Fatal("Tables count != max count")
        }

        db, err := sql.Open("mysql", mysqlUser+":"+mysqlPassword+"@tcp("+mysqlHost+":"+mysqlPort+")/"+mysqlDatabase)
        defer db.Close()
        if err != nil {
            log.Fatal(err)
        }

        checkCache()
        readCache()

        var count int

        for idx, table := range mysqlTables {
            err := db.QueryRow("select count(*) from " + table).Scan(&count)
            if err != nil {
                log.Fatal(err)
            }
            cacheKey := mysqlHost + "-" + mysqlPort + "-" + mysqlDatabase + "-" + table
            prev, ok := cache[cacheKey]
            if !ok {
                prev = count
            }
            if prev > 0 {
                delta := float64(count)*100/float64(prev) - 100
                if delta > float64(tablesMaxChanges[idx]) {
                    writeToLog(table, prev, count, tablesMaxChanges[idx], delta)
                }
            }
            cache[cacheKey] = count
        }

        saveCache()

    },
}

func Execute() {
    err := rootCmd.Execute()
    if err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVarP(&mysqlHost, "host", "", "", "MySQL host")
    if err := rootCmd.MarkPersistentFlagRequired("host"); err != nil {
        log.Fatal(err)
    }

    rootCmd.PersistentFlags().StringVarP(&mysqlPort, "port", "", "3306", "MySQL port")

    rootCmd.PersistentFlags().StringVarP(&mysqlDatabase, "database", "", "", "MySQL database")
    if err := rootCmd.MarkPersistentFlagRequired("database"); err != nil {
        log.Fatal(err)
    }

    rootCmd.PersistentFlags().StringVarP(&mysqlUser, "user", "", "", "MySQL user")
    if err := rootCmd.MarkPersistentFlagRequired("user"); err != nil {
        log.Fatal(err)
    }

    rootCmd.PersistentFlags().StringVarP(&mysqlPassword, "password", "", "", "MySQL password")
    if err := rootCmd.MarkPersistentFlagRequired("password"); err != nil {
        log.Fatal(err)
    }

    rootCmd.PersistentFlags().StringSliceVarP(&mysqlTables, "table", "", []string{}, "MySQL table")
    if err := rootCmd.MarkPersistentFlagRequired("table"); err != nil {
        log.Fatal(err)
    }

    rootCmd.PersistentFlags().IntSliceVarP(&tablesMaxChanges, "max", "", []int{}, "MySQL max changes per table")
    if err := rootCmd.MarkPersistentFlagRequired("max"); err != nil {
        log.Fatal(err)
    }

}

func checkCache() {
    if _, err := os.Stat(cacheFile); err == nil {
        // ok
    } else if errors.Is(err, os.ErrNotExist) {
        if err := os.WriteFile(cacheFile, []byte("{}"), 0644); err != nil {
            log.Fatal(err)
        }
    } else {
        // oops
        log.Fatal("Something strange...")
    }
}

func readCache() {
    file, err := os.Open(cacheFile)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    bytes, err := io.ReadAll(file)
    if err != nil {
        log.Fatal(err)
    }

    if err := json.Unmarshal(bytes, &cache); err != nil {
        log.Fatal(err)
    }

}

func saveCache() {
    bytes, err := json.Marshal(cache)
    if err != nil {
        log.Fatal(err)
    }
    if err := os.WriteFile(cacheFile, bytes, 0644); err != nil {
        log.Fatal(err)
    }
}

func writeToLog(table string, prev int, current int, max int, delta float64) {
    file, err := os.OpenFile("mysql-monitor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    if _, err := file.WriteString(mysqlHost + ";" + mysqlPort + ";" + mysqlDatabase + ";" + table + ";" + strconv.Itoa(prev) + ";" + strconv.Itoa(current) + ";" + strconv.Itoa(max) + ";" + fmt.Sprintf("%f", delta) + "\n"); err != nil {
        log.Fatal(err)
    }
}
