CC = g++
CCFLAGS = -Wall -O2

all: voraz backtracking costes

voraz: voraz.cpp
	$(CC) $(CCFLAGS) -o voraz voraz.cpp

backtracking: bt.cpp
	$(CC) $(CCFLAGS) -o backtracking bt.cpp

costes: costes.cpp
	$(CC) $(CCFLAGS) -o costes costes.cpp

clean:
	rm -f voraz backtracking costes *.o *.dat *.png
