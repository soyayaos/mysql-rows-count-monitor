## Description

The program checks the number of rows in a table or tables. If the received value exceeds the previous one by the specified value, then a message is written to the log.

## Example record in the log

```
localhost;3306;test;table1;7;10;10;42.857143
```

where

* `localhost` - MySQL host
* `3306` - MySQL port
* `test` - MySQL DB
* `table1` - MySQL table
* `7` - previous `count()` value
* `10` - current `count()` value
* `10` - maximum delta value
* `42.857143` - current delta value

## Build
```
make build
```

## Run

```
./mysql-monitor.elf \
    --host=localhost \
    --database=test \
    --user=test \
    --password=123 \
    --table=table1 \
    --max=10 \
    --table=table2 \
    --max=15
```

* table `table1` maximum quantity change is `10%`
* table `table2` maximum quantity change is `15%`
