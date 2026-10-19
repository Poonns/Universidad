# Uncomment one of these to compile the program with each heuristics
# Non-informative
CHOICE=ni
# Give a name to your heuristic and use that name in heuristic.cpp 
#CHOICE=anyanme
#
# Add here other heuristics you want to try..


CHOSEN_H=${CHOICE}_HEURISTIC

all: takuzu_test

takuzu_test: takuzu_test.cpp takuzu.o heuristic.o 
	g++ -Wall -std=c++20 takuzu.o heuristic.o takuzu_test.cpp -o takuzu_test

takuzu.o: takuzu.cpp takuzu.h
	g++ -Wall -std=c++20 -c takuzu.cpp

heuristic.o: heuristic.cpp heuristic.h takuzu.h Makefile
	g++ -Wall -std=c++20 -D${CHOSEN_H} -c heuristic.cpp

clean:
	-rm -f *.o
	-rm -f takuzu_test

