build:
	go build -o mysql-monitor.elf

check:
	make build && ./mysql-monitor.elf --host=localhost --database=test --user=test --password=123 --table=table1 --max=10 --table=table2 --max=15
