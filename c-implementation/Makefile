.PHONY: main
main:
	c++ main.cpp helpers.cpp -ltins -lcurl -pthread -o newData -std=c++11 -Wno-psabi -Wall -Wextra -pedantic
debug:
	c++ main.cpp helpers.cpp -ltins -lcurl -pthread -o debug.bin -std=c++11 -Wno-psabi -Wall -Wextra -pedantic -g

clean:
	rm newData