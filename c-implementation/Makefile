.PHONY: main
main: lib/sqlite3.o
	c++ main.cpp helpers.cpp lib/sqlite3.o -ltins -lcurl -ldl -pthread -o newData -std=c++11 -Wno-psabi -Wall -Wextra -pedantic

lib/sqlite3.o:
	cd lib; gcc -c -I ./lib sqlite3.c shell.c -ldl -pthread

.PHONY: debug
debug:
	c++ main.cpp helpers.cpp -ltins -lcurl -pthread -o debug.bin -std=c++11 -Wno-psabi -Wall -Wextra -pedantic -g

.PHONY: clean
clean:
	rm newData